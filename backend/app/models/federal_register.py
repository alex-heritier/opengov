from datetime import datetime
from sqlalchemy import Column, Integer, String, JSON, DateTime, Boolean, Index
from app.database import Base


class FederalRegister(Base):
    __tablename__ = "federal_register_entries"

    id = Column(Integer, primary_key=True, index=True)
    document_number = Column(String(50), nullable=False, unique=True, index=True)
    raw_data = Column(JSON, nullable=False)
    fetched_at = Column(DateTime, default=datetime.utcnow, nullable=False)
    processed = Column(Boolean, default=False, nullable=False, index=True)
    
    # Index for finding unprocessed entries
    __table_args__ = (
        Index("idx_processed_fetched", "processed", "fetched_at"),
    )
