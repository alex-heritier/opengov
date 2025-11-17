import logging
from typing import Optional
from fastapi import APIRouter, Depends, HTTPException, Request
from sqlalchemy.orm import Session
from app.routers.common import get_db, limiter
from app.models import User
from app.schemas import LikeToggle, LikeResponse
from app.services import like as like_service
from app.auth import current_active_user

logger = logging.getLogger(__name__)

router = APIRouter(prefix="/api/likes", tags=["likes"])


@router.post("/toggle", response_model=Optional[LikeResponse])
@limiter.limit("100/minute")
async def toggle_like(
    request: Request,
    like_data: LikeToggle,
    current_user: User = Depends(current_active_user),
    db: Session = Depends(get_db),
):
    """
    Toggle like/dislike status for an article.

    Requires authentication. If the user clicks the same vote twice, it will be removed.
    If they click a different vote, it will be updated.
    """
    try:
        like = like_service.toggle_like(
            db=db,
            user_id=current_user.id,
            frarticle_id=like_data.frarticle_id,
            is_positive=like_data.is_positive
        )

        if like is None:
            # Like was removed
            return None

        return LikeResponse.model_validate(like)
    except ValueError as e:
        logger.warning(f"Invalid like toggle request: {e}")
        raise HTTPException(status_code=404, detail=str(e))
    except Exception as e:
        logger.error(f"Error toggling like: {e}", exc_info=True)
        raise HTTPException(status_code=500, detail="Failed to toggle like")


@router.delete("/{frarticle_id}")
@limiter.limit("100/minute")
async def remove_like(
    request: Request,
    frarticle_id: int,
    current_user: User = Depends(current_active_user),
    db: Session = Depends(get_db),
):
    """
    Remove a like/dislike for a specific article.

    Requires authentication.
    """
    try:
        success = like_service.remove_like(
            db=db,
            user_id=current_user.id,
            frarticle_id=frarticle_id
        )

        if success:
            return {"success": True, "message": "Like removed"}
        else:
            raise HTTPException(status_code=404, detail="Like not found")
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Error removing like: {e}", exc_info=True)
        raise HTTPException(status_code=500, detail="Failed to remove like")


@router.get("/status/{frarticle_id}", response_model=dict)
@limiter.limit("100/minute")
async def get_like_status(
    request: Request,
    frarticle_id: int,
    current_user: User = Depends(current_active_user),
    db: Session = Depends(get_db),
):
    """
    Check the like status of the current user for a specific article.

    Requires authentication.
    Returns: {"is_positive": true/false/null} where null means no vote
    """
    try:
        is_positive = like_service.get_like_status(
            db=db,
            user_id=current_user.id,
            frarticle_id=frarticle_id
        )

        return {"is_positive": is_positive}
    except Exception as e:
        logger.error(f"Error checking like status: {e}", exc_info=True)
        raise HTTPException(status_code=500, detail="Failed to check like status")


@router.get("/counts/{frarticle_id}", response_model=dict)
@limiter.limit("100/minute")
async def get_like_counts(
    request: Request,
    frarticle_id: int,
    db: Session = Depends(get_db),
):
    """
    Get like and dislike counts for a specific article.

    Does not require authentication.
    """
    try:
        counts = like_service.get_article_like_counts(
            db=db,
            frarticle_id=frarticle_id
        )

        return counts
    except Exception as e:
        logger.error(f"Error getting like counts: {e}", exc_info=True)
        raise HTTPException(status_code=500, detail="Failed to get like counts")
