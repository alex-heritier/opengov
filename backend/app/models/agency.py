from datetime import datetime, timezone
from sqlalchemy import Column, Integer, String, Text, JSON, DateTime, Index
from app.database import Base


class Agency(Base):
    __tablename__ = "agencies"

    id = Column(Integer, primary_key=True, index=True)
    # Federal Register agency ID
    fr_agency_id = Column(Integer, nullable=False, unique=True, index=True)
    name = Column(String(500), nullable=False)
    short_name = Column(String(200), nullable=True)
    slug = Column(String(200), nullable=False, unique=True, index=True)
    description = Column(Text, nullable=True)
    url = Column(String(500), nullable=True)
    json_url = Column(String(500), nullable=True)
    parent_id = Column(Integer, nullable=True)
    # Store complete API response for any additional fields
    raw_data = Column(JSON, nullable=False)
    created_at = Column(
        DateTime, default=lambda: datetime.now(timezone.utc), nullable=False
    )
    updated_at = Column(
        DateTime,
        default=lambda: datetime.now(timezone.utc),
        onupdate=lambda: datetime.now(timezone.utc),
        nullable=False
    )

    # Index for querying by name
    __table_args__ = (
        Index("idx_agency_name", "name"),
    )
