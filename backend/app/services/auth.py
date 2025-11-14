"""Refactored authentication service with dependency injection and caching"""
import logging
from datetime import datetime, timedelta, timezone
from typing import Optional, Protocol

import httpx
from jose import JWTError, jwt
from sqlalchemy.orm import Session
from cachetools import TTLCache

from app.config import settings
from app.models.user import User
from app.exceptions import (
    TokenExpiredError,
    InvalidTokenError,
    UserNotFoundError,
    InactiveUserError,
    OAuthCodeExchangeError,
    OAuthUserInfoError,
    OAuthNotConfiguredError,
)

logger = logging.getLogger(__name__)


# Protocol for HTTP client (enables dependency injection)
class HTTPClient(Protocol):
    """Protocol for async HTTP client"""
    async def post(self, url: str, **kwargs) -> httpx.Response: ...
    async def get(self, url: str, **kwargs) -> httpx.Response: ...


# Token cache: hash(token) -> User ID
# TTL of 5 minutes to balance performance and freshness
_token_cache = TTLCache(maxsize=1000, ttl=300)


def _get_jwt_secret() -> str:
    """Get JWT secret key as string"""
    if hasattr(settings.JWT_SECRET_KEY, "get_secret_value"):
        return settings.JWT_SECRET_KEY.get_secret_value()
    return str(settings.JWT_SECRET_KEY)


