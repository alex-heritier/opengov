import logging
import hashlib
import json
from fastapi import APIRouter, Depends, HTTPException, Query, Response, Request
from sqlalchemy import desc
from sqlalchemy.orm import Session
from app.routers.common import get_db, limiter
from app.models import Article
from app.schemas import ArticleResponse, ArticleDetail, FeedResponse

logger = logging.getLogger(__name__)

router = APIRouter(prefix="/api/feed", tags=["feed"])


@router.get("", response_model=FeedResponse)
@limiter.limit("100/minute")
async def get_feed(
    request: Request,
    response: Response,
    page: int = Query(1, ge=1, description="Page number"),
    limit: int = Query(20, ge=1, le=100, description="Items per page (max 100)"),
    sort: str = Query("newest", regex="^(newest|oldest)$", description="Sort order"),
    db: Session = Depends(get_db),
):
    """Get paginated list of articles with rate limiting (100 req/min)"""
    
    # Build query
    query = db.query(Article)
    
    # Count total
    total = query.count()
    
    # Sort
    if sort == "newest":
        query = query.order_by(desc(Article.published_at))
    else:
        query = query.order_by(Article.published_at)
    
    # Paginate
    skip = (page - 1) * limit
    articles = query.offset(skip).limit(limit).all()
    
    # Add cache headers (5 minute TTL)
    response.headers["Cache-Control"] = "public, max-age=300"

    # Stable ETag: hash of serialized data (use JSON for consistency)
    articles_data = [
        {
            "id": a.id,
            "title": a.title,
            "summary": a.summary,
            "published_at": str(a.published_at)
        }
        for a in articles
    ]
    articles_json = json.dumps(articles_data, sort_keys=True)
    etag_hash = hashlib.md5(articles_json.encode()).hexdigest()
    response.headers["ETag"] = f'"{etag_hash}"'

    return FeedResponse(
        articles=[ArticleResponse.from_orm(a) for a in articles],
        page=page,
        limit=limit,
        total=total,
        has_next=(skip + limit) < total,
    )


@router.get("/{article_id}", response_model=ArticleDetail)
@limiter.limit("100/minute")
async def get_article(request: Request, article_id: int, db: Session = Depends(get_db)):
    """Get specific article details with rate limiting"""
    
    article = db.query(Article).filter(Article.id == article_id).first()
    
    if not article:
        raise HTTPException(status_code=404, detail="Article not found")
    
    return ArticleDetail.from_orm(article)
