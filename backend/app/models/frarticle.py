from datetime import datetime, timezone
from sqlalchemy import Column, Integer, String, Text, JSON, DateTime, Boolean, Index
from app.database import Base


class FRArticle(Base):
    """
    Unified model combining Federal Register raw data and processed article content.

    Each Federal Register document becomes one article with both raw API data
    and AI-processed summary for the public feed.
    """
    __tablename__ = "frarticles"

    # Primary key
    id = Column(Integer, primary_key=True, index=True)

    # Federal Register raw data fields
    document_number = Column(String(50), nullable=False, unique=True, index=True)
    raw_data = Column(JSON, nullable=False)  # Complete API response for audit/debugging
    fetched_at = Column(DateTime, nullable=False)

    # Processed article fields (user-facing)
    title = Column(String(500), nullable=False)
    summary = Column(Text, nullable=False)  # AI-generated viral summary
    source_url = Column(String(500), nullable=False, unique=True, index=True)
    published_at = Column(DateTime, nullable=False, index=True)

    # Metadata
    created_at = Column(
        DateTime, default=lambda: datetime.now(timezone.utc), nullable=False
    )
    updated_at = Column(
        DateTime,
        default=lambda: datetime.now(timezone.utc),
        onupdate=lambda: datetime.now(timezone.utc)
    )

    # Indexes for efficient querying
    __table_args__ = (
        Index("idx_frarticles_published_at_desc", "published_at"),
        Index("idx_frarticles_fetched_at", "fetched_at"),
    )
