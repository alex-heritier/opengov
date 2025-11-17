from datetime import datetime
from pydantic import BaseModel, ConfigDict


class BookmarkToggle(BaseModel):
    """Request schema for toggling bookmark status"""
    frarticle_id: int


class BookmarkResponse(BaseModel):
    """Response schema for bookmark operations"""
    model_config = ConfigDict(from_attributes=True)

    id: int
    user_id: int
    frarticle_id: int
    is_bookmarked: bool
    created_at: datetime
    updated_at: datetime


class BookmarkedArticleResponse(BaseModel):
    """Response schema for bookmarked articles with article details"""
    model_config = ConfigDict(from_attributes=True)

    id: int
    document_number: str
    title: str
    summary: str
    source_url: str
    published_at: datetime
    created_at: datetime
    bookmarked_at: datetime  # When the user bookmarked this article
