# Twitter/X OAuth 2.0 Integration Plan

Add Twitter/X OAuth login to the existing fastapi-users authentication system. Users can sign in with their X account, and their account will be linked to their email and stored in the User model with profile data.

**Current System:** fastapi-users with email/password + Google OAuth + cookies (see `docs/auth.md`)

**This adds:** X (Twitter) OAuth option alongside email/password and Google

---

## Overview

When implemented:
1. User clicks "Sign in with X"
2. Redirects to X authorization page
3. On approval, user is authenticated
4. OpenGov creates/links user account automatically
5. Auth cookie is set (same as email/password flow)
6. User is logged in

No changes needed to the frontend's token handling or auth store—fastapi-users handles everything via cookies.

---

## Phase 1: Database Schema Updates

### 1. Add Twitter Fields to User Model

`backend/app/models/user.py`:

Add these optional fields after the existing Google OAuth fields:

```python
# Twitter/X OAuth fields (optional)
twitter_id: Mapped[Optional[str]] = mapped_column(
    String(255), unique=True, nullable=True, index=True
)
twitter_username: Mapped[Optional[str]] = mapped_column(
    String(255), nullable=True
)
```

### 2. Create Database Migration

```bash
cd backend
make db-migrate msg="Add Twitter OAuth fields to User model"
make db-upgrade
```

The migration will add:
- `twitter_id` (VARCHAR(255), unique, indexed) - X user ID (e.g., "123456789")
- `twitter_username` (VARCHAR(255)) - X handle (e.g., "@username")

**Note:** Existing `name` and `picture_url` fields will be reused for Twitter profile data.

---

## Phase 2: Backend Setup

### 1. Add Dependencies

Twitter OAuth 2.0 with PKCE requires the `authlib` library (already installed for Google OAuth):

```bash
cd backend
# authlib should already be installed from Google OAuth
# If not: uv add authlib[client]
```

### 2. Environment Variables

Add to `backend/.env`:

```bash
# X (Twitter) OAuth (get from https://developer.twitter.com/)
TWITTER_CLIENT_ID=your-client-id
TWITTER_CLIENT_SECRET=your-client-secret
TWITTER_REDIRECT_URI=http://localhost:8000/api/auth/twitter/callback

# In production:
# TWITTER_REDIRECT_URI=https://yourdomain.com/api/auth/twitter/callback
```

### 3. Update Config

`backend/app/config.py`:

```python
from pydantic import SecretStr

class Settings:
    # ... existing settings ...

    # Twitter/X OAuth
    TWITTER_CLIENT_ID: str = Field(default="", description="Twitter OAuth Client ID")
    TWITTER_CLIENT_SECRET: SecretStr = Field(
        default=SecretStr(""),
        description="Twitter OAuth Client Secret"
    )
    TWITTER_REDIRECT_URI: str = Field(
        default="http://localhost:8000/api/auth/twitter/callback",
        description="Twitter OAuth redirect URI"
    )

    @field_validator("TWITTER_CLIENT_SECRET")
    @classmethod
    def validate_twitter_secret(cls, v):
        """Validate Twitter OAuth secret"""
        # Skip validation during testing
        if "pytest" in sys.modules or "unittest" in sys.modules:
            return v
        return v

    def validate_twitter_oauth_config(self) -> bool:
        """Check if Twitter OAuth is properly configured"""
        has_client_id = bool(self.TWITTER_CLIENT_ID)
        has_client_secret = bool(
            self.TWITTER_CLIENT_SECRET.get_secret_value()
            if hasattr(self.TWITTER_CLIENT_SECRET, "get_secret_value")
            else self.TWITTER_CLIENT_SECRET
        )

        # Both should be set or both should be empty
        if has_client_id != has_client_secret:
            import logging
            logging.warning(
                "Twitter OAuth partially configured. Both TWITTER_CLIENT_ID and "
                "TWITTER_CLIENT_SECRET must be set for authentication to work."
            )
            return False

        return has_client_id and has_client_secret
```

Add validation check at the bottom of config.py:

```python
if not settings.validate_twitter_oauth_config():
    logging.warning(
        "Twitter OAuth is not configured. X login will not work. "
        "See docs/twitter_oauth_plan.md for setup instructions."
    )
```

### 4. Create Twitter OAuth Service

`backend/app/services/twitter_oauth.py` (new file):

