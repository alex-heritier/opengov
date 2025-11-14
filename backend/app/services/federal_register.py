import logging
import asyncio
from datetime import datetime, timedelta, timezone
import httpx
from tenacity import retry, stop_after_attempt, wait_exponential, retry_if_exception_type
from sqlalchemy.orm import Session
from app.config import settings
from app.models import Agency

logger = logging.getLogger(__name__)


@retry(
    stop=stop_after_attempt(3),
    wait=wait_exponential(multiplier=1, min=2, max=10),
    retry=retry_if_exception_type((httpx.TimeoutException, httpx.HTTPStatusError)),
    reraise=True
)
async def fetch_recent_documents(days: int = 1) -> list:
    """
    Fetch recent Federal Register documents.

    Args:
        days: Number of days to look back

    Returns:
        List of document dictionaries
    """
    # Calculate date range
    end_date = datetime.now(timezone.utc).date()
    start_date = end_date - timedelta(days=days)

    params = {
        "filter[publication_date][gte]": start_date.isoformat(),
        "filter[publication_date][lte]": end_date.isoformat(),
        "per_page": settings.FEDERAL_REGISTER_PER_PAGE,
        "page": 1,
    }

    documents = []
    page_count = 0

    try:
        logger.info(
            f"Starting Federal Register API fetch for {days} day(s) "
            f"({start_date} to {end_date})"
        )
        async with httpx.AsyncClient(timeout=settings.FEDERAL_REGISTER_TIMEOUT) as client:
            while page_count < settings.FEDERAL_REGISTER_MAX_PAGES:
                api_url = f"{settings.FEDERAL_REGISTER_API_URL}/documents"
                logger.debug(f"Fetching page {params['page']} from {api_url}")
                logger.debug(f"Query params: {params}")

                response = await client.get(api_url, params=params)

                logger.info(
                    f"Federal Register API response: "
                    f"status={response.status_code}, page={params['page']}"
                )
                response.raise_for_status()

                data = response.json()
                results = data.get("results", [])
                total_results = data.get("total_documents", 0)

                logger.info(
                    f"Page {params['page']}: Got {len(results)} results "
                    f"(total in API: {total_results})"
                )
                documents.extend(results)

                # Check for pagination
                if len(results) < params["per_page"]:
                    logger.info(
                        f"Reached last page (got {len(results)} results "
                        f"< {params['per_page']} per_page)"
                    )
                    break

                params["page"] += 1
                page_count += 1
                # Be respectful to the API: 0.5 second delay between paginated requests
                await asyncio.sleep(0.5)

        logger.info(
            f"Successfully fetched {len(documents)} documents "
            f"from Federal Register API"
        )
        return documents

    except httpx.TimeoutException:
        logger.error(f"Federal Register API timeout after {settings.FEDERAL_REGISTER_TIMEOUT}s")
        return []
    except httpx.HTTPError as e:
        status = e.response.status_code if hasattr(e, 'response') else 'unknown'
        logger.error(f"Federal Register API HTTP error: {status} - {e}")
        return []
    except Exception as e:
        logger.error(f"Unexpected error fetching Federal Register: {e}", exc_info=True)
        return []


@retry(
    stop=stop_after_attempt(3),
    wait=wait_exponential(multiplier=1, min=2, max=10),
    retry=retry_if_exception_type((httpx.TimeoutException, httpx.HTTPStatusError)),
    reraise=True
)
async def fetch_agencies() -> list:
    """
    Fetch all agencies from Federal Register API.

    Returns:
        List of agency dictionaries
    """
    try:
        logger.info("Starting Federal Register agencies fetch")
        async with httpx.AsyncClient(timeout=settings.FEDERAL_REGISTER_TIMEOUT) as client:
            api_url = f"{settings.FEDERAL_REGISTER_API_URL}/agencies"
            logger.debug(f"Fetching agencies from {api_url}")

            response = await client.get(api_url)
            logger.info(
                f"Federal Register agencies API response: "
                f"status={response.status_code}"
            )
            response.raise_for_status()

            agencies = response.json()

            logger.info(
                f"Successfully fetched {len(agencies)} agencies "
                f"from Federal Register API"
            )
            return agencies

    except httpx.TimeoutException:
        logger.error(
            f"Federal Register agencies API timeout after "
            f"{settings.FEDERAL_REGISTER_TIMEOUT}s"
        )
        return []
    except httpx.HTTPError as e:
        status = e.response.status_code if hasattr(e, 'response') else 'unknown'
        logger.error(f"Federal Register agencies API HTTP error: {status} - {e}")
        return []
    except Exception as e:
        logger.error(f"Unexpected error fetching agencies: {e}", exc_info=True)
        return []


