from datetime import datetime
from pydantic import BaseModel, EmailStr


class UserBase(BaseModel):
    """Base user schema"""
    email: EmailStr
    name: str | None = None
    picture_url: str | None = None


class UserCreate(UserBase):
    """Schema for creating a user"""
    google_id: str


class UserUpdate(BaseModel):
    """Schema for updating a user"""
    name: str | None = None
    picture_url: str | None = None


class UserResponse(UserBase):
    """Schema for user response"""
    id: int
    google_id: str | None
    is_active: bool
    is_verified: bool
    created_at: datetime
    updated_at: datetime
    last_login_at: datetime | None

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