```python
"""Twitter/X OAuth provider for fastapi-users integration"""
from authlib.integrations.httpx_client import AsyncOAuth2Client
from app.config import settings


class TwitterOAuthProvider:
    """Helper for Twitter/X OAuth 2.0 with PKCE"""

    # Twitter API v2 OAuth 2.0 endpoints
    AUTHORIZATION_URL = "https://twitter.com/i/oauth2/authorize"
    TOKEN_URL = "https://api.twitter.com/2/oauth2/token"
    USER_INFO_URL = "https://api.twitter.com/2/users/me"

    # Scopes needed for login
    # tweet.read and users.read are required
    # users.email is required to get email (needs special approval from Twitter)
    SCOPES = ["tweet.read", "users.read", "offline.access"]

    @staticmethod
    def get_authorization_url(state: str, code_verifier: str) -> str:
        """
        Generate Twitter OAuth authorization URL with PKCE

        Args:
            state: Random state for CSRF protection
            code_verifier: PKCE code verifier (store this for callback)

        Returns:
            Authorization URL to redirect user to
        """
        client = AsyncOAuth2Client(
            client_id=settings.TWITTER_CLIENT_ID,
            redirect_uri=settings.TWITTER_REDIRECT_URI,
            scope=" ".join(TwitterOAuthProvider.SCOPES),
        )

        # Generate code challenge from verifier
        from authlib.common.security import generate_token
        import hashlib
        import base64

        code_challenge = base64.urlsafe_b64encode(
            hashlib.sha256(code_verifier.encode()).digest()
        ).decode().rstrip("=")

        url, _ = client.create_authorization_url(
            TwitterOAuthProvider.AUTHORIZATION_URL,
            state=state,
            code_challenge=code_challenge,
            code_challenge_method="S256",
        )

        return url

    @staticmethod
    async def exchange_code_for_token(code: str, code_verifier: str) -> dict:
        """
        Exchange authorization code for access token

        Args:
            code: Authorization code from callback
            code_verifier: PKCE code verifier from login initiation

        Returns:
            Token response with access_token, token_type, etc.
        """
        async with AsyncOAuth2Client(
            client_id=settings.TWITTER_CLIENT_ID,
            client_secret=settings.TWITTER_CLIENT_SECRET.get_secret_value(),
            redirect_uri=settings.TWITTER_REDIRECT_URI,
        ) as client:
            token = await client.fetch_token(
                TwitterOAuthProvider.TOKEN_URL,
                code=code,
                code_verifier=code_verifier,
                grant_type="authorization_code",
            )
            return token

    @staticmethod
    async def get_user_info(token: dict) -> dict:
        """
        Fetch user info from Twitter API v2

        Args:
            token: Token dict with access_token

        Returns:
            User info dict with id, name, username, profile_image_url, etc.
        """
        async with AsyncOAuth2Client(
            client_id=settings.TWITTER_CLIENT_ID,
            token=token,
        ) as client:
            # Request additional user fields
            # Note: email requires special approval from Twitter
            response = await client.get(
                TwitterOAuthProvider.USER_INFO_URL,
                params={
                    "user.fields": "id,name,username,profile_image_url,description,created_at"
                }
            )
            data = response.json()

            # Twitter API v2 returns data in a nested "data" object
            if "data" in data:
                return data["data"]
            return data
```

### 5. Add Twitter OAuth Router

Update `backend/app/routers/oauth.py`:

