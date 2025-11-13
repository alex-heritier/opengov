from datetime import datetime, timezone
from sqlalchemy import Column, Integer, String, Text, DateTime, Index, ForeignKey
from app.database import Base


class Article(Base):
    __tablename__ = "articles"

    id = Column(Integer, primary_key=True, index=True)
    federal_register_id = Column(Integer, ForeignKey("federal_register_entries.id"), nullable=False, index=True)
    title = Column(String(500), nullable=False)
    summary = Column(Text, nullable=False)
    source_url = Column(String(500), nullable=False, unique=True, index=True)
    published_at = Column(DateTime, nullable=False, index=True)
    created_at = Column(DateTime, default=lambda: datetime.now(timezone.utc), nullable=False)
    updated_at = Column(DateTime, default=lambda: datetime.now(timezone.utc), onupdate=lambda: datetime.now(timezone.utc))

    # Index for efficient sorting and filtering
    __table_args__ = (
        Index("idx_published_at_desc", "published_at"),
    )