class AuthService:
    """Authentication service with dependency injection"""

    def __init__(self, http_client: Optional[HTTPClient] = None):
        """
        Initialize auth service

        Args:
            http_client: Optional HTTP client for OAuth requests.
                        If None, uses httpx.AsyncClient
        """
        self._http_client = http_client

    async def _get_http_client(self):
        """Get HTTP client (creates new AsyncClient if not injected)"""
        if self._http_client:
            return self._http_client

        # Return a context manager for httpx.AsyncClient
        return httpx.AsyncClient()

    def create_access_token(
        self,
        data: dict,
        expires_delta: Optional[timedelta] = None
    ) -> str:
        """
        Create a JWT access token

        Args:
            data: Data to encode in the token
            expires_delta: Optional expiration time delta

        Returns:
            Encoded JWT token
        """
        to_encode = data.copy()

        # Convert sub to string if it's an integer (JWT RFC 7519 requires string)
        if "sub" in to_encode and isinstance(to_encode["sub"], int):
            to_encode["sub"] = str(to_encode["sub"])

        if expires_delta:
            expire = datetime.now(timezone.utc) + expires_delta
        else:
            expire = datetime.now(timezone.utc) + timedelta(
                minutes=settings.JWT_ACCESS_TOKEN_EXPIRE_MINUTES
            )
        to_encode.update({"exp": expire, "iat": datetime.now(timezone.utc)})

        encoded_jwt = jwt.encode(
            to_encode,
            _get_jwt_secret(),
            algorithm=settings.JWT_ALGORITHM
        )
        return encoded_jwt

    def decode_access_token(self, token: str) -> dict:
        """
        Decode and validate a JWT access token

        Args:
            token: JWT token to decode

        Returns:
            Decoded token payload

        Raises:
            TokenExpiredError: If token has expired
            InvalidTokenError: If token is invalid
        """
        try:
            payload = jwt.decode(
                token,
                _get_jwt_secret(),
                algorithms=[settings.JWT_ALGORITHM]
            )
            return payload
        except jwt.ExpiredSignatureError:
            logger.warning("Token expired")
            raise TokenExpiredError()
        except JWTError as e:
            logger.warning(f"Invalid token: {e}")
            raise InvalidTokenError()

    def get_user_from_token(self, token: str, db: Session) -> User:
        """
        Get user from JWT token with caching

        Args:
            token: JWT token
            db: Database session

        Returns:
            User object

        Raises:
            TokenExpiredError: If token has expired
            InvalidTokenError: If token is invalid
            UserNotFoundError: If user not found
            InactiveUserError: If user is inactive
        """
        # Check cache first
        token_hash = hash(token)
        cached_user_id = _token_cache.get(token_hash)

        if cached_user_id:
            user = db.query(User).filter(User.id == cached_user_id).first()
            if user and user.is_active:
                return user
            # Cache miss or invalid - continue to full validation

        # Decode token (validates expiration)
        payload = self.decode_access_token(token)
        user_id_str = payload.get("sub")

        if user_id_str is None:
            raise InvalidTokenError()

        # Convert sub back to int (it's stored as string in JWT per RFC 7519)
        try:
            user_id = int(user_id_str)
        except (ValueError, TypeError):
            raise InvalidTokenError()

        user = db.query(User).filter(User.id == user_id).first()
        if user is None:
            raise UserNotFoundError()

        if not user.is_active:
            raise InactiveUserError()

        # Cache the user ID
        _token_cache[token_hash] = user.id

        return user

    def create_or_update_user(
        self,
        db: Session,
        google_id: str,
        email: str,
        name: str,
        picture_url: str
    ) -> User:
        """
        Create or update a user from Google OAuth data

        Args:
            db: Database session
            google_id: Google user ID
            email: User email
            name: User name
            picture_url: User picture URL

        Returns:
            User object
        """
        # Try to find existing user by google_id or email
        user = db.query(User).filter(
            (User.google_id == google_id) | (User.email == email)
        ).first()

        if user:
            # Update existing user
            user.google_id = google_id
            user.email = email
            user.name = name
            user.picture_url = picture_url
            user.is_verified = True
            user.last_login_at = datetime.now(timezone.utc)
            user.updated_at = datetime.now(timezone.utc)
            logger.info(f"Updated existing user: {email}")
        else:
            # Create new user
            user = User(
                google_id=google_id,
                email=email,
                name=name,
                picture_url=picture_url,
                is_active=True,
                is_verified=True,
                last_login_at=datetime.now(timezone.utc),
            )
            db.add(user)
            logger.info(f"Created new user: {email}")

        db.commit()
        db.refresh(user)
        return user

    async def exchange_code_for_user_info(self, code: str) -> dict:
        """
        Exchange authorization code for user info from Google

        Uses authlib-style OAuth flow with manual HTTP requests

        Args:
            code: Authorization code from Google

        Returns:
            User info dict with google_id, email, name, picture_url

        Raises:
            OAuthNotConfiguredError: If OAuth credentials not configured
            OAuthCodeExchangeError: If code exchange fails
            OAuthUserInfoError: If user info fetch fails
        """
        # Check OAuth configuration
        if not settings.GOOGLE_CLIENT_ID or not settings.validate_oauth_config():
            raise OAuthNotConfiguredError()

        client_secret = (
            settings.GOOGLE_CLIENT_SECRET.get_secret_value()
            if hasattr(settings.GOOGLE_CLIENT_SECRET, "get_secret_value")
            else str(settings.GOOGLE_CLIENT_SECRET)
        )

        # Get HTTP client
        if self._http_client:
            client = self._http_client
            # Exchange code for token
            token_response = await client.post(
                "https://oauth2.googleapis.com/token",
                data={
                    "code": code,
                    "client_id": settings.GOOGLE_CLIENT_ID,
                    "client_secret": client_secret,
                    "redirect_uri": settings.GOOGLE_REDIRECT_URI,
                    "grant_type": "authorization_code",
                },
                timeout=10.0,
            )

            if token_response.status_code != 200:
                logger.error(f"Token exchange failed: {token_response.text}")
                raise OAuthCodeExchangeError(details=token_response.text)

            token_data = token_response.json()
            access_token = token_data.get("access_token")

            # Get user info
            userinfo_response = await client.get(
                "https://www.googleapis.com/oauth2/v2/userinfo",
                headers={"Authorization": f"Bearer {access_token}"},
                timeout=10.0,
            )

            if userinfo_response.status_code != 200:
                logger.error(f"User info fetch failed: {userinfo_response.text}")
                raise OAuthUserInfoError()

            user_info = userinfo_response.json()
        else:
            # Use httpx.AsyncClient context manager
            async with httpx.AsyncClient() as client:
                # Exchange code for token
                token_response = await client.post(
                    "https://oauth2.googleapis.com/token",
                    data={
                        "code": code,
                        "client_id": settings.GOOGLE_CLIENT_ID,
                        "client_secret": client_secret,
                        "redirect_uri": settings.GOOGLE_REDIRECT_URI,
                        "grant_type": "authorization_code",
                    },
                    timeout=10.0,
                )

                if token_response.status_code != 200:
                    logger.error(f"Token exchange failed: {token_response.text}")
                    raise OAuthCodeExchangeError(details=token_response.text)

                token_data = token_response.json()
                access_token = token_data.get("access_token")

                # Get user info
                userinfo_response = await client.get(
                    "https://www.googleapis.com/oauth2/v2/userinfo",
                    headers={"Authorization": f"Bearer {access_token}"},
                    timeout=10.0,
                )

                if userinfo_response.status_code != 200:
                    logger.error(f"User info fetch failed: {userinfo_response.text}")
                    raise OAuthUserInfoError()

                user_info = userinfo_response.json()

        return {
            "google_id": user_info.get("id"),
            "email": user_info.get("email"),
            "name": user_info.get("name"),
            "picture_url": user_info.get("picture"),
        }

    def renew_token(self, current_token: str, db: Session) -> str:
        """
        Renew an access token

        Args:
            current_token: Current valid JWT token
            db: Database session

        Returns:
            New JWT token

        Raises:
            TokenExpiredError: If current token is expired
            InvalidTokenError: If current token is invalid
            UserNotFoundError: If user not found
            InactiveUserError: If user is inactive
        """
        # Validate current token and get user
        user = self.get_user_from_token(current_token, db)

        # Create new token
        access_token = self.create_access_token(data={"sub": user.id})

        logger.info(f"Renewed token for user: {user.email}")
        return access_token

    def invalidate_token(self, token: str):
        """
        Invalidate a specific token from the cache

        Args:
            token: JWT token to invalidate
        """
        token_hash = hash(token)
        if token_hash in _token_cache:
            del _token_cache[token_hash]
            logger.info("Token invalidated from cache")

    def invalidate_user_tokens(self, user_id: int):
        """
        Invalidate all cached tokens for a specific user

        Args:
            user_id: User ID whose tokens should be invalidated
        """
        # Find and remove all cache entries for this user
        keys_to_remove = [
            key for key, cached_user_id in _token_cache.items()
            if cached_user_id == user_id
        ]
        for key in keys_to_remove:
            del _token_cache[key]
        if keys_to_remove:
            logger.info(f"Invalidated {len(keys_to_remove)} cached tokens for user {user_id}")

    @staticmethod
    def clear_token_cache():
        """Clear the entire token cache (useful for testing)"""
        _token_cache.clear()


# Create singleton auth service
auth_service = AuthService()
