# Google OAuth Implementation Plan (Simplified)

## Overview

This document provides detailed implementation steps for Google OAuth 2.0 authentication with JWT tokens and **localStorage** storage. This is a simplified approach with **no refresh tokens** and **no database token storage**.

See `docs/auth.md` for high-level architecture and design decisions.

## Implementation Phases

### Phase 1: Backend Setup

#### 1.1 Add Dependencies

**File:** `backend/requirements.txt`

```txt
authlib==1.3.0
python-jose[cryptography]==3.3.0
python-multipart==0.0.6
```

**Install:**
```bash
cd backend
uv add authlib python-jose[cryptography] python-multipart
```

#### 1.2 Environment Variables

**File:** `backend/.env`

```bash
# Google OAuth
GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-client-secret
GOOGLE_REDIRECT_URI=http://localhost:8000/api/auth/google/callback

# JWT
JWT_SECRET_KEY=your-random-secret-key-min-32-chars
JWT_ALGORITHM=HS256
JWT_ACCESS_TOKEN_EXPIRE_MINUTES=60

# Frontend
FRONTEND_URL=http://localhost:5173
```

**Generate JWT secret:**
```bash
python -c "import secrets; print(secrets.token_urlsafe(32))"
```

#### 1.3 Update Config

**File:** `backend/app/config.py`

```python
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
        os.getenv("JWT_ACCESS_TOKEN_EXPIRE_MINUTES", "60")
    )

    # Frontend URL
    FRONTEND_URL: str = os.getenv("FRONTEND_URL", "http://localhost:5173")

    def validate(self):
        # ... existing validation ...

        # Validate OAuth settings in production
        if not self.DEBUG:
            if not self.GOOGLE_CLIENT_ID or not self.GOOGLE_CLIENT_SECRET:
                raise ValueError("Google OAuth credentials must be set")
            if not self.JWT_SECRET_KEY or len(self.JWT_SECRET_KEY) < 32:
                raise ValueError("JWT_SECRET_KEY must be at least 32 characters")
```

#### 1.4 Create User Model

**File:** `backend/app/models/user.py`

```python
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
    updated_at = Column(
        DateTime,
        default=lambda: datetime.now(timezone.utc),
        onupdate=lambda: datetime.now(timezone.utc),
        nullable=False,
    )
    last_login_at = Column(DateTime, nullable=True)
```

**Update:** `backend/app/models/__init__.py`

```python
from .article import Article
from .federal_register import FederalRegister
from .agency import Agency
from .user import User

__all__ = ["Article", "FederalRegister", "Agency", "User"]
```

#### 1.5 Create Database Migration

**File:** `backend/migrations/versions/003_add_users_table.py`

```python
"""Add users table

Revision ID: 003_add_users_table
Revises: 002_add_agencies_table
Create Date: 2025-11-14
"""

from alembic import op
import sqlalchemy as sa

revision = "003_add_users_table"
down_revision = "002_add_agencies_table"
branch_labels = None
depends_on = None


def upgrade() -> None:
    op.create_table(
        "users",
        sa.Column("id", sa.Integer(), nullable=False),
        sa.Column("email", sa.String(length=255), nullable=False),
        sa.Column("google_id", sa.String(length=255), nullable=True),
        sa.Column("name", sa.String(length=255), nullable=True),
        sa.Column("picture_url", sa.String(length=500), nullable=True),
        sa.Column("is_active", sa.Boolean(), nullable=False, server_default="1"),
        sa.Column("is_verified", sa.Boolean(), nullable=False, server_default="0"),
        sa.Column("created_at", sa.DateTime(), nullable=False),
        sa.Column("updated_at", sa.DateTime(), nullable=False),
        sa.Column("last_login_at", sa.DateTime(), nullable=True),
        sa.PrimaryKeyConstraint("id"),
        sa.UniqueConstraint("email"),
        sa.UniqueConstraint("google_id"),
    )
    op.create_index("ix_users_id", "users", ["id"])
    op.create_index("ix_users_email", "users", ["email"])
    op.create_index("ix_users_google_id", "users", ["google_id"])


def downgrade() -> None:
    op.drop_index("ix_users_google_id", table_name="users")
    op.drop_index("ix_users_email", table_name="users")
    op.drop_index("ix_users_id", table_name="users")
    op.drop_table("users")
```

**Run migration:**
```bash
cd backend
# Migration is already created, just run it
alembic upgrade head
```

