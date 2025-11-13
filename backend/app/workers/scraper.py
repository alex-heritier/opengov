import logging
from datetime import datetime
from sqlalchemy.orm import Session
from app.database import SessionLocal
from app.models import Article, FederalRegister, ScraperRun
from app.services.federal_register import fetch_recent_documents
from app.services.grok import summarize_text
from app.config import settings

logger = logging.getLogger(__name__)


async def fetch_and_process():
    """
    Fetch recent Federal Register documents, process them, and insert into database.

    This function:
    1. Checks for running scrapers (simple lock via DB)
    2. Fetches new Federal Register items
    3. Checks for duplicates in database
    4. Summarizes with Grok API
    5. Inserts Article rows
    6. Tracks scraper execution in database
    """
    db: Session = SessionLocal()

    # Check for running scraper (simple lock via DB)
    running_run = db.query(ScraperRun).filter(ScraperRun.completed_at == None).first()
    if running_run:
        logger.warning(f"Scraper already running (ID: {running_run.id}), skipping")
        db.close()
        return

    # Create scraper run record
    run = ScraperRun(started_at=datetime.utcnow())
    db.add(run)
    db.commit()
    
    try:
        logger.info(f"Starting scraper run {run.id} at {run.started_at}")
        logger.info(f"Looking back {settings.SCRAPER_DAYS_LOOKBACK} day(s)")
        
        # Fetch documents from Federal Register API
        logger.info("Fetching documents from Federal Register API...")
        documents = await fetch_recent_documents(days=settings.SCRAPER_DAYS_LOOKBACK)
        
        if not documents:
            logger.warning("No documents fetched from Federal Register API")
            run.completed_at = datetime.utcnow()
            run.success = True
            db.commit()
            return
        
        logger.info(f"Starting to process {len(documents)} documents")
        
        processed_count = 0
        skipped_count = 0
        error_count = 0
        
        for i, doc in enumerate(documents, 1):
            try:
                doc_number = doc.get("document_number", "UNKNOWN")
                title = doc.get("title", "Untitled")
                logger.debug(f"[{i}/{len(documents)}] Processing: {doc_number} - {title[:60]}")
                
                # Check if already in database
                if db.query(FederalRegister).filter(
                    FederalRegister.document_number == doc_number
                ).first():
                    logger.debug(f"  → Already in database, skipping")
                    skipped_count += 1
                    continue
                
                # Store raw entry
                fed_entry = FederalRegister(
                    document_number=doc_number,
                    raw_data=doc,
                    fetched_at=datetime.utcnow(),
                    processed=False,
                )
                db.add(fed_entry)
                db.flush()
                logger.debug(f"  → Stored raw entry in database")
                
                # Extract fields
                abstract = doc.get("abstract", "")
                summary_text = abstract or doc.get("full_text", "")[:1000]
                source_url = doc.get("html_url", "")
                
                # Summarize with Grok
                logger.debug(f"  → Summarizing with Grok API...")
                summary = await summarize_text(summary_text)
                logger.debug(f"  → Summary generated ({len(summary)} chars)")
                
                # Parse published date
                published_at_str = doc.get("publication_date", "")
                try:
                    published_at = datetime.fromisoformat(published_at_str)
                except (ValueError, TypeError):
                    published_at = datetime.utcnow()
                
                # Create article
                article = Article(
                    federal_register_id=fed_entry.id,
                    title=title,
                    summary=summary,
                    source_url=source_url,
                    published_at=published_at,
                )
                
                db.add(article)
                fed_entry.processed = True
                
                db.commit()
                processed_count += 1
                logger.info(f"  ✓ Article created: {doc_number}")
                
            except Exception as e:
                error_count += 1
                logger.error(f"  ✗ Error processing document {doc.get('document_number', 'UNKNOWN')}: {e}", exc_info=True)
                db.rollback()
                continue
        
        logger.info(
            f"Scraper run {run.id} complete. Processed: {processed_count}, "
            f"Skipped: {skipped_count}, Errors: {error_count}"
        )
        
        # Update run record
        run.completed_at = datetime.utcnow()
        run.processed_count = processed_count
        run.skipped_count = skipped_count
        run.error_count = error_count
        run.success = True
        db.commit()
        
    except Exception as e:
        logger.error(f"Fatal error in scraper: {e}")
        run.completed_at = datetime.utcnow()
        run.error_message = str(e)
        run.success = False
        db.commit()
    
    finally:
        db.close()
