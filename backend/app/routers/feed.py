import logging
import hashlib
import json
from typing import Optional
from fastapi import APIRouter, Depends, HTTPException, Query, Response, Request
from sqlalchemy import desc
from sqlalchemy.orm import Session
from app.routers.common import get_db, limiter
from app.models import FRArticle, User
from app.schemas import ArticleResponse, ArticleDetail, FeedResponse
from app.auth import optional_current_user
from app.services.bookmark import get_bookmark_status

logger = logging.getLogger(__name__)

router = APIRouter(prefix="/api/feed", tags=["feed"])


@router.get("", response_model=FeedResponse)
@limiter.limit("100/minute")
def get_feed(
    request: Request,
    response: Response,
    page: int = Query(1, ge=1, description="Page number"),
    limit: int = Query(20, ge=1, le=100, description="Items per page (max 100)"),
    sort: str = Query("newest", pattern="^(newest|oldest)$", description="Sort order"),
    db: Session = Depends(get_db),
    current_user: Optional[User] = Depends(optional_current_user),
):
    """Get paginated list of articles with rate limiting (100 req/min)"""
    # Prevent DoS from large offsets
    MAX_OFFSET = 10000
    offset = (page - 1) * limit
    if offset > MAX_OFFSET:
        raise HTTPException(status_code=400, detail="Page number too high")

    # Build query - no joins needed, document_number is a direct field
    query = db.query(FRArticle)

    # Count total
    total = query.count()

    # Sort
    if sort == "newest":
        query = query.order_by(desc(FRArticle.published_at))
    else:
        query = query.order_by(FRArticle.published_at)

    # Paginate
    articles = query.offset(offset).limit(limit).all()

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
    etag_hash = hashlib.sha256(articles_json.encode()).hexdigest()
    response.headers["ETag"] = f'"{etag_hash}"'

    # Build article responses with bookmark status
    article_responses = []
    for article in articles:
        article_dict = ArticleResponse.model_validate(article).model_dump()
        # Add bookmark status if user is authenticated
        if current_user:
            article_dict["is_bookmarked"] = get_bookmark_status(db, current_user.id, article.id)
        else:
            article_dict["is_bookmarked"] = False
        article_responses.append(ArticleResponse(**article_dict))

    return FeedResponse(
        articles=article_responses,
        page=page,
        limit=limit,
        total=total,
        has_next=(offset + limit) < total,
    )


@router.get("/document/{document_number}", response_model=ArticleDetail)
@limiter.limit("100/minute")
def get_article_by_document_number(
    request: Request,
    document_number: str,
    db: Session = Depends(get_db),
    current_user: Optional[User] = Depends(optional_current_user),
):
    """Get article by Federal Register document number with rate limiting"""

    # Query article directly by document_number (no join needed)
    article = (
        db.query(FRArticle)
        .filter(FRArticle.document_number == document_number)
        .first()
    )

    if not article:
        raise HTTPException(
            status_code=404,
            detail=f"Article with document number '{document_number}' not found"
        )

    # Add bookmark status
    article_dict = ArticleDetail.model_validate(article).model_dump()
    if current_user:
        article_dict["is_bookmarked"] = get_bookmark_status(db, current_user.id, article.id)
    else:
        article_dict["is_bookmarked"] = False

    return ArticleDetail(**article_dict)


@router.get("/{article_id}", response_model=ArticleDetail)
@limiter.limit("100/minute")
def get_article(
    request: Request,
    article_id: int,
    db: Session = Depends(get_db),
    current_user: Optional[User] = Depends(optional_current_user),
):
    """Get specific article details with rate limiting"""

    article = (
        db.query(FRArticle)
        .filter(FRArticle.id == article_id)
        .first()
    )

    if not article:
        raise HTTPException(status_code=404, detail="Article not found")

    # Add bookmark status
    article_dict = ArticleDetail.model_validate(article).model_dump()
    if current_user:
        article_dict["is_bookmarked"] = get_bookmark_status(db, current_user.id, article.id)
    else:
        article_dict["is_bookmarked"] = False

    return ArticleDetail(**article_dict)
