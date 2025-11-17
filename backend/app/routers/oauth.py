"""Google OAuth endpoints"""
import logging
import secrets
from datetime import datetime, timezone
from fastapi import APIRouter, Depends, HTTPException, Query, Request
from fastapi.responses import RedirectResponse
from sqlalchemy.orm import Session
from app.routers.common import get_db, limiter
from app.models.user import User
from app.services.google_oauth import GoogleOAuthProvider
from app.config import settings
from app.auth import auth_backend

logger = logging.getLogger(__name__)
router = APIRouter(prefix="/api/auth/google", tags=["oauth"])

# OAuth state store with TTL (in-memory, suitable for single-instance deployments)
# In production with multiple instances, use Redis instead
_oauth_states = {}  # Format: {state: timestamp_created}


def _cleanup_expired_states(max_age_seconds: int = 600) -> None:
    """Remove expired OAuth states (10 minute default TTL)"""
    now = datetime.now(timezone.utc).timestamp()
    expired = [state for state, ts in _oauth_states.items() if now - ts > max_age_seconds]
    for state in expired:
        _oauth_states.pop(state, None)


@router.get("/login")
@limiter.limit("10/minute")
async def google_login(request: Request):
    """Initiate Google OAuth flow"""
    _cleanup_expired_states()

    state = secrets.token_urlsafe(32)
    _oauth_states[state] = datetime.now(timezone.utc).timestamp()

    auth_url = GoogleOAuthProvider.get_authorization_url(state)
    return RedirectResponse(url=auth_url)


@router.get("/callback")
@limiter.limit("10/minute")
async def google_callback(
    request: Request,
    code: str = Query(...),
    state: str = Query(...),
    db: Session = Depends(get_db),
):
    """Handle Google OAuth callback"""

    # Verify state exists and is not expired
    if state not in _oauth_states:
        logger.warning(f"Invalid or expired OAuth state: {state}")
        raise HTTPException(status_code=400, detail="Invalid state parameter")

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

        # Set auth cookie via fastapi-users
        auth_token = await auth_backend.get_strategy().write_token(user)
        response = await auth_backend.transport.get_login_response(auth_token)

        # Redirect to frontend (cookie is set in response)
        response.status_code = 307
        response.headers["location"] = f"{settings.FRONTEND_URL}/feed"

        logger.info(f"Google OAuth successful for user: {user.id} ({user.email})")
        return response

    except ValueError as e:
        logger.error(f"Invalid Google OAuth response: {e}")
        return RedirectResponse(
            url=f"{settings.FRONTEND_URL}/login?error=invalid_response"
        )
    except RuntimeError as e:
        logger.error(f"Google OAuth token exchange failed: {e}")
        return RedirectResponse(
            url=f"{settings.FRONTEND_URL}/login?error=token_exchange_failed"
        )
    except Exception as e:
        logger.error(f"Unexpected OAuth error: {e}", exc_info=True)
        return RedirectResponse(
            url=f"{settings.FRONTEND_URL}/login?error=oauth_error"
        )
    finally:
        # Clean up state after handling callback (success or failure)
        _oauth_states.pop(state, None)
