from datetime import datetime, timezone
from sqlalchemy import Boolean, Column, DateTime, ForeignKey, Integer, Index, UniqueConstraint
from sqlalchemy.orm import Mapped, mapped_column, relationship
from app.database import Base


class Bookmark(Base):
    """
    User bookmarks for articles.

    Tracks which articles a user has bookmarked for later reading.
    Uses a unique constraint on (user_id, frarticle_id) to prevent duplicate bookmarks.
    """
    __tablename__ = "bookmarks"

    # Primary key
    id: Mapped[int] = mapped_column(Integer, primary_key=True, index=True)

    # Foreign keys
    user_id: Mapped[int] = mapped_column(
        Integer,
        ForeignKey("users.id", ondelete="CASCADE"),
        nullable=False,
        index=True
    )
    frarticle_id: Mapped[int] = mapped_column(
        Integer,
        ForeignKey("frarticles.id", ondelete="CASCADE"),
        nullable=False,
        index=True
    )

    # Bookmark status
    is_bookmarked: Mapped[bool] = mapped_column(Boolean, default=True, nullable=False)

    # Timestamps
    created_at: Mapped[datetime] = mapped_column(
        DateTime,
        default=lambda: datetime.now(timezone.utc),
        nullable=False
    )
    updated_at: Mapped[datetime] = mapped_column(
        DateTime,
        default=lambda: datetime.now(timezone.utc),
        onupdate=lambda: datetime.now(timezone.utc),
        nullable=False
    )

    # Unique constraint: one bookmark per user per article
    __table_args__ = (
        UniqueConstraint("user_id", "frarticle_id", name="uix_user_article_bookmark"),
        Index("idx_bookmarks_user_id", "user_id"),
        Index("idx_bookmarks_frarticle_id", "frarticle_id"),
        Index("idx_bookmarks_user_bookmarked", "user_id", "is_bookmarked"),
    )

    def __repr__(self):
        return f"<Bookmark(id={self.id}, user_id={self.user_id}, frarticle_id={self.frarticle_id}, is_bookmarked={self.is_bookmarked})>"
