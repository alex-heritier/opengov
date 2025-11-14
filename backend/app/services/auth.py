"""Authentication service for Google OAuth and JWT token management"""
import logging
from datetime import datetime, timedelta
from typing import Optional

import httpx
from authlib.integrations.starlette_client import OAuth
from fastapi import HTTPException, status
from jose import JWTError, jwt
from sqlalchemy.orm import Session

from app.config import settings
from app.models.user import User

logger = logging.getLogger(__name__)

# OAuth configuration
oauth = OAuth()
oauth.register(
    name="google",
    client_id=settings.GOOGLE_CLIENT_ID,
    client_secret=settings.GOOGLE_CLIENT_SECRET,
    server_metadata_url="https://accounts.google.com/.well-known/openid-configuration",
    client_kwargs={"scope": "openid email profile"},
)


def create_access_token(data: dict, expires_delta: Optional[timedelta] = None) -> str:
    """
    Create a JWT access token

    Args:
        data: Data to encode in the token
        expires_delta: Optional expiration time delta

    Returns:
        Encoded JWT token
    """
    to_encode = data.copy()
    if expires_delta:
        expire = datetime.utcnow() + expires_delta
    else:
        expire = datetime.utcnow() + timedelta(
            minutes=settings.JWT_ACCESS_TOKEN_EXPIRE_MINUTES
        )
    to_encode.update({"exp": expire, "iat": datetime.utcnow()})
    encoded_jwt = jwt.encode(
        to_encode, settings.JWT_SECRET_KEY, algorithm=settings.JWT_ALGORITHM
    )
    return encoded_jwt


def decode_access_token(token: str) -> dict:
    """
    Decode and validate a JWT access token

    Args:
        token: JWT token to decode

    Returns:
        Decoded token payload

    Raises:
        HTTPException: If token is invalid or expired
    """
    try:
        payload = jwt.decode(
            token, settings.JWT_SECRET_KEY, algorithms=[settings.JWT_ALGORITHM]
        )
        return payload
    except JWTError as e:
        logger.warning(f"Invalid token: {e}")
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Could not validate credentials",
            headers={"WWW-Authenticate": "Bearer"},
        )


def get_user_from_token(token: str, db: Session) -> User:
    """
    Get user from JWT token

    Args:
        token: JWT token
        db: Database session

    Returns:
        User object

    Raises:
        HTTPException: If token is invalid or user not found
    """
    payload = decode_access_token(token)
    user_id: int = payload.get("sub")
    if user_id is None:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Could not validate credentials",
            headers={"WWW-Authenticate": "Bearer"},
        )

    user = db.query(User).filter(User.id == user_id).first()
    if user is None:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="User not found",
            headers={"WWW-Authenticate": "Bearer"},
        )

    if not user.is_active:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="Inactive user",
        )

    return user


def create_or_update_user(db: Session, google_id: str, email: str, name: str, picture_url: str) -> User:
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
        user.last_login_at = datetime.utcnow()
        user.updated_at = datetime.utcnow()
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
            last_login_at=datetime.utcnow(),
        )
        db.add(user)
        logger.info(f"Created new user: {email}")

    db.commit()
    db.refresh(user)
    return user


async def exchange_code_for_user_info(code: str) -> dict:
    """
    Exchange authorization code for user info from Google

    Args:
        code: Authorization code from Google

    Returns:
        User info dict with google_id, email, name, picture_url

    Raises:
        HTTPException: If exchange fails
    """
    try:
        # Exchange code for token
        async with httpx.AsyncClient() as client:
            token_response = await client.post(
                "https://oauth2.googleapis.com/token",
                data={
                    "code": code,
                    "client_id": settings.GOOGLE_CLIENT_ID,
                    "client_secret": settings.GOOGLE_CLIENT_SECRET,
                    "redirect_uri": settings.GOOGLE_REDIRECT_URI,
                    "grant_type": "authorization_code",
                },
                timeout=10.0,
            )

            if token_response.status_code != 200:
                logger.error(f"Token exchange failed: {token_response.text}")
                raise HTTPException(
                    status_code=status.HTTP_400_BAD_REQUEST,
                    detail="Failed to exchange authorization code",
                )

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
                raise HTTPException(
                    status_code=status.HTTP_400_BAD_REQUEST,
                    detail="Failed to fetch user information",
                )

            user_info = userinfo_response.json()

            return {
                "google_id": user_info.get("id"),
                "email": user_info.get("email"),
                "name": user_info.get("name"),
                "picture_url": user_info.get("picture"),
            }
    except httpx.RequestError as e:
        logger.error(f"HTTP request failed during OAuth: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to communicate with Google OAuth",
        )


def renew_token(current_token: str, db: Session) -> str:
    """
    Renew an access token

    Args:
        current_token: Current valid JWT token
        db: Database session

    Returns:
        New JWT token

    Raises:
        HTTPException: If current token is invalid
    """
    # Validate current token and get user
    user = get_user_from_token(current_token, db)

    # Create new token
    access_token = create_access_token(data={"sub": user.id})

    logger.info(f"Renewed token for user: {user.email}")
    return access_token
