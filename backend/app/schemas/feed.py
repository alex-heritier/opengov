from typing import List
from pydantic import BaseModel, ConfigDict
from .article import ArticleResponse


class FeedResponse(BaseModel):
    model_config = ConfigDict(from_attributes=True)

    articles: List[ArticleResponse]
    page: int
    limit: int
    total: int
    has_next: bool