```python
"""OAuth endpoints for Google and Twitter"""
import logging
import secrets
from datetime import datetime, timezone
from fastapi import APIRouter, Depends, HTTPException, Query, Request
from fastapi.responses import RedirectResponse
from sqlalchemy.ext.asyncio import AsyncSession
from app.auth import get_user_db, get_user_manager, get_async_session, auth_backend
from app.models.user import User
from app.services.google_oauth import GoogleOAuthProvider
from app.services.twitter_oauth import TwitterOAuthProvider
from app.config import settings
from fastapi_users.db import SQLAlchemyUserDatabase
from fastapi_users import BaseUserManager
from sqlalchemy import select

logger = logging.getLogger(__name__)
router = APIRouter(prefix="/api/auth", tags=["oauth"])

# Store for OAuth state and PKCE verifiers
# In production, use Redis or database with expiry
_oauth_states = {}


# ============= Google OAuth =============
# [Keep existing Google OAuth endpoints]


# ============= Twitter OAuth =============

@router.get("/twitter/login")
async def twitter_login(request: Request):
    """Initiate Twitter OAuth flow with PKCE"""
    # Generate state for CSRF protection
    state = secrets.token_urlsafe(32)

    # Generate PKCE code verifier (store for callback)
    code_verifier = secrets.token_urlsafe(32)

    # Store state and verifier (expires in 10 minutes)
    _oauth_states[state] = {
        "provider": "twitter",
        "code_verifier": code_verifier,
        "created_at": datetime.now(timezone.utc),
    }

    # Generate authorization URL
    auth_url = TwitterOAuthProvider.get_authorization_url(state, code_verifier)

    logger.info(f"Twitter OAuth initiated with state: {state}")
    return RedirectResponse(url=auth_url)


@router.get("/twitter/callback")
async def twitter_callback(
    code: str = Query(...),
    state: str = Query(...),
    session: AsyncSession = Depends(get_async_session),
    user_db: SQLAlchemyUserDatabase = Depends(get_user_db),
    user_manager: BaseUserManager = Depends(get_user_manager),
):
    """Handle Twitter OAuth callback"""

    # Verify state and get code verifier
    if state not in _oauth_states:
        logger.warning(f"Invalid OAuth state: {state}")
        return RedirectResponse(
            url=f"{settings.FRONTEND_URL}/login?error=invalid_state"
        )

    oauth_data = _oauth_states[state]
    code_verifier = oauth_data["code_verifier"]
    del _oauth_states[state]

    try:
        # Exchange code for token
        token = await TwitterOAuthProvider.exchange_code_for_token(code, code_verifier)

        # Get user info from Twitter
        user_info = await TwitterOAuthProvider.get_user_info(token)

        logger.info(f"Twitter OAuth user info: {user_info}")

        if "id" not in user_info:
            logger.error("No Twitter ID in user info")
            return RedirectResponse(
                url=f"{settings.FRONTEND_URL}/login?error=invalid_user_info"
            )

        twitter_id = user_info["id"]
        twitter_username = user_info.get("username")
        name = user_info.get("name")
        profile_image = user_info.get("profile_image_url")

        # Note: Email is NOT available by default from Twitter OAuth
        # It requires special approval and the "users.email" scope
        email = user_info.get("email")

        # Find existing user by Twitter ID
        result = await session.execute(
            select(User).where(User.twitter_id == twitter_id)
        )
        user = result.scalar_one_or_none()

        if not user and email:
            # Check if email already exists (from email/password or Google OAuth)
            result = await session.execute(
                select(User).where(User.email == email)
            )
            user = result.scalar_one_or_none()

            if user:
                # Link Twitter account to existing user
                user.twitter_id = twitter_id
                user.twitter_username = twitter_username
                user.name = name or user.name
                user.picture_url = profile_image or user.picture_url
                logger.info(f"Linked Twitter account to existing user: {user.id}")

        if not user:
            # Create new user from Twitter info
            # Problem: Twitter doesn't provide email by default
            # Solution: Generate a placeholder email or require email collection

            if not email:
                # Generate placeholder email
                # User will need to update it later or link with email/password
                email = f"{twitter_username}@twitter.placeholder"

            user = User(
                email=email,
                twitter_id=twitter_id,
                twitter_username=twitter_username,
                name=name,
                picture_url=profile_image,
                is_verified=False,  # Twitter doesn't provide email verification
                hashed_password="",  # OAuth users have no password initially
                is_active=True,
            )
            session.add(user)
            logger.info(f"Created new user from Twitter: {twitter_username}")
        else:
            # Update profile info
            user.name = name or user.name
            user.picture_url = profile_image or user.picture_url
            user.twitter_username = twitter_username

        user.last_login_at = datetime.now(timezone.utc)
        await session.commit()
        await session.refresh(user)

        # Login user via fastapi-users (sets auth cookie)
        response = RedirectResponse(url=f"{settings.FRONTEND_URL}/feed")

        # Get token from fastapi-users and set cookie
        from app.auth import get_jwt_strategy
        strategy = get_jwt_strategy()
        token_str = await strategy.write_token(user)

        # Set login cookie
        from app.auth import cookie_transport
        await cookie_transport.get_login_response(token_str, response)

        logger.info(f"Twitter OAuth successful for user: {user.id} ({user.email})")
        return response

    except Exception as e:
        logger.error(f"Twitter OAuth error: {e}", exc_info=True)
        return RedirectResponse(
            url=f"{settings.FRONTEND_URL}/login?error=oauth_error"
        )
```

### 6. Register OAuth Router

The OAuth router should already be registered in `backend/app/main.py`. If not, add:

```python
# Include OAuth router
from app.routers import oauth
app.include_router(oauth.router)
```

---

## Phase 3: Frontend

### 1. Twitter Login Button Component

`frontend/src/components/auth/TwitterLoginButton.tsx` (new file):