def store_agencies(db: Session, agencies_data: list) -> dict:
    """
    Store agencies in database, avoiding duplicates.

    Args:
        db: Database session
        agencies_data: List of agency dictionaries from API

    Returns:
        Dictionary with counts of created, updated, and skipped agencies
    """
    created_count = 0
    updated_count = 0
    skipped_count = 0
    error_count = 0

    for agency_data in agencies_data:
        try:
            # Extract agency ID from the data
            fr_agency_id = agency_data.get("id")
            if not fr_agency_id:
                logger.warning(
                    f"Agency data missing 'id' field, skipping: "
                    f"{agency_data.get('name', 'Unknown')}"
                )
                error_count += 1
                continue

            # Check if agency already exists
            existing_agency = db.query(Agency).filter(Agency.fr_agency_id == fr_agency_id).first()

            if existing_agency:
                # Update existing agency if data has changed
                updated = False
                if existing_agency.name != agency_data.get("name"):
                    existing_agency.name = agency_data.get("name")
                    updated = True
                if existing_agency.short_name != agency_data.get("short_name"):
                    existing_agency.short_name = agency_data.get("short_name")
                    updated = True
                if existing_agency.slug != agency_data.get("slug"):
                    existing_agency.slug = agency_data.get("slug")
                    updated = True
                if existing_agency.description != agency_data.get("description"):
                    existing_agency.description = agency_data.get("description")
                    updated = True
                if existing_agency.url != agency_data.get("url"):
                    existing_agency.url = agency_data.get("url")
                    updated = True
                if existing_agency.json_url != agency_data.get("json_url"):
                    existing_agency.json_url = agency_data.get("json_url")
                    updated = True
                if existing_agency.parent_id != agency_data.get("parent_id"):
                    existing_agency.parent_id = agency_data.get("parent_id")
                    updated = True

                # Always update raw_data and updated_at
                existing_agency.raw_data = agency_data
                existing_agency.updated_at = datetime.now(timezone.utc)

                if updated:
                    updated_count += 1
                    logger.debug(f"Updated agency: {agency_data.get('name')}")
                else:
                    skipped_count += 1
                    logger.debug(
                        f"Skipped agency (no changes): {agency_data.get('name')}"
                    )
            else:
                # Create new agency
                new_agency = Agency(
                    fr_agency_id=fr_agency_id,
                    name=agency_data.get("name", ""),
                    short_name=agency_data.get("short_name"),
                    slug=agency_data.get("slug", ""),
                    description=agency_data.get("description"),
                    url=agency_data.get("url"),
                    json_url=agency_data.get("json_url"),
                    parent_id=agency_data.get("parent_id"),
                    raw_data=agency_data
                )
                db.add(new_agency)
                created_count += 1
                logger.debug(f"Created agency: {agency_data.get('name')}")

        except Exception as e:
            logger.error(
                f"Error processing agency {agency_data.get('name', 'Unknown')}: "
                f"{e}",
                exc_info=True
            )
            error_count += 1
            continue

    # Commit all changes
    try:
        db.commit()
        logger.info(
            f"Stored agencies: {created_count} created, {updated_count} updated, "
            f"{skipped_count} skipped, {error_count} errors"
        )
    except Exception as e:
        db.rollback()
        logger.error(f"Error committing agencies to database: {e}", exc_info=True)
        raise

    return {
        "created": created_count,
        "updated": updated_count,
        "skipped": skipped_count,
        "errors": error_count,
        "total": len(agencies_data)
    }
