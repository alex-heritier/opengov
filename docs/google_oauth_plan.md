# Google OAuth Implementation Guide

Complete step-by-step implementation for Google OAuth 2.0 authentication with JWT tokens and localStorage.

See `docs/auth.md` for architecture overview.

---

## Phase 1: Backend Setup

### 1. Add Dependencies

```bash
cd backend
uv add authlib python-jose[cryptography] python-multipart
```

### 2. Environment Variables

Create `backend/.env`:
```bash
GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-client-secret
GOOGLE_REDIRECT_URI=http://localhost:8000/api/auth/google/callback
JWT_SECRET_KEY=$(python -c "import secrets; print(secrets.token_urlsafe(32))")
JWT_ALGORITHM=HS256
JWT_ACCESS_TOKEN_EXPIRE_MINUTES=60
FRONTEND_URL=http://localhost:5173
```

### 3. Update Config

`backend/app/config.py`:
```python
class Settings:
    # ... existing ...

    # Google OAuth
    GOOGLE_CLIENT_ID: str = os.getenv("GOOGLE_CLIENT_ID", "")
    GOOGLE_CLIENT_SECRET: str = os.getenv("GOOGLE_CLIENT_SECRET", "")
    GOOGLE_REDIRECT_URI: str = os.getenv("GOOGLE_REDIRECT_URI", "http://localhost:8000/api/auth/google/callback")

    # JWT
    JWT_SECRET_KEY: str = os.getenv("JWT_SECRET_KEY", "")
    JWT_ALGORITHM: str = os.getenv("JWT_ALGORITHM", "HS256")
    JWT_ACCESS_TOKEN_EXPIRE_MINUTES: int = int(os.getenv("JWT_ACCESS_TOKEN_EXPIRE_MINUTES", "60"))

    # Frontend
    FRONTEND_URL: str = os.getenv("FRONTEND_URL", "http://localhost:5173")
```

### 4. Create User Model

`backend/app/models/user.py`:
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
    is_active = Column(Boolean, default=True, nullable=False)
    is_verified = Column(Boolean, default=False, nullable=False)
    created_at = Column(DateTime, default=lambda: datetime.now(timezone.utc), nullable=False)
    updated_at = Column(DateTime, default=lambda: datetime.now(timezone.utc), onupdate=lambda: datetime.now(timezone.utc), nullable=False)
    last_login_at = Column(DateTime, nullable=True)
```

Update `backend/app/models/__init__.py`:
```python
from .user import User
__all__ = ["Article", "FederalRegister", "Agency", "User"]
```

### 5. Create Migration

`backend/migrations/versions/003_add_users_table.py`:
```python
"""Add users table"""
from alembic import op
import sqlalchemy as sa

revision = "003_add_users_table"
down_revision = "002_add_agencies_table"

def upgrade() -> None:
    op.create_table(
        "users",
        sa.Column("id", sa.Integer(), nullable=False),
        sa.Column("email", sa.String(255), nullable=False),
        sa.Column("google_id", sa.String(255), nullable=True),
        sa.Column("name", sa.String(255), nullable=True),
        sa.Column("picture_url", sa.String(500), nullable=True),
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
    op.drop_table("users")
```

Run: `alembic upgrade head`

---

## Phase 2: Backend Services

### Auth Service

`backend/app/services/auth.py`:
```python
from datetime import datetime, timedelta, timezone
from typing import Optional
from jose import JWTError, jwt
from app.config import settings

def create_access_token(data: dict) -> str:
    """Create JWT with 1-hour expiration"""
    to_encode = data.copy()
    expire = datetime.now(timezone.utc) + timedelta(minutes=settings.JWT_ACCESS_TOKEN_EXPIRE_MINUTES)
    to_encode.update({"exp": expire, "iat": datetime.now(timezone.utc)})
    return jwt.encode(to_encode, settings.JWT_SECRET_KEY, algorithm=settings.JWT_ALGORITHM)

def verify_access_token(token: str) -> Optional[dict]:
    """Verify and decode JWT"""
    try:
        return jwt.decode(token, settings.JWT_SECRET_KEY, algorithms=[settings.JWT_ALGORITHM])
    except JWTError:
        return None
```

### Google OAuth Service

`backend/app/services/google_oauth.py`:
```python
from authlib.integrations.starlette_client import OAuth
from app.config import settings

oauth = OAuth()
oauth.register(
    name="google",
    client_id=settings.GOOGLE_CLIENT_ID,
    client_secret=settings.GOOGLE_CLIENT_SECRET,
    server_metadata_url="https://accounts.google.com/.well-known/openid-configuration",
    client_kwargs={"scope": "openid email profile"},
)
```

### Auth Dependency

`backend/app/dependencies/auth.py`:
```python
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
    """Get current authenticated user from JWT"""
    payload = verify_access_token(credentials.credentials)
    if not payload:
        raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail="Invalid token")

    user = db.query(User).filter(User.id == int(payload.get("sub"))).first()
    if not user or not user.is_active:
        raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail="User not found")

    return user
