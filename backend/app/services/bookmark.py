"""Bookmark service for managing user article bookmarks"""
import logging
from datetime import datetime, timezone
from typing import List, Optional
from sqlalchemy.orm import Session
from sqlalchemy import and_
from app.models import Bookmark, FRArticle

logger = logging.getLogger(__name__)


def toggle_bookmark(db: Session, user_id: int, frarticle_id: int) -> Bookmark:
    """
    Toggle bookmark status for an article.

    If bookmark exists, toggle is_bookmarked field.
    If bookmark doesn't exist, create a new one with is_bookmarked=True.

    Args:
        db: Database session
        user_id: ID of the user
        frarticle_id: ID of the article to bookmark

    Returns:
        Bookmark object with current status
    """
    # Check if article exists
    article = db.query(FRArticle).filter(FRArticle.id == frarticle_id).first()
    if not article:
        raise ValueError(f"Article with ID {frarticle_id} not found")

    # Find existing bookmark
    bookmark = db.query(Bookmark).filter(
        and_(
            Bookmark.user_id == user_id,
            Bookmark.frarticle_id == frarticle_id
        )
    ).first()

    if bookmark:
        # Toggle existing bookmark
        bookmark.is_bookmarked = not bookmark.is_bookmarked
        bookmark.updated_at = datetime.now(timezone.utc)
        logger.info(f"Toggled bookmark {bookmark.id} for user {user_id}: is_bookmarked={bookmark.is_bookmarked}")
    else:
        # Create new bookmark
        bookmark = Bookmark(
            user_id=user_id,
            frarticle_id=frarticle_id,
            is_bookmarked=True,
            created_at=datetime.now(timezone.utc),
            updated_at=datetime.now(timezone.utc)
        )
        db.add(bookmark)
        logger.info(f"Created new bookmark for user {user_id}, article {frarticle_id}")

    db.commit()
    db.refresh(bookmark)
    return bookmark


def get_user_bookmarks(db: Session, user_id: int, include_unbookmarked: bool = False) -> List[Bookmark]:
    """
    Get all bookmarks for a user.

    Args:
        db: Database session
        user_id: ID of the user
        include_unbookmarked: If True, include bookmarks with is_bookmarked=False

    Returns:
        List of Bookmark objects
    """
    query = db.query(Bookmark).filter(Bookmark.user_id == user_id)

    if not include_unbookmarked:
        query = query.filter(Bookmark.is_bookmarked == True)

    bookmarks = query.order_by(Bookmark.created_at.desc()).all()
    logger.info(f"Found {len(bookmarks)} bookmarks for user {user_id}")
    return bookmarks


def get_bookmark_status(db: Session, user_id: int, frarticle_id: int) -> bool:
    """
    Check if a user has bookmarked a specific article.

    Args:
        db: Database session
        user_id: ID of the user
        frarticle_id: ID of the article

    Returns:
        True if bookmarked, False otherwise
    """
    bookmark = db.query(Bookmark).filter(
        and_(
            Bookmark.user_id == user_id,
            Bookmark.frarticle_id == frarticle_id,
            Bookmark.is_bookmarked == True
        )
    ).first()

    return bookmark is not None


def get_bookmarked_articles(db: Session, user_id: int) -> List[FRArticle]:
    """
    Get all articles bookmarked by a user.

    Args:
        db: Database session
        user_id: ID of the user

    Returns:
        List of FRArticle objects that are bookmarked
    """
    # Join bookmarks with articles
    articles = (
        db.query(FRArticle)
        .join(Bookmark, Bookmark.frarticle_id == FRArticle.id)
        .filter(
            and_(
                Bookmark.user_id == user_id,
                Bookmark.is_bookmarked == True
            )
        )
        .order_by(Bookmark.created_at.desc())
        .all()
    )

    logger.info(f"Found {len(articles)} bookmarked articles for user {user_id}")
    return articles


def remove_bookmark(db: Session, user_id: int, frarticle_id: int) -> bool:
    """
    Remove a bookmark (set is_bookmarked to False).

    Args:
        db: Database session
        user_id: ID of the user
        frarticle_id: ID of the article

    Returns:
        True if bookmark was removed, False if it didn't exist
    """
    bookmark = db.query(Bookmark).filter(
        and_(
            Bookmark.user_id == user_id,
            Bookmark.frarticle_id == frarticle_id
        )
    ).first()

    if bookmark:
        bookmark.is_bookmarked = False
        bookmark.updated_at = datetime.now(timezone.utc)
        db.commit()
        logger.info(f"Removed bookmark {bookmark.id} for user {user_id}")
        return True

    logger.warning(f"No bookmark found for user {user_id}, article {frarticle_id}")
    return False
