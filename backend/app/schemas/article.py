from datetime import datetime
from pydantic import BaseModel, ConfigDict


class ArticleResponse(BaseModel):
    """Response schema for FRArticle (Federal Register Article)"""
    model_config = ConfigDict(from_attributes=True)

    id: int
    document_number: str  # Now required - direct field from FRArticle
    title: str
    summary: str
    source_url: str
    published_at: datetime
    created_at: datetime


class ArticleDetail(ArticleResponse):
    """Detailed response including update timestamp"""
    updated_at: datetime
    fetched_at: datetime  # When raw data was fetched from Federal Register