---

### Phase 2: Backend Services

#### 2.1 Auth Service (JWT Utilities)

**File:** `backend/app/services/auth.py`

```python
from datetime import datetime, timedelta, timezone
from typing import Optional
from jose import JWTError, jwt
from app.config import settings


def create_access_token(data: dict) -> str:
    """Create JWT access token with 1-hour expiration"""
    to_encode = data.copy()
    expire = datetime.now(timezone.utc) + timedelta(
        minutes=settings.JWT_ACCESS_TOKEN_EXPIRE_MINUTES
    )
    to_encode.update({"exp": expire, "iat": datetime.now(timezone.utc)})
    encoded_jwt = jwt.encode(
        to_encode, settings.JWT_SECRET_KEY, algorithm=settings.JWT_ALGORITHM
    )
    return encoded_jwt


def verify_access_token(token: str) -> Optional[dict]:
    """Verify and decode JWT access token"""
    try:
        payload = jwt.decode(
            token, settings.JWT_SECRET_KEY, algorithms=[settings.JWT_ALGORITHM]
        )
        return payload
    except JWTError:
        return None
```

#### 2.2 Google OAuth Service

**File:** `backend/app/services/google_oauth.py`

```python
from authlib.integrations.starlette_client import OAuth
from app.config import settings

oauth = OAuth()

# Register Google OAuth provider
oauth.register(
    name="google",
    client_id=settings.GOOGLE_CLIENT_ID,
    client_secret=settings.GOOGLE_CLIENT_SECRET,
    server_metadata_url="https://accounts.google.com/.well-known/openid-configuration",
    client_kwargs={"scope": "openid email profile"},
)
```

#### 2.3 Auth Dependency

**File:** `backend/app/dependencies/auth.py`

```python
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
    db: Session = Depends(get_db),
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
            status_code=status.HTTP_401_UNAUTHORIZED, detail="Invalid token payload"
        )

    user = db.query(User).filter(User.id == int(user_id)).first()
    if not user:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED, detail="User not found"
        )

    if not user.is_active:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN, detail="User account is inactive"
        )

    return user


async def get_current_user_optional(
    credentials: Optional[HTTPAuthorizationCredentials] = Depends(security),
    db: Session = Depends(get_db),
) -> Optional[User]:
    """Get current user if authenticated, otherwise return None"""
    if not credentials:
        return None

    try:
        return await get_current_user(credentials, db)
    except HTTPException:
        return None
```

---

### Phase 3: Backend Endpoints

#### 3.1 Auth Schemas

**File:** `backend/app/schemas/auth.py`

```python
from datetime import datetime
from pydantic import BaseModel


class TokenResponse(BaseModel):
    access_token: str
    token_type: str = "bearer"
    expires_in: int  # seconds


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

**Update:** `backend/app/schemas/__init__.py`

```python
from .article import ArticleResponse, ArticleDetail
from .auth import TokenResponse, UserResponse

__all__ = ["ArticleResponse", "ArticleDetail", "TokenResponse", "UserResponse"]
```

#### 3.2 Auth Router

**File:** `backend/app/routers/auth.py`

```python
import logging
from datetime import datetime, timezone
from fastapi import APIRouter, Depends, Request
from fastapi.responses import RedirectResponse
from sqlalchemy.orm import Session
from app.routers.common import get_db, limiter
from app.models import User
from app.schemas.auth import TokenResponse, UserResponse
from app.services.auth import create_access_token
from app.services.google_oauth import oauth
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
        user_info = token.get("userinfo")
        if not user_info:
            raise ValueError("No user info in token")

        # Find or create user
        user = db.query(User).filter(User.google_id == user_info["sub"]).first()

        if not user:
            # Check if email already exists
            user = db.query(User).filter(User.email == user_info["email"]).first()
            if user:
                # Link Google account to existing user
                user.google_id = user_info["sub"]
                user.is_verified = user_info.get("email_verified", False)
            else:
                # Create new user
                user = User(
                    email=user_info["email"],
                    google_id=user_info["sub"],
                    name=user_info.get("name"),
                    picture_url=user_info.get("picture"),
                    is_verified=user_info.get("email_verified", False),
                )
                db.add(user)
        else:
            # Update existing user info
            user.name = user_info.get("name", user.name)
            user.picture_url = user_info.get("picture", user.picture_url)
            user.is_verified = user_info.get("email_verified", user.is_verified)

        # Update last login
        user.last_login_at = datetime.now(timezone.utc)
        db.commit()
        db.refresh(user)

        # Create JWT token
        access_token = create_access_token(
            data={"sub": str(user.id), "email": user.email}
        )

        # Redirect to frontend with token
        frontend_url = f"{settings.FRONTEND_URL}/auth/callback"
        redirect_url = f"{frontend_url}?access_token={access_token}"

        return RedirectResponse(url=redirect_url)

    except Exception as e:
        logger.error(f"Google OAuth callback error: {e}", exc_info=True)
        error_url = f"{settings.FRONTEND_URL}/auth/error?message=Authentication failed"
        return RedirectResponse(url=error_url)


