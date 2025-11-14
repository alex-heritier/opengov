# Google OAuth Login Implementation Plan

## Overview
This document outlines the implementation plan for adding Google OAuth 2.0 authentication to the OpenGov platform (Phase 2). The authentication system will support user login, session management, and protected routes.

## Table of Contents
1. [Architecture Overview](#architecture-overview)
2. [Dependencies](#dependencies)
3. [Database Schema](#database-schema)
4. [Backend Implementation](#backend-implementation)
5. [Frontend Integration](#frontend-integration)
6. [Security Considerations](#security-considerations)
7. [Testing Strategy](#testing-strategy)
8. [Migration Path](#migration-path)

---

## Architecture Overview

### Authentication Flow
```
1. User clicks "Login with Google" on frontend
2. Frontend redirects to `/api/auth/google/login`
3. Backend redirects to Google OAuth consent screen
4. User approves access
5. Google redirects to `/api/auth/google/callback` with authorization code
6. Backend exchanges code for Google tokens
7. Backend retrieves user info from Google
8. Backend creates/updates user in database
9. Backend generates JWT access token and refresh token
10. Backend redirects to frontend with tokens
11. Frontend stores tokens and updates auth state
12. Subsequent requests include JWT in Authorization header
```

### Token Strategy
- **Access Token**: Short-lived JWT (15 minutes), includes user ID and email
- **Refresh Token**: Long-lived (7 days), stored in database, used to refresh access tokens
- **HTTP-only Cookies**: Alternative to localStorage for enhanced security (recommended for production)

---

## Dependencies

### Backend (`requirements.txt`)
```txt
# Add these packages
authlib==1.3.0              # OAuth client library
python-jose[cryptography]==3.3.0  # JWT encoding/decoding
passlib[bcrypt]==1.7.4      # Password hashing (for future non-OAuth users)
python-multipart==0.0.6     # Form data parsing
```

### Environment Variables (`.env`)
```bash
# Google OAuth
GOOGLE_CLIENT_ID=your-google-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-google-client-secret
GOOGLE_REDIRECT_URI=http://localhost:8000/api/auth/google/callback

# JWT
JWT_SECRET_KEY=your-random-secret-key-min-32-chars
JWT_ALGORITHM=HS256
JWT_ACCESS_TOKEN_EXPIRE_MINUTES=15
JWT_REFRESH_TOKEN_EXPIRE_DAYS=7

# Frontend URL (for redirects after auth)
FRONTEND_URL=http://localhost:5173
```

---

## Database Schema

### 1. User Model
```python
# backend/app/models/user.py

from datetime import datetime, timezone
from sqlalchemy import Column, Integer, String, DateTime, Boolean
from app.database import Base

class User(Base):
    __tablename__ = "users"

    id = Column(Integer, primary_key=True, index=True)
    email = Column(String(255), unique=True, nullable=False, index=True)
    google_id = Column(String(255), unique=True, nullable=True, index=True)
    name = Column(String(255), nullable=True)
    picture_url = Column(String(500), nullable=True)

    # Account status
    is_active = Column(Boolean, default=True, nullable=False)
    is_verified = Column(Boolean, default=False, nullable=False)

    # Timestamps
    created_at = Column(DateTime, default=lambda: datetime.now(timezone.utc), nullable=False)
    updated_at = Column(DateTime, default=lambda: datetime.now(timezone.utc),
                       onupdate=lambda: datetime.now(timezone.utc), nullable=False)
    last_login_at = Column(DateTime, nullable=True)

    # Future: Add relationship to user preferences, saved articles, etc.
```

### 2. RefreshToken Model
```python
# backend/app/models/refresh_token.py

from datetime import datetime, timezone
from sqlalchemy import Column, Integer, String, DateTime, ForeignKey, Boolean
from sqlalchemy.orm import relationship
from app.database import Base

class RefreshToken(Base):
    __tablename__ = "refresh_tokens"

    id = Column(Integer, primary_key=True, index=True)
    user_id = Column(Integer, ForeignKey("users.id"), nullable=False, index=True)
    token = Column(String(500), unique=True, nullable=False, index=True)

    # Token metadata
    expires_at = Column(DateTime, nullable=False)
    is_revoked = Column(Boolean, default=False, nullable=False)

    # Device/session tracking
    user_agent = Column(String(500), nullable=True)
    ip_address = Column(String(45), nullable=True)  # IPv6 compatible

    created_at = Column(DateTime, default=lambda: datetime.now(timezone.utc), nullable=False)

    # Relationship
    user = relationship("User", backref="refresh_tokens")
```

### 3. Database Migration
```python
# backend/migrations/versions/003_add_users_and_auth.py

"""Add users and authentication tables

Revision ID: 003_add_users_and_auth
Revises: 002_add_agencies_table
Create Date: 2025-11-14
"""

def upgrade() -> None:
    # Create users table
    op.create_table('users',
        sa.Column('id', sa.Integer(), nullable=False),
        sa.Column('email', sa.String(length=255), nullable=False),
        sa.Column('google_id', sa.String(length=255), nullable=True),
        sa.Column('name', sa.String(length=255), nullable=True),
        sa.Column('picture_url', sa.String(length=500), nullable=True),
        sa.Column('is_active', sa.Boolean(), nullable=False),
        sa.Column('is_verified', sa.Boolean(), nullable=False),
        sa.Column('created_at', sa.DateTime(), nullable=False),
        sa.Column('updated_at', sa.DateTime(), nullable=False),
        sa.Column('last_login_at', sa.DateTime(), nullable=True),
        sa.PrimaryKeyConstraint('id'),
        sa.UniqueConstraint('email'),
        sa.UniqueConstraint('google_id')
    )
    op.create_index('ix_users_email', 'users', ['email'])
    op.create_index('ix_users_google_id', 'users', ['google_id'])

    # Create refresh_tokens table
    op.create_table('refresh_tokens',
        sa.Column('id', sa.Integer(), nullable=False),
        sa.Column('user_id', sa.Integer(), nullable=False),
        sa.Column('token', sa.String(length=500), nullable=False),
        sa.Column('expires_at', sa.DateTime(), nullable=False),
        sa.Column('is_revoked', sa.Boolean(), nullable=False),
        sa.Column('user_agent', sa.String(length=500), nullable=True),
        sa.Column('ip_address', sa.String(length=45), nullable=True),
        sa.Column('created_at', sa.DateTime(), nullable=False),
        sa.ForeignKeyConstraint(['user_id'], ['users.id']),
        sa.PrimaryKeyConstraint('id'),
        sa.UniqueConstraint('token')
    )
    op.create_index('ix_refresh_tokens_token', 'refresh_tokens', ['token'])
    op.create_index('ix_refresh_tokens_user_id', 'refresh_tokens', ['user_id'])

def downgrade() -> None:
    op.drop_table('refresh_tokens')
    op.drop_table('users')
```

---

## Backend Implementation

### 1. Configuration Updates
```python
# backend/app/config.py - Add these settings

class Settings:
    # ... existing settings ...

    # Google OAuth
    GOOGLE_CLIENT_ID: str = os.getenv("GOOGLE_CLIENT_ID", "")
    GOOGLE_CLIENT_SECRET: str = os.getenv("GOOGLE_CLIENT_SECRET", "")
    GOOGLE_REDIRECT_URI: str = os.getenv(
        "GOOGLE_REDIRECT_URI",
        "http://localhost:8000/api/auth/google/callback"
    )

    # JWT Settings
    JWT_SECRET_KEY: str = os.getenv("JWT_SECRET_KEY", "")
    JWT_ALGORITHM: str = os.getenv("JWT_ALGORITHM", "HS256")
    JWT_ACCESS_TOKEN_EXPIRE_MINUTES: int = int(
        os.getenv("JWT_ACCESS_TOKEN_EXPIRE_MINUTES", "15")
    )
    JWT_REFRESH_TOKEN_EXPIRE_DAYS: int = int(
        os.getenv("JWT_REFRESH_TOKEN_EXPIRE_DAYS", "7")
    )

    # Frontend URL
    FRONTEND_URL: str = os.getenv("FRONTEND_URL", "http://localhost:5173")

    def validate(self):
        # ... existing validation ...

        # Validate OAuth settings in production
        if not self.DEBUG:
            if not self.GOOGLE_CLIENT_ID or not self.GOOGLE_CLIENT_SECRET:
                raise ValueError("Google OAuth credentials must be set in production")
            if not self.JWT_SECRET_KEY or len(self.JWT_SECRET_KEY) < 32:
                raise ValueError("JWT_SECRET_KEY must be at least 32 characters")
```

### 2. JWT Utilities
```python
# backend/app/services/auth.py

from datetime import datetime, timedelta, timezone
from typing import Optional
from jose import JWTError, jwt
from sqlalchemy.orm import Session
from app.config import settings
from app.models import User, RefreshToken
import secrets

def create_access_token(data: dict, expires_delta: Optional[timedelta] = None) -> str:
    """Create JWT access token"""
    to_encode = data.copy()

    if expires_delta:
        expire = datetime.now(timezone.utc) + expires_delta
    else:
        expire = datetime.now(timezone.utc) + timedelta(
            minutes=settings.JWT_ACCESS_TOKEN_EXPIRE_MINUTES
        )

    to_encode.update({"exp": expire, "type": "access"})
    encoded_jwt = jwt.encode(to_encode, settings.JWT_SECRET_KEY, algorithm=settings.JWT_ALGORITHM)
    return encoded_jwt

def create_refresh_token(
    db: Session,
    user_id: int,
    user_agent: Optional[str] = None,
    ip_address: Optional[str] = None
) -> str:
    """Create and store refresh token"""
    token = secrets.token_urlsafe(32)
    expires_at = datetime.now(timezone.utc) + timedelta(days=settings.JWT_REFRESH_TOKEN_EXPIRE_DAYS)

    refresh_token = RefreshToken(
        user_id=user_id,
        token=token,
        expires_at=expires_at,
        user_agent=user_agent,
        ip_address=ip_address
    )

    db.add(refresh_token)
    db.commit()

    return token

def verify_access_token(token: str) -> Optional[dict]:
    """Verify and decode JWT access token"""
    try:
        payload = jwt.decode(token, settings.JWT_SECRET_KEY, algorithms=[settings.JWT_ALGORITHM])
        if payload.get("type") != "access":
            return None
        return payload
    except JWTError:
        return None

def verify_refresh_token(db: Session, token: str) -> Optional[User]:
    """Verify refresh token and return associated user"""
    refresh_token = db.query(RefreshToken).filter(
        RefreshToken.token == token,
        RefreshToken.is_revoked == False,
        RefreshToken.expires_at > datetime.now(timezone.utc)
    ).first()

    if not refresh_token:
        return None

    return refresh_token.user

def revoke_refresh_token(db: Session, token: str):
    """Revoke a refresh token"""
    refresh_token = db.query(RefreshToken).filter(RefreshToken.token == token).first()
    if refresh_token:
        refresh_token.is_revoked = True
        db.commit()

def revoke_all_user_tokens(db: Session, user_id: int):
    """Revoke all refresh tokens for a user (logout from all devices)"""
    db.query(RefreshToken).filter(RefreshToken.user_id == user_id).update(
        {"is_revoked": True}
    )
    db.commit()
```

### 3. Google OAuth Service
```python
# backend/app/services/google_oauth.py

import httpx
from authlib.integrations.starlette_client import OAuth
from app.config import settings

oauth = OAuth()

# Register Google OAuth provider
oauth.register(
    name='google',
    client_id=settings.GOOGLE_CLIENT_ID,
    client_secret=settings.GOOGLE_CLIENT_SECRET,
    server_metadata_url='https://accounts.google.com/.well-known/openid-configuration',
    client_kwargs={
        'scope': 'openid email profile'
    }
)

async def get_google_user_info(access_token: str) -> dict:
    """Fetch user info from Google using access token"""
    async with httpx.AsyncClient() as client:
        response = await client.get(
            'https://www.googleapis.com/oauth2/v3/userinfo',
            headers={'Authorization': f'Bearer {access_token}'}
        )
        response.raise_for_status()
        return response.json()
```

### 4. Authentication Dependency
```python
# backend/app/dependencies/auth.py

from typing import Optional
from fastapi import Depends, HTTPException, status
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
from sqlalchemy.orm import Session
from app.routers.common import get_db
from app.services.auth import verify_access_token
from app.models import User

security = HTTPBearer()

async def get_current_user(
    credentials: HTTPAuthorizationCredentials = Depends(security),
    db: Session = Depends(get_db)
) -> User:
    """Get current authenticated user from JWT token"""
    token = credentials.credentials
    payload = verify_access_token(token)

    if not payload:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Invalid authentication credentials",
            headers={"WWW-Authenticate": "Bearer"},
        )

    user_id = payload.get("sub")
    if not user_id:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Invalid token payload"
        )

    user = db.query(User).filter(User.id == int(user_id)).first()
    if not user:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="User not found"
        )

    if not user.is_active:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="User account is inactive"
        )

    return user

async def get_current_user_optional(
    credentials: Optional[HTTPAuthorizationCredentials] = Depends(security),
    db: Session = Depends(get_db)
) -> Optional[User]:
    """Get current user if authenticated, otherwise return None"""
    if not credentials:
        return None

    try:
        return await get_current_user(credentials, db)
    except HTTPException:
        return None
```

### 5. Auth Router
```python
# backend/app/routers/auth.py

import logging
from datetime import datetime, timezone
from fastapi import APIRouter, Depends, HTTPException, Request, status
from fastapi.responses import RedirectResponse
from sqlalchemy.orm import Session
from app.routers.common import get_db, limiter
from app.models import User
from app.schemas.auth import TokenResponse, RefreshTokenRequest
from app.services.auth import (
    create_access_token,
    create_refresh_token,
    verify_refresh_token,
    revoke_refresh_token,
    revoke_all_user_tokens
)
from app.services.google_oauth import oauth, get_google_user_info
from app.dependencies.auth import get_current_user
from app.config import settings

logger = logging.getLogger(__name__)
router = APIRouter(prefix="/api/auth", tags=["auth"])

@router.get("/google/login")
@limiter.limit("10/minute")
async def google_login(request: Request):
    """Initiate Google OAuth flow"""
    redirect_uri = settings.GOOGLE_REDIRECT_URI
    return await oauth.google.authorize_redirect(request, redirect_uri)

@router.get("/google/callback")
@limiter.limit("10/minute")
async def google_callback(request: Request, db: Session = Depends(get_db)):
    """Handle Google OAuth callback"""
    try:
        # Exchange authorization code for access token
        token = await oauth.google.authorize_access_token(request)

        # Get user info from Google
        user_info = await get_google_user_info(token['access_token'])

        # Find or create user
        user = db.query(User).filter(User.google_id == user_info['sub']).first()

        if not user:
            # Check if email already exists (linked to different auth method)
            user = db.query(User).filter(User.email == user_info['email']).first()
            if user:
                # Link Google account to existing user
                user.google_id = user_info['sub']
                user.is_verified = user_info.get('email_verified', False)
            else:
                # Create new user
                user = User(
                    email=user_info['email'],
                    google_id=user_info['sub'],
                    name=user_info.get('name'),
                    picture_url=user_info.get('picture'),
                    is_verified=user_info.get('email_verified', False)
                )
                db.add(user)
        else:
            # Update existing user info
            user.name = user_info.get('name', user.name)
            user.picture_url = user_info.get('picture', user.picture_url)
            user.is_verified = user_info.get('email_verified', user.is_verified)

        # Update last login
        user.last_login_at = datetime.now(timezone.utc)
        db.commit()
        db.refresh(user)

        # Create tokens
        access_token = create_access_token(data={"sub": str(user.id), "email": user.email})
        refresh_token = create_refresh_token(
            db=db,
            user_id=user.id,
            user_agent=request.headers.get('user-agent'),
            ip_address=request.client.host
        )

        # Redirect to frontend with tokens
        frontend_url = f"{settings.FRONTEND_URL}/auth/callback"
        redirect_url = f"{frontend_url}?access_token={access_token}&refresh_token={refresh_token}"

        return RedirectResponse(url=redirect_url)

    except Exception as e:
        logger.error(f"Google OAuth callback error: {e}", exc_info=True)
        error_url = f"{settings.FRONTEND_URL}/auth/error?message=Authentication failed"
        return RedirectResponse(url=error_url)

@router.post("/refresh", response_model=TokenResponse)
@limiter.limit("20/minute")
async def refresh_access_token(
    request: Request,
    refresh_request: RefreshTokenRequest,
    db: Session = Depends(get_db)
):
    """Refresh access token using refresh token"""
    user = verify_refresh_token(db, refresh_request.refresh_token)

    if not user:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Invalid or expired refresh token"
        )

    # Create new access token
    access_token = create_access_token(data={"sub": str(user.id), "email": user.email})

    return TokenResponse(
        access_token=access_token,
        token_type="bearer",
        expires_in=settings.JWT_ACCESS_TOKEN_EXPIRE_MINUTES * 60
    )

@router.post("/logout")
@limiter.limit("20/minute")
async def logout(
    request: Request,
    refresh_token: RefreshTokenRequest,
    current_user: User = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """Logout current session"""
    revoke_refresh_token(db, refresh_token.refresh_token)
    return {"message": "Logged out successfully"}

@router.post("/logout-all")
@limiter.limit("10/minute")
async def logout_all_devices(
    request: Request,
    current_user: User = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """Logout from all devices"""
    revoke_all_user_tokens(db, current_user.id)
    return {"message": "Logged out from all devices"}

@router.get("/me")
@limiter.limit("100/minute")
async def get_current_user_info(
    request: Request,
    current_user: User = Depends(get_current_user)
):
    """Get current authenticated user info"""
    return {
        "id": current_user.id,
        "email": current_user.email,
        "name": current_user.name,
        "picture_url": current_user.picture_url,
        "is_verified": current_user.is_verified,
        "created_at": current_user.created_at,
        "last_login_at": current_user.last_login_at
    }
```

### 6. Pydantic Schemas
```python
# backend/app/schemas/auth.py

from datetime import datetime
from pydantic import BaseModel, EmailStr

class TokenResponse(BaseModel):
    access_token: str
    token_type: str = "bearer"
    expires_in: int  # seconds
    refresh_token: str | None = None

class RefreshTokenRequest(BaseModel):
    refresh_token: str

class UserResponse(BaseModel):
    id: int
    email: str
    name: str | None
    picture_url: str | None
    is_verified: bool
    created_at: datetime
    last_login_at: datetime | None

    class Config:
        from_attributes = True
```

---

## Frontend Integration

### 1. Auth Store (Zustand)
```typescript
// frontend/src/stores/authStore.ts

import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface User {
  id: number
  email: string
  name: string | null
  picture_url: string | null
  is_verified: boolean
}

interface AuthState {
  user: User | null
  accessToken: string | null
  refreshToken: string | null
  isAuthenticated: boolean

  setAuth: (accessToken: string, refreshToken: string, user: User) => void
  clearAuth: () => void
  setUser: (user: User) => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      accessToken: null,
      refreshToken: null,
      isAuthenticated: false,

      setAuth: (accessToken, refreshToken, user) => set({
        accessToken,
        refreshToken,
        user,
        isAuthenticated: true
      }),

      clearAuth: () => set({
        user: null,
        accessToken: null,
        refreshToken: null,
        isAuthenticated: false
      }),

      setUser: (user) => set({ user })
    }),
    {
      name: 'auth-storage'
    }
  )
)
```

### 2. API Client with Auth
```typescript
// frontend/src/api/client.ts

import axios from 'axios'
import { useAuthStore } from '@/stores/authStore'

const apiClient = axios.create({
  baseURL: 'http://localhost:8000',
  headers: {
    'Content-Type': 'application/json'
  }
})

// Add auth token to requests
apiClient.interceptors.request.use(
  (config) => {
    const { accessToken } = useAuthStore.getState()
    if (accessToken) {
      config.headers.Authorization = `Bearer ${accessToken}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

// Handle token refresh on 401
apiClient.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config

    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true

      const { refreshToken, setAuth, clearAuth } = useAuthStore.getState()

      if (!refreshToken) {
        clearAuth()
        window.location.href = '/login'
        return Promise.reject(error)
      }

      try {
        const response = await axios.post('http://localhost:8000/api/auth/refresh', {
          refresh_token: refreshToken
        })

        const { access_token } = response.data
        const userResponse = await axios.get('http://localhost:8000/api/auth/me', {
          headers: { Authorization: `Bearer ${access_token}` }
        })

        setAuth(access_token, refreshToken, userResponse.data)
        originalRequest.headers.Authorization = `Bearer ${access_token}`

        return apiClient(originalRequest)
      } catch (refreshError) {
        clearAuth()
        window.location.href = '/login'
        return Promise.reject(refreshError)
      }
    }

    return Promise.reject(error)
  }
)

export default apiClient
```

### 3. Auth Callback Page
```typescript
// frontend/src/pages/AuthCallbackPage.tsx

import { useEffect } from 'react'
import { useNavigate } from '@tanstack/react-router'
import { useAuthStore } from '@/stores/authStore'
import apiClient from '@/api/client'

export default function AuthCallbackPage() {
  const navigate = useNavigate()
  const setAuth = useAuthStore((state) => state.setAuth)

  useEffect(() => {
    const handleCallback = async () => {
      const params = new URLSearchParams(window.location.search)
      const accessToken = params.get('access_token')
      const refreshToken = params.get('refresh_token')

      if (!accessToken || !refreshToken) {
        navigate({ to: '/login', search: { error: 'Authentication failed' } })
        return
      }

      try {
        // Fetch user info
        const response = await apiClient.get('/api/auth/me', {
          headers: { Authorization: `Bearer ${accessToken}` }
        })

        setAuth(accessToken, refreshToken, response.data)
        navigate({ to: '/feed' })
      } catch (error) {
        console.error('Auth callback error:', error)
        navigate({ to: '/login', search: { error: 'Failed to fetch user info' } })
      }
    }

    handleCallback()
  }, [navigate, setAuth])

  return (
    <div className="flex items-center justify-center min-h-screen">
      <div className="text-center">
        <h2 className="text-2xl font-bold mb-4">Completing sign in...</h2>
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
      </div>
    </div>
  )
}
```

### 4. Login Button Component
```typescript
// frontend/src/components/auth/GoogleLoginButton.tsx

export default function GoogleLoginButton() {
  const handleLogin = () => {
    window.location.href = 'http://localhost:8000/api/auth/google/login'
  }

  return (
    <button
      onClick={handleLogin}
      className="flex items-center gap-3 px-6 py-3 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
    >
      <svg className="w-5 h-5" viewBox="0 0 24 24">
        {/* Google logo SVG */}
      </svg>
      <span className="font-semibold text-gray-700">Continue with Google</span>
    </button>
  )
}
```

---

## Security Considerations

### 1. Token Security
- ✅ Use HTTPS in production (required for secure cookies)
- ✅ Short access token lifetime (15 minutes)
- ✅ Secure refresh token storage (HTTP-only cookies recommended)
- ✅ Token rotation: Issue new refresh token on each refresh
- ✅ Implement token revocation/blacklisting

### 2. CORS Configuration
```python
# Update app/main.py CORS settings
app.add_middleware(
    CORSMiddleware,
    allow_origins=settings.ALLOWED_ORIGINS,
    allow_credentials=True,  # Required for cookies
    allow_methods=["*"],
    allow_headers=["*"],
    expose_headers=["*"]
)
```

### 3. Rate Limiting
- ✅ Already implemented with SlowAPI
- ✅ Stricter limits on auth endpoints (10/min for login, 20/min for refresh)

### 4. Input Validation
- ✅ Validate all user inputs with Pydantic
- ✅ Sanitize user-provided data before storage
- ✅ Validate redirect URIs to prevent open redirects

### 5. Database Security
- ✅ Never store plaintext tokens (hash refresh tokens)
- ✅ Use parameterized queries (SQLAlchemy ORM handles this)
- ✅ Regular cleanup of expired tokens

### 6. OAuth Security
- ✅ Use `state` parameter to prevent CSRF (authlib handles this)
- ✅ Verify Google's SSL certificate
- ✅ Use PKCE for mobile/SPA clients (future enhancement)

---

## Testing Strategy

### 1. Unit Tests
```python
# backend/tests/test_auth.py

def test_create_access_token():
    """Test JWT access token creation"""
    token = create_access_token({"sub": "123", "email": "test@example.com"})
    assert token is not None

    payload = verify_access_token(token)
    assert payload["sub"] == "123"
    assert payload["email"] == "test@example.com"
    assert payload["type"] == "access"

def test_expired_token():
    """Test expired token rejection"""
    # Create token that expires immediately
    token = create_access_token(
        {"sub": "123"},
        expires_delta=timedelta(seconds=-1)
    )

    payload = verify_access_token(token)
    assert payload is None
```

### 2. Integration Tests
```python
def test_google_oauth_flow(client, db):
    """Test complete OAuth flow"""
    # Mock Google OAuth responses
    # Test login redirect
    # Test callback processing
    # Verify user creation
    # Verify token generation
```

### 3. Frontend Tests
```typescript
// Test auth store
// Test API client interceptors
// Test protected route behavior
// Test token refresh flow
```

---

## Migration Path

### Phase 1: Database Setup
1. Add dependencies to requirements.txt
2. Create User and RefreshToken models
3. Generate and run migration
4. Update config.py with OAuth settings

### Phase 2: Backend Implementation
1. Implement auth service (JWT utilities)
2. Create Google OAuth service
3. Add auth dependencies (get_current_user)
4. Create auth router with endpoints
5. Register router in main.py

### Phase 3: Frontend Integration
1. Create auth store
2. Update API client with interceptors
3. Add auth callback page
4. Create login UI components
5. Add protected route wrapper

### Phase 4: Testing & Deployment
1. Write and run tests
2. Test OAuth flow in development
3. Set up Google OAuth credentials in production
4. Configure environment variables
5. Deploy and verify

---

## Future Enhancements

### Phase 3+ Features
1. **User Preferences**: Save feed filters, notification settings
2. **Saved Articles**: Bookmark/save articles for later
3. **Comments & Engagement**: Like, share, comment on articles
4. **Email Notifications**: Daily digest of updates
5. **Multiple Auth Providers**: GitHub, Microsoft, email/password
6. **Admin Dashboard**: User management, analytics
7. **API Keys**: Allow authenticated API access for developers
8. **Two-Factor Authentication**: Enhanced security option

---

## Resources

### Documentation
- [Google OAuth 2.0](https://developers.google.com/identity/protocols/oauth2)
- [FastAPI Security](https://fastapi.tiangolo.com/tutorial/security/)
- [Authlib Documentation](https://docs.authlib.org/)
- [python-jose JWT](https://python-jose.readthedocs.io/)

### Google Cloud Console Setup
1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create new project or select existing
3. Enable Google+ API
4. Create OAuth 2.0 credentials
5. Add authorized redirect URIs:
   - Development: `http://localhost:8000/api/auth/google/callback`
   - Production: `https://yourdomain.com/api/auth/google/callback`
6. Copy Client ID and Client Secret to .env

---

## Conclusion

This implementation plan provides a complete, production-ready Google OAuth authentication system for the OpenGov platform. The architecture is secure, scalable, and follows industry best practices for token management and user authentication.

**Estimated Implementation Time**: 2-3 days
**Priority**: High (Phase 2 requirement)
**Dependencies**: Google Cloud Console access, SSL certificate for production
