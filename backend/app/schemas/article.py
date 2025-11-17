from datetime import datetime
from pydantic import BaseModel, ConfigDict


class ArticleResponse(BaseModel):
    model_config = ConfigDict(from_attributes=True)

    id: int
    title: str
    summary: str
    source_url: str
    published_at: datetime
    created_at: datetime
    document_number: str | None = None


class ArticleDetail(ArticleResponse):
    updated_at: datetime
