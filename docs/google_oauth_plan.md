# Google OAuth 2.0 Integration (Phase 2)

Add Google OAuth login to the existing fastapi-users authentication system. Users can sign in with Google, and their account will be linked to their email and stored in the User model with profile data.

**Current System:** fastapi-users with email/password + cookies (see `docs/auth.md`)

**This adds:** Google OAuth option alongside email/password

---

## Overview

When implemented:
1. User clicks "Sign in with Google"
2. Redirects to Google login
3. On approval, user is authenticated
4. OpenGov creates/links user account automatically
5. Auth cookie is set (same as email/password flow)
6. User is logged in

No changes needed to the frontend's token handling or auth store—fastapi-users handles everything via cookies.

---

## Phase 1: Backend Setup

### 1. Add Dependencies

```bash
cd backend
uv add authlib[client]
```

### 2. Environment Variables

Add to `backend/.env`:
```bash
# Google OAuth (get from https://console.cloud.google.com/)
GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-client-secret
GOOGLE_REDIRECT_URI=http://localhost:8000/api/auth/google/callback

# In production:
# GOOGLE_REDIRECT_URI=https://yourdomain.com/api/auth/google/callback
```

### 3. Update Config

`backend/app/config.py`:
```python
from pydantic import SecretStr

class Settings:
    # ... existing settings ...

    # Google OAuth (Phase 2)
    GOOGLE_CLIENT_ID: str = os.getenv("GOOGLE_CLIENT_ID", "")
    GOOGLE_CLIENT_SECRET: SecretStr = SecretStr(
        os.getenv("GOOGLE_CLIENT_SECRET", "")
    )
    GOOGLE_REDIRECT_URI: str = os.getenv(
        "GOOGLE_REDIRECT_URI", 
        "http://localhost:8000/api/auth/google/callback"
    )
```

### 4. Create OAuth Service

`backend/app/services/google_oauth.py` (new file):
```python
"""Google OAuth provider for fastapi-users integration"""
from authlib.integrations.httpx_client import AsyncOAuth2Client
from app.config import settings


class GoogleOAuthProvider:
    """Helper for Google OAuth with fastapi-users"""

    @staticmethod
    def get_authorization_url(state: str) -> str:
        """Generate Google OAuth authorization URL"""
        client = AsyncOAuth2Client(
            client_id=settings.GOOGLE_CLIENT_ID,
            redirect_uri=settings.GOOGLE_REDIRECT_URI,
        )
        return client.create_authorization_url(
            "https://accounts.google.com/o/oauth2/v2/auth",
            scope=["openid", "email", "profile"],
            state=state,
        )[0]

    @staticmethod
    async def exchange_code_for_token(code: str) -> dict:
        """Exchange authorization code for access token"""
        async with AsyncOAuth2Client(
            client_id=settings.GOOGLE_CLIENT_ID,
            client_secret=settings.GOOGLE_CLIENT_SECRET.get_secret_value(),
            redirect_uri=settings.GOOGLE_REDIRECT_URI,
        ) as client:
            token = await client.fetch_token(
                "https://oauth2.googleapis.com/token",
                code=code,
            )
            return token

    @staticmethod
    async def get_user_info(token: dict) -> dict:
        """Fetch user info from Google"""
        async with AsyncOAuth2Client(
            client_id=settings.GOOGLE_CLIENT_ID,
            token=token,
        ) as client:
            response = await client.get(
                "https://openidconnect.googleapis.com/v1/userinfo",
            )
            return response.json()
```

### 5. Add OAuth Router

