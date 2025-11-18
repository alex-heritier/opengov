from datetime import datetime
from pydantic import BaseModel, ConfigDict


class LikeToggle(BaseModel):
    """Request schema for toggling like/dislike status"""
    frarticle_id: int
    is_positive: bool  # True for like, False for dislike


class LikeResponse(BaseModel):
    """Response schema for like operations"""
    model_config = ConfigDict(from_attributes=True)

    id: int
    user_id: int
    frarticle_id: int
    is_positive: bool
    created_at: datetime
    updated_at: datetime
