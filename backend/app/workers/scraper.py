import logging
from datetime import datetime, timezone
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession
from app.database import WorkerAsyncSessionLocal
from app.models import FRArticle
from app.services.federal_register import fetch_recent_documents
from app.services.grok import summarize_text
from app.config import settings

logger = logging.getLogger(__name__)

# Batch size for persisting articles to database
BATCH_SIZE = 50


async def _insert_batch(db: AsyncSession, new_articles: list) -> None:
    """Insert a batch of FRArticle objects into the database."""
    if not new_articles:
        return

    logger.debug(f"Inserting batch of {len(new_articles)} articles")
    db.add_all(new_articles)
    await db.commit()
    logger.info(f"✓ Batch persisted: {len(new_articles)} articles")

    # Clear the list
    new_articles.clear()


async def fetch_and_process():
    """
    Fetch recent Federal Register documents, process them, and insert into database.

    This function:
    1. Fetches new Federal Register items
    2. Checks for duplicates in database
    3. Summarizes with Grok API
    4. Inserts FRArticle rows
    """
    db: AsyncSession = WorkerAsyncSessionLocal()

    try:
        logger.info(
            f"Starting scraper (looking back "
            f"{settings.SCRAPER_DAYS_LOOKBACK} day(s))"
        )

        # Fetch documents from Federal Register API
        logger.info("Fetching documents from Federal Register API...")
        documents = await fetch_recent_documents(
            days=settings.SCRAPER_DAYS_LOOKBACK
        )

        if not documents:
            logger.warning("No documents fetched from Federal Register API")
            return

        logger.info(f"Starting to process {len(documents)} documents")

        processed_count = 0
        skipped_count = 0
        error_count = 0

        # Collect objects for bulk insert
        new_articles = []

        for i, doc in enumerate(documents, 1):
            try:
                doc_number = doc.get("document_number", "UNKNOWN")
                title = doc.get("title", "Untitled")
                logger.debug(
                    f"[{i}/{len(documents)}] Processing: "
                    f"{doc_number} - {title[:60]}"
                )

                # Prepare source URL for duplicate check
                source_url = doc.get("html_url", "")

                # Check for duplicates by document_number or source_url (both unique)
                existing_article_result = await db.execute(
                    select(FRArticle).where(
                        (FRArticle.document_number == doc_number) |
                        (FRArticle.source_url == source_url)
                    )
                )
                existing_article = existing_article_result.scalar_one_or_none()

                if existing_article:
                    logger.debug("  → Already in database, skipping")
                    skipped_count += 1
                    continue

                # Extract fields
                abstract = doc.get("abstract", "")
                summary_text = abstract or doc.get("full_text", "")[:1000]

                # Summarize with Grok
                logger.debug("  → Summarizing with Grok API...")
                summary = await summarize_text(summary_text)
                logger.debug(f"  → Summary generated ({len(summary)} chars)")

                # Parse published date
                published_at_str = doc.get("publication_date", "")
                try:
                    published_at = datetime.fromisoformat(published_at_str)
                except (ValueError, TypeError):
                    published_at = datetime.now(timezone.utc)

                # Create unified FRArticle with both raw and processed data
                article = FRArticle(
                    # Federal Register raw data
                    document_number=doc_number,
                    raw_data=doc,
                    fetched_at=datetime.now(timezone.utc),
                    # Processed article data
                    title=title,
                    summary=summary,
                    source_url=source_url,
                    published_at=published_at,
                )
                new_articles.append(article)

                processed_count += 1
                logger.info(f"  ✓ Article prepared: {doc_number}")

                # Flush batch if threshold reached
                if len(new_articles) >= BATCH_SIZE:
                    logger.info(f"Batch size {len(new_articles)} reached, persisting...")
                    await _insert_batch(db, new_articles)

            except Exception as e:
                error_count += 1
                logger.error(
                    f"  ✗ Error processing document "
                    f"{doc.get('document_number', 'UNKNOWN')}: {e}",
                    exc_info=True,
                )
                continue  # Skip failed items, don't rollback

        logger.info(f"Processing complete. Final batch: {len(new_articles)} articles")

        # Insert remaining articles
        if new_articles:
            logger.info(f"Persisting final batch of {len(new_articles)} articles...")
            await _insert_batch(db, new_articles)

        logger.info(
            f"Scraper complete. Processed: {processed_count}, "
            f"Skipped: {skipped_count}, Errors: {error_count}"
        )

    except Exception as e:
        logger.error(f"Fatal error in scraper: {e}", exc_info=True)
        await db.rollback()  # Rollback entire batch on fatal error

    finally:
        await db.close()
