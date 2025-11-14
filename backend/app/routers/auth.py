"""Authentication routes for Google OAuth and JWT token management"""
import logging
from typing import Annotated

from fastapi import APIRouter, Depends, Header, Query
from fastapi.responses import RedirectResponse
from sqlalchemy.orm import Session

from app.config import settings
from app.models.user import User
from app.routers.common import get_db
from app.schemas.user import TokenResponse, UserResponse
from app.services.auth_service import auth_service
from app.exceptions import (
    MissingTokenError,
    InvalidAuthHeaderError,
    OAuthNotConfiguredError,
    OpenGovException,
)

logger = logging.getLogger(__name__)

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
async def google_login():
    """
    Initiate Google OAuth login flow

    Redirects user to Google's OAuth consent screen

    Raises:
        OAuthNotConfiguredError: If OAuth credentials not configured
    """
    # Validate configuration
    if not settings.validate_oauth_config():
        logger.error("Google OAuth not configured")
        raise OAuthNotConfiguredError()

    # Build Google OAuth URL
    google_oauth_url = (
        f"https://accounts.google.com/o/oauth2/v2/auth?"
        f"client_id={settings.GOOGLE_CLIENT_ID}&"
        f"redirect_uri={settings.GOOGLE_REDIRECT_URI}&"
        f"response_type=code&"
        f"scope=openid%20email%20profile&"
        f"access_type=offline&"
        f"prompt=consent"
    )

    return RedirectResponse(url=google_oauth_url)


@router.get("/google/callback")
async def google_callback(
    code: str = Query(..., description="Authorization code from Google"),
    db: Session = Depends(get_db),
):
    """
    Handle Google OAuth callback

    Exchanges authorization code for user info, creates/updates user,
    and redirects to frontend with JWT token

    Args:
        code: Authorization code from Google
        db: Database session

    Returns:
        Redirect to frontend with token in URL fragment or error page
    """
    try:
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
async def renew_access_token(
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
async def logout():
    """
    Logout endpoint (client-side only)

    Since we use stateless JWT tokens, logout is handled client-side
    by removing the token from localStorage. This endpoint exists for
    API completeness but doesn't perform any server-side action.

    Returns:
        Success message
    """
    return {"message": "Logout successful. Clear your token client-side."}
