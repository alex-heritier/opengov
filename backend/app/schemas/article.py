from datetime import datetime
from pydantic import BaseModel


class ArticleResponse(BaseModel):
    id: int
    title: str
    summary: str
    source_url: str
    published_at: datetime
    created_at: datetime

    class Config:
        from_attributes = True


class ArticleDetail(ArticleResponse):
    updated_at: datetime

    class Config:
        from_attributes = True
