from datetime import datetime
from pydantic import BaseModel, Field


class ArticleCreate(BaseModel):
    title: str = Field(..., max_length=500, min_length=1)
    summary: str = Field(..., max_length=5000, min_length=1)
    source_url: str = Field(..., max_length=500, min_length=1)
    published_at: datetime


class ArticleResponse(BaseModel):
    id: int
    title: str
    summary: str
    source_url: str
    published_at: datetime
    created_at: datetime
    document_number: str | None = None

    class Config:
        from_attributes = True


class ArticleDetail(ArticleResponse):
    updated_at: datetime
    document_number: str | None = None

    class Config:
        from_attributes = True
