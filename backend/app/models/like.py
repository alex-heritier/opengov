from datetime import datetime, timezone
from sqlalchemy import Boolean, Column, DateTime, ForeignKey, Integer, Index, UniqueConstraint
from sqlalchemy.orm import Mapped, mapped_column, relationship
from app.database import Base


class Like(Base):
    """
    User likes/dislikes for articles.

    Tracks whether a user likes (positive) or dislikes (negative) an article.
    Uses a unique constraint on (user_id, frarticle_id) to prevent duplicate votes.
    """
    __tablename__ = "likes"

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

    # Like status: True for like, False for dislike
    is_positive: Mapped[bool] = mapped_column(Boolean, nullable=False)

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

    # Unique constraint: one like/dislike per user per article
    __table_args__ = (
        UniqueConstraint("user_id", "frarticle_id", name="uix_user_article_like"),
        Index("idx_likes_user_id", "user_id"),
        Index("idx_likes_frarticle_id", "frarticle_id"),
        Index("idx_likes_user_positive", "user_id", "is_positive"),
    )

    def __repr__(self):
        return f"<Like(id={self.id}, user_id={self.user_id}, frarticle_id={self.frarticle_id}, is_positive={self.is_positive})>"