`backend/app/routers/oauth.py` (new file):
```python
"""Google OAuth endpoints"""
import logging
import secrets
from datetime import datetime, timezone
from fastapi import APIRouter, Depends, HTTPException, Query
from fastapi.responses import RedirectResponse
from sqlalchemy.orm import Session
from app.auth import get_user_db, get_user_manager
from app.routers.common import get_db, limiter
from app.models.user import User
from app.services.google_oauth import GoogleOAuthProvider
from app.config import settings
from fastapi_users.db import SQLAlchemyUserDatabase
from fastapi_users import BaseUserManager

logger = logging.getLogger(__name__)
router = APIRouter(prefix="/api/auth/google", tags=["oauth"])

# Store for OAuth state (in production, use Redis or database)
_oauth_states = {}


@router.get("/login")
@limiter.limit("10/minute")
async def google_login(request):
    """Initiate Google OAuth flow"""
    state = secrets.token_urlsafe(32)
    _oauth_states[state] = True
    
    auth_url = GoogleOAuthProvider.get_authorization_url(state)
    return RedirectResponse(url=auth_url)


@router.get("/callback")
@limiter.limit("10/minute")
async def google_callback(
    code: str = Query(...),
    state: str = Query(...),
    db: Session = Depends(get_db),
    user_db: SQLAlchemyUserDatabase = Depends(get_user_db),
    user_manager: BaseUserManager = Depends(get_user_manager),
):
    """Handle Google OAuth callback"""
    
    # Verify state
    if state not in _oauth_states:
        logger.warning(f"Invalid OAuth state: {state}")
        raise HTTPException(status_code=400, detail="Invalid state parameter")
    
    del _oauth_states[state]
    
    try:
        # Exchange code for token
        token = await GoogleOAuthProvider.exchange_code_for_token(code)
        
        # Get user info from Google
        user_info = await GoogleOAuthProvider.get_user_info(token)
        
        if "sub" not in user_info:
            logger.error("No Google ID in user info")
            return RedirectResponse(
                url=f"{settings.FRONTEND_URL}/login?error=invalid_user_info"
            )
        
        google_id = user_info["sub"]
        email = user_info.get("email")
        
        # Find or create user
        user = db.query(User).filter(User.google_id == google_id).first()
        
        if not user:
            # Check if email already exists (email/password user)
            user = db.query(User).filter(User.email == email).first()
            
            if user:
                # Link Google account to existing user
                user.google_id = google_id
                user.name = user_info.get("name", user.name)
                user.picture_url = user_info.get("picture", user.picture_url)
                logger.info(f"Linked Google account to existing user: {user.id}")
            else:
                # Create new user from Google info
                user = User(
                    email=email,
                    google_id=google_id,
                    name=user_info.get("name"),
                    picture_url=user_info.get("picture"),
                    is_verified=user_info.get("email_verified", False),
                    hashed_password="",  # OAuth users have no password initially
                    is_active=True,
                )
                db.add(user)
                logger.info(f"Created new user from Google: {email}")
        else:
            # Update profile info
            user.name = user_info.get("name", user.name)
            user.picture_url = user_info.get("picture", user.picture_url)
            if user_info.get("email_verified"):
                user.is_verified = True
        
        user.last_login_at = datetime.now(timezone.utc)
        db.commit()
        db.refresh(user)
        
        # Login user via fastapi-users (sets auth cookie)
        from app.auth import auth_backend
        response = RedirectResponse(url=f"{settings.FRONTEND_URL}/feed")
        
        # Get token from fastapi-users and set cookie
        token = auth_backend.get_strategy().encode_token(
            {"user_id": str(user.id)}
        )
        auth_backend.transport.set_login_cookie(response, token)
        
        logger.info(f"Google OAuth successful for user: {user.id} ({user.email})")
        return response
        
    except Exception as e:
        logger.error(f"Google OAuth error: {e}", exc_info=True)
        return RedirectResponse(
            url=f"{settings.FRONTEND_URL}/login?error=oauth_error"
        )
```

### 6. Register OAuth Router

`backend/app/main.py` (add after auth router includes):
```python
# Include OAuth router
from app.routers import oauth
app.include_router(oauth.router)
```

---

## Phase 2: Frontend

### Google Login Button

`frontend/src/components/auth/GoogleLoginButton.tsx`:
```typescript
import { Button } from '@/components/ui/button'

export default function GoogleLoginButton() {
  const handleGoogleLogin = () => {
    // Redirect to backend OAuth flow
    // Backend handles everything: Google redirect, login, cookie setting
    window.location.href = `${import.meta.env.VITE_API_URL}/api/auth/google/login`
  }

  return (
    <Button
      onClick={handleGoogleLogin}
      variant="outline"
      className="w-full flex items-center gap-3"
    >
      <svg className="w-5 h-5" viewBox="0 0 24 24">
        <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
        <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
        <path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/>
        <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/>
      </svg>
      Sign in with Google
    </Button>
  )
}
```