```

---

## Phase 3: Backend Endpoints

### Auth Schemas

`backend/app/schemas/auth.py`:
```python
from datetime import datetime
from pydantic import BaseModel

class TokenResponse(BaseModel):
    access_token: str
    token_type: str = "bearer"
    expires_in: int

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

### Auth Router

`backend/app/routers/auth.py`:
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
    """Initiate Google OAuth"""
    return await oauth.google.authorize_redirect(request, settings.GOOGLE_REDIRECT_URI)

@router.get("/google/callback")
@limiter.limit("10/minute")
async def google_callback(request: Request, db: Session = Depends(get_db)):
    """Handle OAuth callback"""
    try:
        token = await oauth.google.authorize_access_token(request)
        user_info = token.get("userinfo")
        if not user_info:
            raise ValueError("No user info")

        # Find or create user
        user = db.query(User).filter(User.google_id == user_info["sub"]).first()
        if not user:
            user = db.query(User).filter(User.email == user_info["email"]).first()
            if user:
                user.google_id = user_info["sub"]
            else:
                user = User(
                    email=user_info["email"],
                    google_id=user_info["sub"],
                    name=user_info.get("name"),
                    picture_url=user_info.get("picture"),
                    is_verified=user_info.get("email_verified", False),
                )
                db.add(user)
        else:
            user.name = user_info.get("name", user.name)
            user.picture_url = user_info.get("picture", user.picture_url)

        user.last_login_at = datetime.now(timezone.utc)
        db.commit()
        db.refresh(user)

        access_token = create_access_token({"sub": str(user.id), "email": user.email})
        return RedirectResponse(f"{settings.FRONTEND_URL}/auth/callback?access_token={access_token}")

    except Exception as e:
        logger.error(f"OAuth error: {e}", exc_info=True)
        return RedirectResponse(f"{settings.FRONTEND_URL}/auth/error?message=Authentication failed")

@router.post("/renew", response_model=TokenResponse)
@limiter.limit("20/minute")
async def renew_token(request: Request, current_user: User = Depends(get_current_user)):
    """Renew access token"""
    access_token = create_access_token({"sub": str(current_user.id), "email": current_user.email})
    return TokenResponse(
        access_token=access_token,
        expires_in=settings.JWT_ACCESS_TOKEN_EXPIRE_MINUTES * 60,
    )

@router.post("/logout")
async def logout():
    """Logout (client clears token)"""
    return {"message": "Logged out"}

@router.get("/me", response_model=UserResponse)
@limiter.limit("100/minute")
async def get_me(request: Request, current_user: User = Depends(get_current_user)):
    """Get current user"""
    return current_user
```

### Register Router

`backend/app/main.py`:
```python
from app.routers import auth
app.include_router(auth.router)

# Add security headers
@app.middleware("http")
async def add_security_headers(request: Request, call_next):
    response = await call_next(request)
    response.headers["Content-Security-Policy"] = "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline';"
    response.headers["X-Frame-Options"] = "DENY"
    response.headers["X-Content-Type-Options"] = "nosniff"
    return response
