import logging
from typing import List
from fastapi import APIRouter, Depends, HTTPException, Request
from sqlalchemy.orm import Session
from app.routers.common import get_db, limiter
from app.models import User, Bookmark, FRArticle
from app.schemas import BookmarkToggle, BookmarkResponse, BookmarkedArticleResponse
from app.services import bookmark as bookmark_service
from app.auth import current_active_user

logger = logging.getLogger(__name__)

router = APIRouter(prefix="/api/bookmarks", tags=["bookmarks"])


@router.post("/toggle", response_model=BookmarkResponse)
@limiter.limit("100/minute")
async def toggle_bookmark(
    request: Request,
    bookmark_data: BookmarkToggle,
    current_user: User = Depends(current_active_user),
    db: Session = Depends(get_db),
):
    """
    Toggle bookmark status for an article.

    Requires authentication. If the article is not bookmarked, it will be bookmarked.
    If it's already bookmarked, it will be unbookmarked.
    """
    try:
        bookmark = bookmark_service.toggle_bookmark(
            db=db,
            user_id=current_user.id,
            frarticle_id=bookmark_data.frarticle_id
        )
        return BookmarkResponse.model_validate(bookmark)
    except ValueError as e:
        logger.warning(f"Invalid bookmark toggle request: {e}")
        raise HTTPException(status_code=404, detail=str(e))
    except Exception as e:
        logger.error(f"Error toggling bookmark: {e}", exc_info=True)
        raise HTTPException(status_code=500, detail="Failed to toggle bookmark")


@router.get("", response_model=List[BookmarkedArticleResponse])
@limiter.limit("100/minute")
async def get_bookmarks(
    request: Request,
    current_user: User = Depends(current_active_user),
    db: Session = Depends(get_db),
):
    """
    Get all bookmarked articles for the current user.

    Requires authentication. Returns articles with bookmark metadata.
    """
    try:
        # Get bookmarked articles with bookmark timestamps
        bookmarks = bookmark_service.get_user_bookmarks(db=db, user_id=current_user.id)

        # Build response with article details and bookmark timestamp
        result = []
        for bookmark in bookmarks:
            article = db.query(FRArticle).filter(FRArticle.id == bookmark.frarticle_id).first()
            if article:
                result.append(BookmarkedArticleResponse(
                    id=article.id,
                    document_number=article.document_number,
                    title=article.title,
                    summary=article.summary,
                    source_url=article.source_url,
                    published_at=article.published_at,
                    created_at=article.created_at,
                    bookmarked_at=bookmark.created_at
                ))

        return result
    except Exception as e:
        logger.error(f"Error fetching bookmarks: {e}", exc_info=True)
        raise HTTPException(status_code=500, detail="Failed to fetch bookmarks")


@router.delete("/{frarticle_id}")
@limiter.limit("100/minute")
async def remove_bookmark(
    request: Request,
    frarticle_id: int,
    current_user: User = Depends(current_active_user),
    db: Session = Depends(get_db),
):
    """
    Remove a bookmark for a specific article.

    Requires authentication.
    """
    try:
        success = bookmark_service.remove_bookmark(
            db=db,
            user_id=current_user.id,
            frarticle_id=frarticle_id
        )

        if success:
            return {"success": True, "message": "Bookmark removed"}
        else:
            raise HTTPException(status_code=404, detail="Bookmark not found")
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Error removing bookmark: {e}", exc_info=True)
        raise HTTPException(status_code=500, detail="Failed to remove bookmark")


@router.get("/status/{frarticle_id}", response_model=dict)
@limiter.limit("100/minute")
async def get_bookmark_status(
    request: Request,
    frarticle_id: int,
    current_user: User = Depends(current_active_user),
    db: Session = Depends(get_db),
):
    """
    Check if the current user has bookmarked a specific article.

    Requires authentication.
    """
    try:
        is_bookmarked = bookmark_service.get_bookmark_status(
            db=db,
            user_id=current_user.id,
            frarticle_id=frarticle_id
        )

        return {"is_bookmarked": is_bookmarked}
    except Exception as e:
        logger.error(f"Error checking bookmark status: {e}", exc_info=True)
        raise HTTPException(status_code=500, detail="Failed to check bookmark status")
