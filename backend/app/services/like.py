"""Like service for managing user article likes and dislikes"""
import logging
from datetime import datetime, timezone
from typing import List, Optional
from sqlalchemy.orm import Session
from sqlalchemy import and_
from app.models import Like, FRArticle

logger = logging.getLogger(__name__)


def toggle_like(db: Session, user_id: int, frarticle_id: int, is_positive: bool) -> Like:
    """
    Toggle like/dislike status for an article.

    If like exists with same is_positive value, remove it.
    If like exists with different is_positive value, update it.
    If like doesn't exist, create a new one.

    Args:
        db: Database session
        user_id: ID of the user
        frarticle_id: ID of the article to like/dislike
        is_positive: True for like, False for dislike

    Returns:
        Like object with current status, or None if removed
    """
    # Check if article exists
    article = db.query(FRArticle).filter(FRArticle.id == frarticle_id).first()
    if not article:
        raise ValueError(f"Article with ID {frarticle_id} not found")

    # Find existing like
    like = db.query(Like).filter(
        and_(
            Like.user_id == user_id,
            Like.frarticle_id == frarticle_id
        )
    ).first()

    if like:
        if like.is_positive == is_positive:
            # Same vote - remove it (toggle off)
            db.delete(like)
            db.commit()
            logger.info(f"Removed like {like.id} for user {user_id}")
            return None
        else:
            # Different vote - update it
            like.is_positive = is_positive
            like.updated_at = datetime.now(timezone.utc)
            logger.info(f"Updated like {like.id} for user {user_id}: is_positive={like.is_positive}")
    else:
        # Create new like
        like = Like(
            user_id=user_id,
            frarticle_id=frarticle_id,
            is_positive=is_positive,
            created_at=datetime.now(timezone.utc),
            updated_at=datetime.now(timezone.utc)
        )
        db.add(like)
        logger.info(f"Created new like for user {user_id}, article {frarticle_id}, is_positive={is_positive}")

    db.commit()
    if like:
        db.refresh(like)
    return like


def get_like_status(db: Session, user_id: int, frarticle_id: int) -> Optional[bool]:
    """
    Check the like status of a user for a specific article.

    Args:
        db: Database session
        user_id: ID of the user
        frarticle_id: ID of the article

    Returns:
        True if liked, False if disliked, None if not voted
    """
    like = db.query(Like).filter(
        and_(
            Like.user_id == user_id,
            Like.frarticle_id == frarticle_id
        )
    ).first()

    return like.is_positive if like else None


def remove_like(db: Session, user_id: int, frarticle_id: int) -> bool:
    """
    Remove a like/dislike.

    Args:
        db: Database session
        user_id: ID of the user
        frarticle_id: ID of the article

    Returns:
        True if like was removed, False if it didn't exist
    """
    like = db.query(Like).filter(
        and_(
            Like.user_id == user_id,
            Like.frarticle_id == frarticle_id
        )
    ).first()

    if like:
        db.delete(like)
        db.commit()
        logger.info(f"Removed like {like.id} for user {user_id}")
        return True

    logger.warning(f"No like found for user {user_id}, article {frarticle_id}")
    return False


def get_article_like_counts(db: Session, frarticle_id: int) -> dict:
    """
    Get like and dislike counts for an article.

    Args:
        db: Database session
        frarticle_id: ID of the article

    Returns:
        Dictionary with 'likes' and 'dislikes' counts
    """
    likes_count = db.query(Like).filter(
        and_(
            Like.frarticle_id == frarticle_id,
            Like.is_positive == True
        )
    ).count()

    dislikes_count = db.query(Like).filter(
        and_(
            Like.frarticle_id == frarticle_id,
            Like.is_positive == False
        )
    ).count()

    return {
        "likes": likes_count,
        "dislikes": dislikes_count
    }