```typescript
import { Button } from '@/components/ui/button'

export default function TwitterLoginButton() {
  const handleTwitterLogin = () => {
    // Redirect to backend OAuth flow
    // Backend handles everything: Twitter redirect, login, cookie setting
    window.location.href = `${import.meta.env.VITE_API_URL}/api/auth/twitter/login`
  }

  return (
    <Button
      onClick={handleTwitterLogin}
      variant="outline"
      className="w-full flex items-center gap-3 bg-black hover:bg-gray-800 text-white border-black"
    >
      <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
        <path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z"/>
      </svg>
      Sign in with X
    </Button>
  )
}
```

### 2. Update Login Page

Add the Twitter button to `frontend/src/pages/AuthLoginPage.tsx`:

```typescript
import GoogleLoginButton from '@/components/auth/GoogleLoginButton'
import TwitterLoginButton from '@/components/auth/TwitterLoginButton'

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

      {/* Social OAuth buttons */}
      <div className="space-y-2">
        <GoogleLoginButton />
        <TwitterLoginButton />
      </div>
    </div>
  )
}
```

---

## How It Works

### User Journey

1. **Frontend:** User clicks "Sign in with X"
2. **Frontend → Backend:** Redirect to `/api/auth/twitter/login`
3. **Backend:** Generate PKCE code verifier and challenge, store state
4. **Backend → Twitter:** Redirect to Twitter authorization page with challenge
5. **Twitter:** User approves
6. **Twitter → Backend:** Redirect to callback with `code` & `state`
7. **Backend:** Exchange code for token using PKCE verifier
8. **Backend:** Fetch user info from Twitter API v2
9. **Backend:** Find or create/link User in database
10. **Backend:** Set auth cookie via fastapi-users
11. **Backend → Frontend:** Redirect to `/feed`
12. **Frontend:** Auth store detects cookie, user is logged in

### Key Differences from Google OAuth

- **PKCE Required:** Twitter enforces PKCE for security (code_verifier + code_challenge)
- **No Email by Default:** Twitter doesn't provide email without special approval
- **Different API Structure:** Twitter API v2 has different endpoint structure and data format
- **Different User Fields:** Twitter provides `id`, `username`, `name`, `profile_image_url`

### Account Linking

