import logging
import asyncio
from datetime import datetime, timedelta, timezone
import httpx
from tenacity import retry, stop_after_attempt, wait_exponential, retry_if_exception_type
from app.config import settings

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
        logger.info(f"Starting Federal Register API fetch for {days} day(s) ({start_date} to {end_date})")
        async with httpx.AsyncClient(timeout=settings.FEDERAL_REGISTER_TIMEOUT) as client:
            while page_count < settings.FEDERAL_REGISTER_MAX_PAGES:
                api_url = f"{settings.FEDERAL_REGISTER_API_URL}/documents"
                logger.debug(f"Fetching page {params['page']} from {api_url}")
                logger.debug(f"Query params: {params}")
                
                response = await client.get(api_url, params=params)
                
                logger.info(f"Federal Register API response: status={response.status_code}, page={params['page']}")
                response.raise_for_status()

                data = response.json()
                results = data.get("results", [])
                total_results = data.get("total_documents", 0)
                
                logger.info(f"Page {params['page']}: Got {len(results)} results (total in API: {total_results})")
                documents.extend(results)

                # Check for pagination
                if len(results) < params["per_page"]:
                    logger.info(f"Reached last page (got {len(results)} results < {params['per_page']} per_page)")
                    break

                params["page"] += 1
                page_count += 1
                # Be respectful to the API: 0.5 second delay between paginated requests
                await asyncio.sleep(0.5)
        
        logger.info(f"Successfully fetched {len(documents)} documents from Federal Register API")
        return documents
        
    except httpx.TimeoutException:
        logger.error(f"Federal Register API timeout after {settings.FEDERAL_REGISTER_TIMEOUT}s")
        return []
    except httpx.HTTPError as e:
        logger.error(f"Federal Register API HTTP error: {e.response.status_code if hasattr(e, 'response') else 'unknown'} - {e}")
        return []
    except Exception as e:
        logger.error(f"Unexpected error fetching Federal Register: {e}", exc_info=True)
        return []