@router.post("/renew", response_model=TokenResponse)
@limiter.limit("20/minute")
async def renew_token(request: Request, current_user: User = Depends(get_current_user)):
    """Renew access token (requires valid token)"""
    # User already validated by dependency
    access_token = create_access_token(
        data={"sub": str(current_user.id), "email": current_user.email}
    )

    return TokenResponse(
        access_token=access_token,
        token_type="bearer",
        expires_in=settings.JWT_ACCESS_TOKEN_EXPIRE_MINUTES * 60,
    )


@router.post("/logout")
@limiter.limit("20/minute")
async def logout(request: Request):
    """Logout (client clears token, no server action needed)"""
    return {"message": "Logged out successfully"}


@router.get("/me", response_model=UserResponse)
@limiter.limit("100/minute")
async def get_current_user_info(
    request: Request, current_user: User = Depends(get_current_user)
):
    """Get current authenticated user info"""
    return current_user
```

#### 3.3 Register Auth Router

**File:** `backend/app/main.py`

```python
# Add import at top
from app.routers import feed, admin, auth

# Add router registration
app.include_router(auth.router)
```

#### 3.4 Add Security Headers Middleware

**File:** `backend/app/main.py`

```python
# Add after app initialization, before routes
@app.middleware("http")
async def add_security_headers(request: Request, call_next):
    response = await call_next(request)

    # Prevent XSS attacks
    response.headers["Content-Security-Policy"] = (
        "default-src 'self'; "
        "script-src 'self' 'unsafe-inline' 'unsafe-eval'; "
        "style-src 'self' 'unsafe-inline';"
    )

    # Prevent clickjacking
    response.headers["X-Frame-Options"] = "DENY"

    # XSS protection
    response.headers["X-Content-Type-Options"] = "nosniff"

    return response
```

---

### Phase 4: Frontend Implementation

#### 4.1 Auth Store

**File:** `frontend/src/stores/authStore.ts`

```typescript
import { create } from 'zustand'
import { persist, createJSONStorage } from 'zustand/middleware'
import { jwtDecode } from 'jwt-decode'

interface User {
  id: number
  email: string
  name: string | null
  picture_url: string | null
  is_verified: boolean
}

interface JWTPayload {
  sub: string
  email: string
  exp: number
  iat: number
}

interface AuthState {
  user: User | null
  accessToken: string | null
  tokenExpiresAt: number | null
  isAuthenticated: boolean

  setAuth: (accessToken: string, user: User) => void
  clearAuth: () => void
  setUser: (user: User) => void
  isTokenExpiringSoon: () => boolean
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      accessToken: null,
      tokenExpiresAt: null,
      isAuthenticated: false,

      setAuth: (accessToken, user) => {
        const decoded = jwtDecode<JWTPayload>(accessToken)
        set({
          accessToken,
          user,
          tokenExpiresAt: decoded.exp * 1000,
          isAuthenticated: true,
        })
      },

      clearAuth: () =>
        set({
          user: null,
          accessToken: null,
          tokenExpiresAt: null,
          isAuthenticated: false,
        }),

      setUser: (user) => set({ user }),

      isTokenExpiringSoon: () => {
        const { tokenExpiresAt } = get()
        if (!tokenExpiresAt) return true
        const now = Date.now()
        const timeLeft = tokenExpiresAt - now
        return timeLeft < 10 * 60 * 1000 // Less than 10 minutes
      },
    }),
    {
      name: 'opengov-auth',
      storage: createJSONStorage(() => localStorage),
    }
  )
)
```

#### 4.2 Update API Client

**File:** `frontend/src/api/client.ts`

Add interceptors for authentication:

```typescript
import axios from 'axios'
import { useAuthStore } from '@/stores/authStore'