**Scenario 1:** User signs up with email/password, then tries Twitter OAuth
- If email matches (rare, Twitter doesn't provide email by default), accounts are linked
- Otherwise, creates separate account with placeholder email

**Scenario 2:** User has Google OAuth account, then tries Twitter OAuth
- If Twitter provides email and it matches Google account email, accounts are linked
- User can now login with Google OR Twitter

**Scenario 3:** User has Twitter account, wants to add email
- User can update their email in account settings
- This allows linking with other OAuth providers or email/password login

---

## Deployment

### 1. Get Twitter OAuth Credentials

1. Go to [Twitter Developer Portal](https://developer.twitter.com/)
2. Create a new app (requires Developer account approval)
3. Enable OAuth 2.0 in app settings
4. Set app type to "Web App"
5. Configure OAuth 2.0 settings:
   - **Type of App:** Web App, Automated App or Bot
   - **Callback URI:**
     - Development: `http://localhost:8000/api/auth/twitter/callback`
     - Production: `https://yourdomain.com/api/auth/twitter/callback`
   - **Website URL:** Your app's homepage
6. Copy **Client ID** and **Client Secret**

**Note:** Getting email access requires applying for elevated access and justifying the need for email data.

### 2. Set Environment Variables

```bash
# Production
TWITTER_CLIENT_ID=your-prod-client-id
TWITTER_CLIENT_SECRET=your-prod-client-secret
TWITTER_REDIRECT_URI=https://yourdomain.com/api/auth/twitter/callback
```

### 3. Production Considerations

- **Redis for State Storage:** Use Redis for OAuth state and PKCE verifiers (instead of in-memory `_oauth_states`)
- **HTTPS Required:** Ensure HTTPS is enabled
- **Secure Cookies:** Set `COOKIE_SECURE=true` in auth config
- **CORS:** Validate CORS origins
- **Rate Limits:** Twitter API v2 has rate limits (25 requests per 24 hours for /users/me)
- **Email Collection:** Consider adding an email collection step if Twitter doesn't provide it

---

## State & PKCE Management (Production)

The current implementation stores OAuth state and PKCE verifiers in memory (`_oauth_states`). For production with multiple workers:

**Redis (Recommended):**

```python
import redis.asyncio as redis

redis_client = redis.from_url("redis://localhost")

@router.get("/twitter/login")
async def twitter_login():
    state = secrets.token_urlsafe(32)
    code_verifier = secrets.token_urlsafe(32)

    # Store state and verifier with 10 min expiry
    await redis_client.set(
        f"oauth_twitter:{state}",
        code_verifier,
        ex=600
    )
    # ...

@router.get("/twitter/callback")
async def twitter_callback(state: str, code: str):
    # Retrieve and delete verifier
    code_verifier = await redis_client.get(f"oauth_twitter:{state}")
    if not code_verifier:
        raise HTTPException(status_code=400, detail="Invalid state")

    await redis_client.delete(f"oauth_twitter:{state}")
    # ...
```

---

## Email Handling Strategy

Since Twitter doesn't provide email by default, you have three options:

### Option 1: Placeholder Email (Implemented)

- Generate placeholder: `{username}@twitter.placeholder`
- Pros: Simple, no friction for user
- Cons: Not a real email, can't use for recovery

### Option 2: Collect Email After OAuth

- After Twitter OAuth succeeds, redirect to email collection page
- Store Twitter ID and require email input
- Pros: Get real email, better for account recovery
- Cons: Extra step, user friction

### Option 3: Apply for Email Access

- Apply to Twitter for elevated access with "users.email" scope
- Requires justification and approval
- Pros: Seamless experience with real email
- Cons: Approval required, not guaranteed

**Recommendation:** Start with Option 1 (placeholder), add Option 2 as enhancement later.

---

## Troubleshooting

**"Invalid state" error**
- State expired (>10 min) or mismatched
- Ensure state is stored correctly and checked before use
- Check Redis connection if using Redis

**"PKCE validation failed" error**
- Code verifier doesn't match code challenge
- Ensure verifier is stored and retrieved correctly
- Check that challenge is generated using SHA256

**"Twitter token error"**
- Check `TWITTER_CLIENT_ID` and `TWITTER_CLIENT_SECRET` are correct
- Verify redirect URI matches Twitter Developer Portal exactly
- Ensure app has OAuth 2.0 enabled

**"User object missing fields"**
- Twitter API v2 returns minimal fields by default
- Add `user.fields` parameter to request additional fields
- Check rate limits (25 requests per 24 hours for /users/me)

**Cookie not being set**
- Ensure HTTPS is enabled in production
- Check `COOKIE_SECURE` matches protocol (false for HTTP, true for HTTPS)
- Verify CORS allows credentials

**"Email not provided" issue**
- Expected behavior unless you have elevated access
- Use placeholder email strategy or collect email after OAuth
- Consider applying for elevated access with "users.email" scope

---

## Testing

```bash
# 1. Start backend
cd backend
make dev-backend

# 2. Visit in browser
http://localhost:8000/api/auth/twitter/login

# Should redirect to Twitter authorization page
# After approval, should redirect to /feed with auth cookie set
```

**Test checklist:**
- [ ] New user can sign up with Twitter
- [ ] Existing user can link Twitter to email/password account
- [ ] Existing user can link Twitter to Google account (if emails match)
- [ ] User profile (name, username, picture) is updated from Twitter
- [ ] Auth cookie is set correctly
- [ ] User is redirected to /feed after successful login
- [ ] Error cases redirect to /login with error parameter

**Check in DevTools:**
- Cookies → verify `opengov_auth` cookie is present
- Network → check OAuth flow requests
- Console → check for any errors

---

## Summary of Changes

### Database
- Add `twitter_id` (VARCHAR, unique, indexed)
- Add `twitter_username` (VARCHAR)
- Migration: `make db-migrate msg="Add Twitter OAuth fields"`

### Backend
- New service: `app/services/twitter_oauth.py`
- Update router: `app/routers/oauth.py` (add Twitter endpoints)
- Update config: `app/config.py` (add Twitter env vars)
- Update env: Add `TWITTER_CLIENT_ID`, `TWITTER_CLIENT_SECRET`, `TWITTER_REDIRECT_URI`

### Frontend
- New component: `TwitterLoginButton.tsx`
- Update: `AuthLoginPage.tsx` (add Twitter button)

### Environment Variables
```bash
TWITTER_CLIENT_ID=your-client-id
TWITTER_CLIENT_SECRET=your-client-secret
TWITTER_REDIRECT_URI=http://localhost:8000/api/auth/twitter/callback
```

---

## Resources

- [Twitter Developer Portal](https://developer.twitter.com/)
- [Twitter OAuth 2.0 Documentation](https://developer.twitter.com/en/docs/authentication/oauth-2-0)
- [Twitter API v2 User Lookup](https://developer.twitter.com/en/docs/twitter-api/users/lookup/api-reference/get-users-me)
- [OAuth 2.0 with PKCE](https://developer.twitter.com/en/docs/authentication/oauth-2-0/authorization-code)
- [Authlib Documentation](https://docs.authlib.org/)