```

---

## Phase 4: Frontend

### Auth Store

`frontend/src/stores/authStore.ts`:
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

interface AuthState {
  user: User | null
  accessToken: string | null
  tokenExpiresAt: number | null
  isAuthenticated: boolean
  setAuth: (accessToken: string, user: User) => void
  clearAuth: () => void
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
        const decoded = jwtDecode<{ exp: number }>(accessToken)
        set({ accessToken, user, tokenExpiresAt: decoded.exp * 1000, isAuthenticated: true })
      },

      clearAuth: () => set({ user: null, accessToken: null, tokenExpiresAt: null, isAuthenticated: false }),

      isTokenExpiringSoon: () => {
        const { tokenExpiresAt } = get()
        if (!tokenExpiresAt) return true
        return (tokenExpiresAt - Date.now()) < 10 * 60 * 1000 // <10 min
      },
    }),
    { name: 'opengov-auth', storage: createJSONStorage(() => localStorage) }
  )
)
```

### API Client

`frontend/src/api/client.ts`:
```typescript
import axios from 'axios'
import { useAuthStore } from '@/stores/authStore'

const apiClient = axios.create({
  baseURL: 'http://localhost:8000',
  headers: { 'Content-Type': 'application/json' },
})

// Auto-renew token if expiring
apiClient.interceptors.request.use(async (config) => {
  const { accessToken, isTokenExpiringSoon, setAuth, clearAuth } = useAuthStore.getState()

  if (!accessToken) return config

  if (isTokenExpiringSoon()) {
    try {
      const { data } = await axios.post('http://localhost:8000/api/auth/renew', {}, {
        headers: { Authorization: `Bearer ${accessToken}` }
      })
      const userRes = await axios.get('http://localhost:8000/api/auth/me', {
        headers: { Authorization: `Bearer ${data.access_token}` }
      })
      setAuth(data.access_token, userRes.data)
      config.headers.Authorization = `Bearer ${data.access_token}`
    } catch {
      clearAuth()
      window.location.href = '/login'
    }
  } else {
    config.headers.Authorization = `Bearer ${accessToken}`
  }

  return config
})

// Handle 401
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

### Auth Callback Page

`frontend/src/pages/AuthCallbackPage.tsx`:
```typescript
import { useEffect } from 'react'
import { useNavigate } from '@tanstack/react-router'
import { useAuthStore } from '@/stores/authStore'
import apiClient from '@/api/client'

export default function AuthCallbackPage() {
  const navigate = useNavigate()
  const setAuth = useAuthStore((state) => state.setAuth)

  useEffect(() => {
    const params = new URLSearchParams(window.location.search)
    const accessToken = params.get('access_token')

    if (!accessToken) {
      navigate({ to: '/login' })
      return
    }

    apiClient.get('/api/auth/me', {
      headers: { Authorization: `Bearer ${accessToken}` }
    }).then(({ data }) => {
      setAuth(accessToken, data)
      navigate({ to: '/feed' })
    }).catch(() => {
      navigate({ to: '/login' })
    })
  }, [navigate, setAuth])

  return (
    <div className="flex items-center justify-center min-h-screen">
      <div className="text-center">
        <h2 className="text-2xl font-bold mb-4">Signing in...</h2>
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto" />
      </div>
    </div>
  )
}
```

### Google Login Button

`frontend/src/components/auth/GoogleLoginButton.tsx`:
```typescript
export default function GoogleLoginButton() {
  return (
    <button
      onClick={() => window.location.href = 'http://localhost:8000/api/auth/google/login'}
      className="flex items-center gap-3 px-6 py-3 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors shadow-sm"
    >
      <svg className="w-5 h-5" viewBox="0 0 24 24">
        <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
        <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
        <path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/>
        <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/>
      </svg>
      <span className="font-semibold text-gray-700">Sign in with Google</span>
    </button>
  )
}
```

---

## Testing

**Backend:**
```bash
cd backend
uvicorn app.main:app --reload
# Visit http://localhost:8000/api/auth/google/login
```

**Frontend:**
```bash
cd frontend
npm run dev
# Click "Sign in with Google"
# Check localStorage for 'opengov-auth'
```

---

## Deployment

1. Set up Google OAuth credentials at https://console.cloud.google.com/
2. Add production redirect URI: `https://yourdomain.com/api/auth/google/callback`
3. Set environment variables in production
4. Enable HTTPS
5. Test OAuth flow

---

## Troubleshooting

**OAuth fails:** Check `GOOGLE_REDIRECT_URI` matches in `.env` and Google Console
**Token not persisting:** Check browser localStorage in DevTools
**Token renewal fails:** Verify `/api/auth/renew` works with valid token
**CORS errors:** Update CORS middleware to allow frontend origin