const apiClient = axios.create({
  baseURL: 'http://localhost:8000',
  headers: {
    'Content-Type': 'application/json',
  },
})

// Add auth token to requests and auto-renew if expiring
apiClient.interceptors.request.use(
  async (config) => {
    const { accessToken, isTokenExpiringSoon, setAuth, clearAuth } = useAuthStore.getState()

    if (!accessToken) return config

    // Renew token if expiring soon
    if (isTokenExpiringSoon()) {
      try {
        const response = await axios.post(
          'http://localhost:8000/api/auth/renew',
          {},
          {
            headers: { Authorization: `Bearer ${accessToken}` },
          }
        )

        const newToken = response.data.access_token
        const userResponse = await axios.get('http://localhost:8000/api/auth/me', {
          headers: { Authorization: `Bearer ${newToken}` },
        })

        setAuth(newToken, userResponse.data)
        config.headers.Authorization = `Bearer ${newToken}`
      } catch (error) {
        clearAuth()
        window.location.href = '/login'
        return Promise.reject(error)
      }
    } else {
      config.headers.Authorization = `Bearer ${accessToken}`
    }

    return config
  },
  (error) => Promise.reject(error)
)

// Handle 401 errors by redirecting to login
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      useAuthStore.getState().clearAuth()
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export default apiClient
```

#### 4.3 Auth Callback Page

**File:** `frontend/src/pages/AuthCallbackPage.tsx`

```typescript
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

      if (!accessToken) {
        navigate({ to: '/login', search: { error: 'Authentication failed' } })
        return
      }

      try {
        // Fetch user info
        const response = await apiClient.get('/api/auth/me', {
          headers: { Authorization: `Bearer ${accessToken}` },
        })

        setAuth(accessToken, response.data)
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

#### 4.4 Google Login Button

**File:** `frontend/src/components/auth/GoogleLoginButton.tsx`

```typescript
export default function GoogleLoginButton() {
  const handleLogin = () => {
    window.location.href = 'http://localhost:8000/api/auth/google/login'
  }

  return (
    <button
      onClick={handleLogin}
      className="flex items-center gap-3 px-6 py-3 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors shadow-sm"
    >
      <svg className="w-5 h-5" viewBox="0 0 24 24">
        <path
          fill="#4285F4"
          d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
        />
        <path
          fill="#34A853"
          d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
        />
        <path
          fill="#FBBC05"
          d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"
        />
        <path
          fill="#EA4335"
          d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
        />
      </svg>
      <span className="font-semibold text-gray-700">Sign in with Google</span>
    </button>
  )
}
```

---

## Testing

### Backend Testing

```bash
# Test OAuth login endpoint
curl http://localhost:8000/api/auth/google/login

# Test renew endpoint (with valid token)
curl -X POST http://localhost:8000/api/auth/renew \
  -H "Authorization: Bearer YOUR_TOKEN"

# Test me endpoint
curl http://localhost:8000/api/auth/me \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Frontend Testing

1. Start backend: `cd backend && uvicorn app.main:app --reload`
2. Start frontend: `cd frontend && npm run dev`
3. Click "Sign in with Google"
4. Verify redirect to Google
5. Approve access
6. Verify redirect back to app with token
7. Check localStorage for `opengov-auth` key
8. Verify token auto-renewal after 30 minutes

---

## Deployment Checklist

- [ ] Set up Google Cloud OAuth credentials
- [ ] Add production redirect URI to Google Console
- [ ] Set all environment variables in production
- [ ] Enable HTTPS in production
- [ ] Update CORS allowed origins
- [ ] Test OAuth flow in production
- [ ] Monitor token renewal behavior
- [ ] Set up error logging for auth failures

---

## Troubleshooting

**Issue**: OAuth callback fails
- Check `GOOGLE_REDIRECT_URI` matches exactly in `.env` and Google Console
- Verify Google Client ID/Secret are correct
- Check backend logs for detailed error

**Issue**: Token not persisting
- Check browser localStorage (DevTools → Application → Local Storage)
- Verify Zustand persist middleware is configured
- Check for browser privacy settings blocking localStorage

**Issue**: Token renewal fails
- Verify `/api/auth/renew` endpoint works with valid token
- Check token expiration time in JWT payload
- Verify interceptor logic in API client

**Issue**: CORS errors
- Update CORS middleware in backend to allow frontend origin
- Add `credentials: true` if using cookies (not needed for localStorage)
