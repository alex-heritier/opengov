from datetime import datetime
from typing import Optional
from fastapi_users import schemas
from pydantic import BaseModel, EmailStr


class UserRead(schemas.BaseUser[int]):
    """Schema for reading user data"""
    id: int
    email: str
    name: Optional[str] = None
    picture_url: Optional[str] = None
    google_id: Optional[str] = None
    is_active: bool
    is_superuser: bool
    is_verified: bool
    created_at: datetime
    updated_at: datetime
    last_login_at: Optional[datetime] = None

    class Config:
        from_attributes = True


class UserCreate(schemas.BaseUserCreate):
    """Schema for creating a user"""
    name: Optional[str] = None


class UserUpdate(schemas.BaseUserUpdate):
    """Schema for updating a user"""
    name: Optional[str] = None
    picture_url: Optional[str] = None


# Legacy schemas for backward compatibility
class UserResponse(BaseModel):
    """Schema for user response (legacy)"""
    id: int
    email: str
    name: Optional[str] = None
    picture_url: Optional[str] = None
    google_id: Optional[str] = None
    is_active: bool
    is_verified: bool
    created_at: datetime
    updated_at: datetime
    last_login_at: Optional[datetime] = None

    class Config:
        from_attributes = True


class TokenResponse(BaseModel):
    """Schema for token response"""
    access_token: str
    token_type: str = "bearer"
    expires_in: int  # seconds


class AuthCallbackResponse(BaseModel):
    """Schema for auth callback response"""
    access_token: str
    user: UserResponse
