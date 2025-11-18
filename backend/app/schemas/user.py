from datetime import datetime
from typing import Optional
from fastapi_users import schemas
from pydantic import ConfigDict


class UserRead(schemas.BaseUser[int]):
    """Schema for reading user data"""
    model_config = ConfigDict(from_attributes=True)

    id: int
    email: str
    name: Optional[str] = None
    picture_url: Optional[str] = None
    google_id: Optional[str] = None
    political_leaning: Optional[str] = None
    is_active: bool
    is_superuser: bool
    is_verified: bool
    created_at: datetime
    updated_at: datetime
    last_login_at: Optional[datetime] = None


class UserCreate(schemas.BaseUserCreate):
    """Schema for creating a user"""
    name: Optional[str] = None


class UserUpdate(schemas.BaseUserUpdate):
    """Schema for updating a user"""
    name: Optional[str] = None
    picture_url: Optional[str] = None
    political_leaning: Optional[str] = None
