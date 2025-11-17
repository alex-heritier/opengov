import logging
from datetime import datetime, timezone
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession
from app.database import WorkerAsyncSessionLocal
from app.models import Article, FederalRegister
from app.services.federal_register import fetch_recent_documents
from app.services.grok import summarize_text
from app.config import settings

logger = logging.getLogger(__name__)

# Batch size for persisting articles to database
BATCH_SIZE = 50


async def _insert_batch(
    db: AsyncSession, new_fed_entries: list, new_articles: list
) -> None:
    """Insert a batch of federal entries and articles into the database."""
    if not new_fed_entries:
        return

    logger.debug(
        f"Inserting batch of {len(new_fed_entries)} fed_entries and "
        f"{len(new_articles)} articles"
    )
    db.add_all(new_fed_entries)
    await db.flush()

    # Link articles to fed entries
    for fed_entry, article in zip(new_fed_entries, new_articles):
        article.federal_register_id = fed_entry.id

    db.add_all(new_articles)
    await db.commit()
    logger.info(
        f"✓ Batch persisted: {len(new_fed_entries)} entries and "
        f"{len(new_articles)} articles"
    )

    # Clear the lists
    new_fed_entries.clear()
    new_articles.clear()


async def fetch_and_process():
    """
    Fetch recent Federal Register documents, process them, and insert into database.

    This function:
    1. Fetches new Federal Register items
    2. Checks for duplicates in database
    3. Summarizes with Grok API
    4. Inserts Article rows
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
        new_fed_entries = []
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

                # Check for duplicates by source_url (unique constraint) or
                # document_number (in FederalRegister)
                existing_article_result = await db.execute(
                    select(Article).where(Article.source_url == source_url)
                )
                existing_article = existing_article_result.scalar_one_or_none()

                existing_fed_entry_result = await db.execute(
                    select(FederalRegister).where(
                        FederalRegister.document_number == doc_number
                    )
                )
                existing_fed_entry = (
                    existing_fed_entry_result.scalar_one_or_none()
                )

                if existing_article or existing_fed_entry:
                    logger.debug("  → Already in database, skipping")
                    skipped_count += 1
                    continue

                # Prepare raw entry
                fed_entry = FederalRegister(
                    document_number=doc_number,
                    raw_data=doc,
                    fetched_at=datetime.now(timezone.utc),
                    processed=True,  # Mark as processed since adding article
                )
                new_fed_entries.append(fed_entry)

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

                # Prepare article (don't set federal_register_id yet)
                article = Article(
                    title=title,
                    summary=summary,
                    source_url=source_url,
                    published_at=published_at,
                )
                new_articles.append(article)

                processed_count += 1
                logger.info(f"  ✓ Article prepared: {doc_number}")

                # Flush batch if threshold reached
                if len(new_fed_entries) >= BATCH_SIZE:
                    logger.info(
                        f"Batch size {len(new_fed_entries)} reached, "
                        "persisting..."
                    )
                    await _insert_batch(db, new_fed_entries, new_articles)

            except Exception as e:
                error_count += 1
                logger.error(
                    f"  ✗ Error processing document "
                    f"{doc.get('document_number', 'UNKNOWN')}: {e}",
                    exc_info=True,
                )
                continue  # Skip failed items, don't rollback

        logger.info(
            f"Processing complete. Final batch: "
            f"{len(new_fed_entries)} fed_entries, "
            f"{len(new_articles)} articles"
        )

        # Insert remaining articles
        if new_fed_entries:
            logger.info(
                f"Persisting final batch of {len(new_fed_entries)} "
                "entries..."
            )
            await _insert_batch(db, new_fed_entries, new_articles)

        logger.info(
            f"Scraper complete. Processed: {processed_count}, "
            f"Skipped: {skipped_count}, Errors: {error_count}"
        )

    except Exception as e:
        logger.error(f"Fatal error in scraper: {e}", exc_info=True)
        await db.rollback()  # Rollback entire batch on fatal error

    finally:
        await db.close()
