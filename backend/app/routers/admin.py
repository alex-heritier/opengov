import logging
from datetime import datetime, timezone
from fastapi import APIRouter, Depends, Query, Request, BackgroundTasks, HTTPException
from sqlalchemy.orm import Session
from app.routers.common import get_db, limiter
from app.models import Article, FederalRegister, Agency
from app.workers.scraper import fetch_and_process
from app.services.federal_register import fetch_agencies, store_agencies

logger = logging.getLogger(__name__)

router = APIRouter(prefix="/api/admin", tags=["admin"])


@router.post("/scrape")
@limiter.limit("10/minute")
async def manual_scrape(request: Request, background_tasks: BackgroundTasks):
    """Manually trigger Federal Register scrape (10 req/min limit)"""
    logger.info("Manual scrape triggered")
    # FastAPI's BackgroundTasks supports async functions directly
    background_tasks.add_task(fetch_and_process)
    return {"status": "queued", "message": "Scrape job queued in background"}


@router.get("/stats")
@limiter.limit("50/minute")
async def get_stats(request: Request, db: Session = Depends(get_db)):
    """Get article statistics and scraper status (50 req/min limit)"""
    total_articles = db.query(Article).count()

    # Get last scrape time
    last_entry = db.query(FederalRegister).order_by(
        FederalRegister.fetched_at.desc()
    ).first()
    last_scrape_time = last_entry.fetched_at if last_entry else None

    return {
        "total_articles": total_articles,
        "last_scrape_time": last_scrape_time,
        "last_scrape_human": (
            f"{int((datetime.now(timezone.utc) - last_scrape_time).total_seconds())} seconds ago"
            if last_scrape_time else "Never"
        ),
    }


@router.post("/sync-agencies")
@limiter.limit("5/minute")
async def sync_agencies(request: Request, db: Session = Depends(get_db)):
    """Fetch agencies from Federal Register API and store in database (5 req/min limit)"""
    logger.info("Manual agency sync triggered")

    try:
        # Fetch agencies from API
        agencies_data = await fetch_agencies()

        if not agencies_data:
            logger.warning("No agencies returned from API")
            return {
                "status": "error",
                "message": "No agencies returned from Federal Register API"
            }

        # Store agencies in database
        result = store_agencies(db, agencies_data)

        logger.info(f"Agency sync completed: {result}")

        return {
            "status": "success",
            "message": "Agencies synced successfully",
            "data": result
        }

    except Exception as e:
        logger.error(f"Error syncing agencies: {e}", exc_info=True)
        raise HTTPException(status_code=500, detail=f"Error syncing agencies: {str(e)}")


@router.get("/agencies")
@limiter.limit("50/minute")
async def get_agencies(
    request: Request,
    limit: int = Query(100, ge=1, le=500, description="Number of agencies to return"),
    offset: int = Query(0, ge=0, description="Offset for pagination"),
    db: Session = Depends(get_db)
):
    """Get agencies from database (50 req/min limit)"""
    agencies = db.query(Agency).order_by(Agency.name).offset(offset).limit(limit).all()
    total = db.query(Agency).count()

    return {
        "agencies": [
            {
                "id": agency.id,
                "fr_agency_id": agency.fr_agency_id,
                "name": agency.name,
                "short_name": agency.short_name,
                "slug": agency.slug,
                "description": agency.description,
                "url": agency.url,
                "parent_id": agency.parent_id,
                "created_at": agency.created_at,
                "updated_at": agency.updated_at
            }
            for agency in agencies
        ],
        "total": total,
        "limit": limit,
        "offset": offset
    }