### Login Page Update

Add the button to `frontend/src/pages/AuthLoginPage.tsx`:
```typescript
import GoogleLoginButton from '@/components/auth/GoogleLoginButton'

export default function AuthLoginPage() {
  return (
    <div className="space-y-4">
      {/* Email/password form */}
      <form>{/* ... */}</form>
      
      {/* Divider */}
      <div className="relative">
        <div className="absolute inset-0 flex items-center">
          <div className="w-full border-t border-gray-300"></div>
        </div>
        <div className="relative flex justify-center text-sm">
          <span className="px-2 bg-white text-gray-500">Or continue with</span>
        </div>
      </div>
      
      {/* Google OAuth */}
      <GoogleLoginButton />
    </div>
  )
}
```

---

## How It Works

### User Journey

1. **Frontend:** User clicks "Sign in with Google"
2. **Frontend → Backend:** Redirect to `/api/auth/google/login`
3. **Backend → Google:** Generate OAuth URL, redirect to Google login
4. **Google:** User approves
5. **Google → Backend:** Redirect to callback with `code` & `state`
6. **Backend:** Exchange code for Google access token, fetch user info
7. **Backend:** Find or create/link User in database
8. **Backend:** Set auth cookie via fastapi-users
9. **Backend → Frontend:** Redirect to `/feed`
10. **Frontend:** Auth store detects cookie, user is logged in

### Key Differences from Email/Password Flow

- No token in URL
- No localStorage needed
- Cookie is set automatically by backend
- User profile fields updated from Google

### Account Linking

If a user signs up with email/password, then later tries Google OAuth with the same email:
- Backend finds existing user by email
- Links Google ID to that account
- User can now login with either method

---

## Deployment

### 1. Get Google OAuth Credentials

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create OAuth 2.0 credentials (Web Application)
3. Add authorized redirect URIs:
   - Development: `http://localhost:8000/api/auth/google/callback`
   - Production: `https://yourdomain.com/api/auth/google/callback`

### 2. Set Environment Variables

```bash
# Production
GOOGLE_CLIENT_ID=your-prod-client-id
GOOGLE_CLIENT_SECRET=your-prod-client-secret
GOOGLE_REDIRECT_URI=https://yourdomain.com/api/auth/google/callback
```

### 3. Production Considerations

- Use Redis for OAuth state storage (instead of in-memory `_oauth_states`)
- Ensure HTTPS is enabled
- Set `COOKIE_SECURE=true` in auth config
- Validate CORS origins

---

## State Management (Production)

The current implementation stores OAuth state in memory (`_oauth_states`). For production with multiple workers:

**Option 1: Redis (Recommended)**
```python
import redis.asyncio as redis

redis_client = redis.from_url("redis://localhost")

@router.get("/login")
async def google_login():
    state = secrets.token_urlsafe(32)
    await redis_client.set(f"oauth_state:{state}", "1", ex=600)  # 10 min expiry
    # ...

@router.get("/callback")
async def google_callback(state: str):
    exists = await redis_client.delete(f"oauth_state:{state}")
    if not exists:
        raise HTTPException(status_code=400, detail="Invalid state")
    # ...
```

**Option 2: Database**
```python
# Add OAuthState table, store state there
# Check state exists and hasn't expired
```

---

## Troubleshooting

**"Invalid state" error**
- State expired or mismatched
- Ensure state is stored correctly and checked before use

**"Google token error"**
- Check `GOOGLE_CLIENT_ID` and `GOOGLE_CLIENT_SECRET` are correct
- Verify redirect URI matches Google Console

**Cookie not being set**
- Ensure HTTPS is enabled in production
- Check `COOKIE_SECURE` matches protocol (false for HTTP, true for HTTPS)
- Verify CORS allows credentials

**User not found after redirect**
- Check database connection during callback
- Review logs for SQL errors

---

## Testing

```bash
# Start backend
cd backend
uv run dev

# Visit in browser
http://localhost:8000/api/auth/google/login

# Should redirect to Google login
# After approval, should redirect to /feed with auth cookie set
```

Check browser DevTools → Cookies to verify `opengov_auth` cookie is present.
