from typing import Optional
from datetime import datetime
from pydantic import BaseModel, computed_field


class ScraperRunResponse(BaseModel):
    id: int
    started_at: datetime
    completed_at: Optional[datetime]
    processed_count: int
    skipped_count: int
    error_count: int
    success: bool
    error_message: Optional[str] = None

    class Config:
        from_attributes = True

    @computed_field
    @property
    def duration_seconds(self) -> Optional[float]:
        if self.completed_at:
            return (self.completed_at - self.started_at).total_seconds()
        return None


class ScraperRunListResponse(BaseModel):
    runs: list[ScraperRunResponse]
    total: int
