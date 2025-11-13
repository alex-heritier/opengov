import logging
from datetime import datetime, timezone
from fastapi import APIRouter, Depends, HTTPException, Query, Request, BackgroundTasks
from sqlalchemy import desc
from sqlalchemy.orm import Session
from app.routers.common import get_db, limiter
from app.models import Article, FederalRegister, ScraperRun
from app.schemas import ScraperRunListResponse
from app.workers.scraper import fetch_and_process

logger = logging.getLogger(__name__)

router = APIRouter(prefix="/api/admin", tags=["admin"])


@router.post("/scrape")
@limiter.limit("10/minute")
async def manual_scrape(request: Request, background_tasks: BackgroundTasks):
    """Manually trigger Federal Register scrape (10 req/min limit)"""
    logger.info("Manual scrape triggered")
    # Add async coroutine to background tasks - FastAPI handles it properly
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


@router.get("/scraper-runs", response_model=ScraperRunListResponse)
@limiter.limit("50/minute")
async def get_scraper_runs(
    request: Request,
    limit: int = Query(10, ge=1, le=50, description="Number of runs to return"),
    db: Session = Depends(get_db),
):
    """Get recent scraper runs (50 req/min limit)"""
    runs = db.query(ScraperRun).order_by(
        desc(ScraperRun.started_at)
    ).limit(limit).all()
    
    total = db.query(ScraperRun).count()
    
    return ScraperRunListResponse(runs=runs, total=total)
