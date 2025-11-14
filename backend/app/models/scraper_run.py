from sqlalchemy import Column, Integer, DateTime, String, Boolean
from app.database import Base


class ScraperRun(Base):
    __tablename__ = "scraper_runs"

    id = Column(Integer, primary_key=True, index=True)
    started_at = Column(DateTime, nullable=False, index=True)
    completed_at = Column(DateTime, nullable=True)
    processed_count = Column(Integer, default=0)
    skipped_count = Column(Integer, default=0)
    error_count = Column(Integer, default=0)
    success = Column(Boolean, default=False)
    error_message = Column(String(500), nullable=True)
