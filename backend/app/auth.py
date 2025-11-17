"""FastAPI-Users authentication configuration"""
import logging
from typing import Optional
from fastapi import Depends, Request
from fastapi_users import BaseUserManager, FastAPIUsers, IntegerIDMixin
from fastapi_users.authentication import (
    AuthenticationBackend,
    CookieTransport,
    JWTStrategy,
)
from fastapi_users.db import SQLAlchemyUserDatabase
from sqlalchemy.ext.asyncio import AsyncSession

from app.config import settings
from app.models.user import User
from app.database import AsyncSessionLocal

logger = logging.getLogger(__name__)


class UserManager(IntegerIDMixin, BaseUserManager[User, int]):
    """User manager for fastapi-users with custom logic"""

    reset_password_token_secret = settings.JWT_SECRET_KEY.get_secret_value()
    verification_token_secret = settings.JWT_SECRET_KEY.get_secret_value()

    async def on_after_register(self, user: User, request: Optional[Request] = None):
        """Called after user registration"""
        logger.info(f"User {user.id} ({user.email}) has registered")

    async def on_after_forgot_password(
        self, user: User, token: str, request: Optional[Request] = None
    ):
        """Called after forgot password request"""
        logger.info(f"User {user.id} has forgot their password. Reset token: {token}")

    async def on_after_request_verify(
        self, user: User, token: str, request: Optional[Request] = None
    ):
        """Called after verification email request"""
        logger.info(f"Verification requested for user {user.id}. Verification token: {token}")


async def get_async_session():
    """Dependency to get async database session"""
    async with AsyncSessionLocal() as session:
        yield session


async def get_user_db(session: AsyncSession = Depends(get_async_session)):
    """Dependency to get the user database adapter"""
    yield SQLAlchemyUserDatabase(session, User)


async def get_user_manager(user_db: SQLAlchemyUserDatabase = Depends(get_user_db)):
    """Dependency to get the user manager"""
    yield UserManager(user_db)


# Cookie transport configuration
cookie_transport = CookieTransport(
    cookie_name="opengov_auth",
    cookie_max_age=settings.JWT_ACCESS_TOKEN_EXPIRE_MINUTES * 60,  # Convert to seconds
    cookie_httponly=True,
    cookie_secure=settings.COOKIE_SECURE,  # Auto-configured from environment
    cookie_samesite="lax",
)


def get_jwt_strategy() -> JWTStrategy:
    """Get JWT strategy for authentication"""
    return JWTStrategy(
        secret=settings.JWT_SECRET_KEY.get_secret_value(),
        lifetime_seconds=settings.JWT_ACCESS_TOKEN_EXPIRE_MINUTES * 60,
    )


# Authentication backend
auth_backend = AuthenticationBackend(
    name="cookie",
    transport=cookie_transport,
    get_strategy=get_jwt_strategy,
)

# FastAPI Users instance
fastapi_users = FastAPIUsers[User, int](
    get_user_manager,
    [auth_backend],
)

# Dependency for current user (requires authentication)
current_user = fastapi_users.current_user()

# Dependency for current active user (requires authentication and active status)
current_active_user = fastapi_users.current_user(active=True)

# Dependency for current superuser (requires authentication and superuser status)
current_superuser = fastapi_users.current_user(active=True, superuser=True)

# Dependency for optional user (no authentication required)
optional_current_user = fastapi_users.current_user(optional=True)
