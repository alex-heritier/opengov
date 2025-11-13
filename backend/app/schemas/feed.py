from typing import List
from pydantic import BaseModel
from .article import ArticleResponse


class FeedResponse(BaseModel):
    articles: List[ArticleResponse]
    page: int
    limit: int
    total: int
    has_next: bool

    class Config:
        from_attributes = True
