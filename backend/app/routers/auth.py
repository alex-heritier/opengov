"""Authentication routes for Google OAuth and JWT token management"""
import logging
import secrets
from typing import Annotated

from fastapi import APIRouter, Depends, Header, Query, Request
from fastapi.responses import RedirectResponse
from sqlalchemy.orm import Session
from cachetools import TTLCache

from app.config import settings
from app.models.user import User
from app.routers.common import get_db, limiter
from app.schemas.user import TokenResponse, UserResponse
from app.services.auth import auth_service
from app.exceptions import (
    MissingTokenError,
    InvalidAuthHeaderError,
    OAuthNotConfiguredError,
    OpenGovException,
)

logger = logging.getLogger(__name__)

# State parameter cache for CSRF protection (5 minute TTL)
_oauth_state_cache = TTLCache(maxsize=1000, ttl=300)

router = APIRouter(prefix="/api/auth", tags=["auth"])


def get_current_user(
    authorization: Annotated[str | None, Header()] = None,
    db: Session = Depends(get_db),
) -> User:
    """
    Dependency to get current authenticated user from JWT token

    Args:
        authorization: Authorization header with Bearer token
        db: Database session

    Returns:
        Current user

    Raises:
        MissingTokenError: If authorization header is missing
        InvalidAuthHeaderError: If authorization header format is invalid
        TokenExpiredError: If token has expired
        InvalidTokenError: If token is invalid
        UserNotFoundError: If user not found
        InactiveUserError: If user is inactive
    """
    if not authorization:
        raise MissingTokenError()

    parts = authorization.split()
    if len(parts) != 2 or parts[0].lower() != "bearer":
        raise InvalidAuthHeaderError()

    token = parts[1]
    return auth_service.get_user_from_token(token, db)


@router.get("/google/login")
@limiter.limit("10/minute")
async def google_login(request: Request):
    """
    Initiate Google OAuth login flow

    Redirects user to Google's OAuth consent screen with CSRF state parameter

    Raises:
        OAuthNotConfiguredError: If OAuth credentials not configured
    """
    # Validate configuration
    if not settings.validate_oauth_config():
        logger.error("Google OAuth not configured")
        raise OAuthNotConfiguredError()

    # Generate cryptographically secure state parameter for CSRF protection
    state = secrets.token_urlsafe(32)
    _oauth_state_cache[state] = True

    # Build Google OAuth URL with state parameter
    google_oauth_url = (
        f"https://accounts.google.com/o/oauth2/v2/auth?"
        f"client_id={settings.GOOGLE_CLIENT_ID}&"
        f"redirect_uri={settings.GOOGLE_REDIRECT_URI}&"
        f"response_type=code&"
        f"scope=openid%20email%20profile&"
        f"state={state}&"
        f"access_type=offline&"
        f"prompt=consent"
    )

    logger.debug("Generated OAuth state parameter for login flow")
    return RedirectResponse(url=google_oauth_url)


@router.get("/google/callback")
@limiter.limit("20/minute")
async def google_callback(
    request: Request,
    code: str = Query(..., description="Authorization code from Google"),
    state: str = Query(..., description="CSRF protection state parameter"),
    db: Session = Depends(get_db),
):
    """
    Handle Google OAuth callback

    Verifies CSRF state, exchanges authorization code for user info,
    creates/updates user, and redirects to frontend with JWT token

    Args:
        code: Authorization code from Google
        state: CSRF protection state parameter
        db: Database session

    Returns:
        Redirect to frontend with token in URL fragment or error page
    """
    try:
        # Verify state parameter to prevent CSRF attacks
        if state not in _oauth_state_cache:
            logger.warning("Invalid or expired OAuth state parameter")
            error_redirect = f"{settings.FRONTEND_URL}/auth/error?message=invalid_state"
            return RedirectResponse(url=error_redirect)

        # Remove state from cache (one-time use)
        del _oauth_state_cache[state]

        # Exchange code for user info
        user_info = await auth_service.exchange_code_for_user_info(code)

        # Create or update user
        user = auth_service.create_or_update_user(
            db=db,
            google_id=user_info["google_id"],
            email=user_info["email"],
            name=user_info["name"],
            picture_url=user_info["picture_url"],
        )

        # Create JWT token
        access_token = auth_service.create_access_token(data={"sub": user.id})

        # Redirect to frontend with token
        # Using URL fragment (#) so token isn't sent to server
        frontend_redirect = (
            f"{settings.FRONTEND_URL}/auth/callback"
            f"#access_token={access_token}"
        )

        logger.info(f"Successful OAuth login for user: {user.email}")
        return RedirectResponse(url=frontend_redirect)

    except OpenGovException:
        # Re-raise custom exceptions (handled by exception handler)
        raise
    except Exception as e:
        # Catch unexpected errors
        logger.error(f"Unexpected OAuth callback error: {e}", exc_info=True)
        error_redirect = f"{settings.FRONTEND_URL}/auth/error?message=unexpected_error"
        return RedirectResponse(url=error_redirect)


@router.post("/renew", response_model=TokenResponse)
@limiter.limit("30/minute")
async def renew_access_token(
    request: Request,
    authorization: Annotated[str | None, Header()] = None,
    db: Session = Depends(get_db),
):
    """
    Renew JWT access token

    Requires a valid non-expired token. Returns a new token with fresh expiration.

    Args:
        authorization: Authorization header with Bearer token
        db: Database session

    Returns:
        New access token

    Raises:
        MissingTokenError: If authorization header is missing
        InvalidAuthHeaderError: If authorization header format is invalid
        TokenExpiredError: If token has expired
        InvalidTokenError: If token is invalid
    """
    if not authorization:
        raise MissingTokenError()

    parts = authorization.split()
    if len(parts) != 2 or parts[0].lower() != "bearer":
        raise InvalidAuthHeaderError()

    current_token = parts[1]
    new_token = auth_service.renew_token(current_token, db)

    return TokenResponse(
        access_token=new_token,
        expires_in=settings.JWT_ACCESS_TOKEN_EXPIRE_MINUTES * 60,
    )


@router.get("/me", response_model=UserResponse)
async def get_current_user_info(
    current_user: User = Depends(get_current_user),
):
    """
    Get current authenticated user information

    Args:
        current_user: Current user from JWT token

    Returns:
        User information
    """
    return current_user


@router.post("/logout")
@limiter.limit("10/minute")
async def logout(
    request: Request,
    authorization: Annotated[str | None, Header()] = None,
):
    """
    Logout endpoint

    Invalidates the current token from the cache and instructs client
    to remove it from localStorage. Since we use stateless JWT tokens,
    the token will still be valid until expiration, but it won't be cached
    server-side.

    Args:
        authorization: Authorization header with Bearer token

    Returns:
        Success message
    """
    # Invalidate token from cache if provided
    if authorization:
        parts = authorization.split()
        if len(parts) == 2 and parts[0].lower() == "bearer":
            token = parts[1]
            auth_service.invalidate_token(token)
            logger.info("User logged out and token invalidated from cache")

    return {"message": "Logout successful. Clear your token client-side."}
