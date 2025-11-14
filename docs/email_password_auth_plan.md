# Email + Password Authentication Implementation Plan

## Overview
This document outlines the implementation plan for adding email/password authentication to the OpenGov platform (Phase 2). The authentication system will support user registration, login, password reset, email verification, session management, and protected routes.

## Table of Contents
1. [Architecture Overview](#architecture-overview)
2. [Dependencies](#dependencies)
3. [Database Schema](#database-schema)
4. [Backend Implementation](#backend-implementation)
5. [Frontend Integration](#frontend-integration)
6. [Security Considerations](#security-considerations)
7. [Testing Strategy](#testing-strategy)
8. [Migration Path](#migration-path)

---

## Architecture Overview

### Registration Flow
```
1. User fills registration form (email, password, name)
2. Frontend sends POST /api/auth/register
3. Backend validates input (email format, password strength)
4. Backend hashes password with bcrypt
5. Backend creates user in database (is_verified=False)
6. Backend generates verification token
7. Backend sends verification email (optional in dev)
8. User clicks link in email → verifies account
9. User can now log in
```

### Login Flow
```
1. User fills login form (email, password)
2. Frontend sends POST /api/auth/login
3. Backend finds user by email
4. Backend verifies password hash
5. Backend generates JWT access token + refresh token
6. Backend returns tokens to frontend
7. Frontend stores tokens in localStorage/auth store
8. Subsequent requests include JWT in Authorization header
```

### Password Reset Flow
```
1. User clicks "Forgot Password"
2. User enters email → POST /api/auth/forgot-password
3. Backend generates reset token (expires in 1 hour)
4. Backend sends reset email with link
5. User clicks link → redirected to reset password page
6. User enters new password → POST /api/auth/reset-password
7. Backend validates token, updates password
8. User can log in with new password
```

### Token Strategy
- **Access Token**: Short-lived JWT (15 minutes), includes user ID and email
- **Refresh Token**: Long-lived (7 days), stored in database, used to refresh access tokens
- **Verification Token**: One-time use, expires in 24 hours, for email verification
- **Reset Token**: One-time use, expires in 1 hour, for password reset

---

## Dependencies

### Backend (`requirements.txt`)
```txt
# Add these packages
python-jose[cryptography]==3.3.0  # JWT encoding/decoding
passlib[bcrypt]==1.7.4            # Password hashing with bcrypt
python-multipart==0.0.6           # Form data parsing
email-validator==2.1.0            # Email validation

# Optional: Email sending (for verification and password reset)
fastapi-mail==1.4.1               # Email sending library
```

### Environment Variables (`.env`)
```bash
# JWT
JWT_SECRET_KEY=your-random-secret-key-min-32-chars
JWT_ALGORITHM=HS256
JWT_ACCESS_TOKEN_EXPIRE_MINUTES=15
JWT_REFRESH_TOKEN_EXPIRE_DAYS=7

# Password
PASSWORD_MIN_LENGTH=8
PASSWORD_REQUIRE_UPPERCASE=true
PASSWORD_REQUIRE_LOWERCASE=true
PASSWORD_REQUIRE_DIGIT=true
PASSWORD_REQUIRE_SPECIAL=true

# Email (optional for dev, required for production)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_FROM_EMAIL=noreply@opengov.app
SMTP_FROM_NAME=OpenGov

# Frontend URL (for email links)
FRONTEND_URL=http://localhost:5173

# Features
REQUIRE_EMAIL_VERIFICATION=false  # Set to true in production
```

---

## Database Schema

### 1. User Model
```python
# backend/app/models/user.py

from datetime import datetime, timezone
from sqlalchemy import Column, Integer, String, DateTime, Boolean
from app.database import Base

class User(Base):
    __tablename__ = "users"

    id = Column(Integer, primary_key=True, index=True)
    email = Column(String(255), unique=True, nullable=False, index=True)
    password_hash = Column(String(255), nullable=False)  # Bcrypt hash
    name = Column(String(255), nullable=True)

    # Account status
    is_active = Column(Boolean, default=True, nullable=False)
    is_verified = Column(Boolean, default=False, nullable=False)
    is_admin = Column(Boolean, default=False, nullable=False)

    # Timestamps
    created_at = Column(DateTime, default=lambda: datetime.now(timezone.utc), nullable=False)
    updated_at = Column(DateTime, default=lambda: datetime.now(timezone.utc),
                       onupdate=lambda: datetime.now(timezone.utc), nullable=False)
    last_login_at = Column(DateTime, nullable=True)

    # Email verification
    verification_token = Column(String(255), nullable=True, unique=True, index=True)
    verification_token_expires = Column(DateTime, nullable=True)

    # Password reset
    reset_token = Column(String(255), nullable=True, unique=True, index=True)
    reset_token_expires = Column(DateTime, nullable=True)
```

### 2. RefreshToken Model
```python
# backend/app/models/refresh_token.py

from datetime import datetime, timezone
from sqlalchemy import Column, Integer, String, DateTime, ForeignKey, Boolean
from sqlalchemy.orm import relationship
from app.database import Base

class RefreshToken(Base):
    __tablename__ = "refresh_tokens"

    id = Column(Integer, primary_key=True, index=True)
    user_id = Column(Integer, ForeignKey("users.id"), nullable=False, index=True)
    token = Column(String(500), unique=True, nullable=False, index=True)

    # Token metadata
    expires_at = Column(DateTime, nullable=False)
    is_revoked = Column(Boolean, default=False, nullable=False)

    # Device/session tracking
    user_agent = Column(String(500), nullable=True)
    ip_address = Column(String(45), nullable=True)  # IPv6 compatible

    created_at = Column(DateTime, default=lambda: datetime.now(timezone.utc), nullable=False)

    # Relationship
    user = relationship("User", backref="refresh_tokens")
```

### 3. Database Migration
```python
# backend/migrations/versions/003_add_users_and_auth.py

"""Add users and authentication tables

Revision ID: 003_add_users_and_auth
Revises: 002_add_agencies_table
Create Date: 2025-11-14
"""

def upgrade() -> None:
    # Create users table
    op.create_table('users',
        sa.Column('id', sa.Integer(), nullable=False),
        sa.Column('email', sa.String(length=255), nullable=False),
        sa.Column('password_hash', sa.String(length=255), nullable=False),
        sa.Column('name', sa.String(length=255), nullable=True),
        sa.Column('is_active', sa.Boolean(), nullable=False),
        sa.Column('is_verified', sa.Boolean(), nullable=False),
        sa.Column('is_admin', sa.Boolean(), nullable=False),
        sa.Column('created_at', sa.DateTime(), nullable=False),
        sa.Column('updated_at', sa.DateTime(), nullable=False),
        sa.Column('last_login_at', sa.DateTime(), nullable=True),
        sa.Column('verification_token', sa.String(length=255), nullable=True),
        sa.Column('verification_token_expires', sa.DateTime(), nullable=True),
        sa.Column('reset_token', sa.String(length=255), nullable=True),
        sa.Column('reset_token_expires', sa.DateTime(), nullable=True),
        sa.PrimaryKeyConstraint('id'),
        sa.UniqueConstraint('email'),
        sa.UniqueConstraint('verification_token'),
        sa.UniqueConstraint('reset_token')
    )
    op.create_index('ix_users_email', 'users', ['email'])
    op.create_index('ix_users_verification_token', 'users', ['verification_token'])
    op.create_index('ix_users_reset_token', 'users', ['reset_token'])

    # Create refresh_tokens table
    op.create_table('refresh_tokens',
        sa.Column('id', sa.Integer(), nullable=False),
        sa.Column('user_id', sa.Integer(), nullable=False),
        sa.Column('token', sa.String(length=500), nullable=False),
        sa.Column('expires_at', sa.DateTime(), nullable=False),
        sa.Column('is_revoked', sa.Boolean(), nullable=False),
        sa.Column('user_agent', sa.String(length=500), nullable=True),
        sa.Column('ip_address', sa.String(length=45), nullable=True),
        sa.Column('created_at', sa.DateTime(), nullable=False),
        sa.ForeignKeyConstraint(['user_id'], ['users.id']),
        sa.PrimaryKeyConstraint('id'),
        sa.UniqueConstraint('token')
    )
    op.create_index('ix_refresh_tokens_token', 'refresh_tokens', ['token'])
    op.create_index('ix_refresh_tokens_user_id', 'refresh_tokens', ['user_id'])

def downgrade() -> None:
    op.drop_table('refresh_tokens')
    op.drop_table('users')
```

---

## Backend Implementation

### 1. Configuration Updates
```python
# backend/app/config.py - Add these settings

class Settings:
    # ... existing settings ...

    # JWT Settings
    JWT_SECRET_KEY: str = os.getenv("JWT_SECRET_KEY", "")
    JWT_ALGORITHM: str = os.getenv("JWT_ALGORITHM", "HS256")
    JWT_ACCESS_TOKEN_EXPIRE_MINUTES: int = int(
        os.getenv("JWT_ACCESS_TOKEN_EXPIRE_MINUTES", "15")
    )
    JWT_REFRESH_TOKEN_EXPIRE_DAYS: int = int(
        os.getenv("JWT_REFRESH_TOKEN_EXPIRE_DAYS", "7")
    )

    # Password Requirements
    PASSWORD_MIN_LENGTH: int = int(os.getenv("PASSWORD_MIN_LENGTH", "8"))
    PASSWORD_REQUIRE_UPPERCASE: bool = os.getenv("PASSWORD_REQUIRE_UPPERCASE", "true").lower() == "true"
    PASSWORD_REQUIRE_LOWERCASE: bool = os.getenv("PASSWORD_REQUIRE_LOWERCASE", "true").lower() == "true"
    PASSWORD_REQUIRE_DIGIT: bool = os.getenv("PASSWORD_REQUIRE_DIGIT", "true").lower() == "true"
    PASSWORD_REQUIRE_SPECIAL: bool = os.getenv("PASSWORD_REQUIRE_SPECIAL", "true").lower() == "true"

    # Email Settings (optional in dev)
    SMTP_HOST: str = os.getenv("SMTP_HOST", "")
    SMTP_PORT: int = int(os.getenv("SMTP_PORT", "587"))
    SMTP_USERNAME: str = os.getenv("SMTP_USERNAME", "")
    SMTP_PASSWORD: str = os.getenv("SMTP_PASSWORD", "")
    SMTP_FROM_EMAIL: str = os.getenv("SMTP_FROM_EMAIL", "noreply@opengov.app")
    SMTP_FROM_NAME: str = os.getenv("SMTP_FROM_NAME", "OpenGov")

    # Feature Flags
    REQUIRE_EMAIL_VERIFICATION: bool = os.getenv("REQUIRE_EMAIL_VERIFICATION", "false").lower() == "true"

    # Frontend URL
    FRONTEND_URL: str = os.getenv("FRONTEND_URL", "http://localhost:5173")

    def validate(self):
        # ... existing validation ...

        # Validate auth settings
        if not self.DEBUG:
            if not self.JWT_SECRET_KEY or len(self.JWT_SECRET_KEY) < 32:
                raise ValueError("JWT_SECRET_KEY must be at least 32 characters in production")
            if self.REQUIRE_EMAIL_VERIFICATION and not self.SMTP_HOST:
                raise ValueError("SMTP settings required when email verification is enabled")
```

### 2. Password Utilities
```python
# backend/app/services/password.py

from passlib.context import CryptContext
from app.config import settings
import re

pwd_context = CryptContext(schemes=["bcrypt"], deprecated="auto")

def hash_password(password: str) -> str:
    """Hash a password using bcrypt"""
    return pwd_context.hash(password)

def verify_password(plain_password: str, hashed_password: str) -> bool:
    """Verify a password against a hash"""
    return pwd_context.verify(plain_password, hashed_password)

def validate_password_strength(password: str) -> tuple[bool, str]:
    """
    Validate password meets strength requirements

    Returns:
        (is_valid, error_message)
    """
    if len(password) < settings.PASSWORD_MIN_LENGTH:
        return False, f"Password must be at least {settings.PASSWORD_MIN_LENGTH} characters"

    if settings.PASSWORD_REQUIRE_UPPERCASE and not re.search(r'[A-Z]', password):
        return False, "Password must contain at least one uppercase letter"

    if settings.PASSWORD_REQUIRE_LOWERCASE and not re.search(r'[a-z]', password):
        return False, "Password must contain at least one lowercase letter"

    if settings.PASSWORD_REQUIRE_DIGIT and not re.search(r'\d', password):
        return False, "Password must contain at least one digit"

    if settings.PASSWORD_REQUIRE_SPECIAL and not re.search(r'[!@#$%^&*(),.?":{}|<>]', password):
        return False, "Password must contain at least one special character"

    return True, ""
```

### 3. JWT & Token Utilities
```python
# backend/app/services/auth.py

from datetime import datetime, timedelta, timezone
from typing import Optional
from jose import JWTError, jwt
from sqlalchemy.orm import Session
from app.config import settings
from app.models import User, RefreshToken
import secrets

def create_access_token(data: dict, expires_delta: Optional[timedelta] = None) -> str:
    """Create JWT access token"""
    to_encode = data.copy()

    if expires_delta:
        expire = datetime.now(timezone.utc) + expires_delta
    else:
        expire = datetime.now(timezone.utc) + timedelta(
            minutes=settings.JWT_ACCESS_TOKEN_EXPIRE_MINUTES
        )

    to_encode.update({"exp": expire, "type": "access"})
    encoded_jwt = jwt.encode(to_encode, settings.JWT_SECRET_KEY, algorithm=settings.JWT_ALGORITHM)
    return encoded_jwt

def create_refresh_token(
    db: Session,
    user_id: int,
    user_agent: Optional[str] = None,
    ip_address: Optional[str] = None
) -> str:
    """Create and store refresh token"""
    token = secrets.token_urlsafe(32)
    expires_at = datetime.now(timezone.utc) + timedelta(days=settings.JWT_REFRESH_TOKEN_EXPIRE_DAYS)

    refresh_token = RefreshToken(
        user_id=user_id,
        token=token,
        expires_at=expires_at,
        user_agent=user_agent,
        ip_address=ip_address
    )

    db.add(refresh_token)
    db.commit()

    return token

def create_verification_token() -> str:
    """Create email verification token"""
    return secrets.token_urlsafe(32)

def create_reset_token() -> str:
    """Create password reset token"""
    return secrets.token_urlsafe(32)

def verify_access_token(token: str) -> Optional[dict]:
    """Verify and decode JWT access token"""
    try:
        payload = jwt.decode(token, settings.JWT_SECRET_KEY, algorithms=[settings.JWT_ALGORITHM])
        if payload.get("type") != "access":
            return None
        return payload
    except JWTError:
        return None

def verify_refresh_token(db: Session, token: str) -> Optional[User]:
    """Verify refresh token and return associated user"""
    refresh_token = db.query(RefreshToken).filter(
        RefreshToken.token == token,
        RefreshToken.is_revoked == False,
        RefreshToken.expires_at > datetime.now(timezone.utc)
    ).first()

    if not refresh_token:
        return None

    return refresh_token.user

def revoke_refresh_token(db: Session, token: str):
    """Revoke a refresh token"""
    refresh_token = db.query(RefreshToken).filter(RefreshToken.token == token).first()
    if refresh_token:
        refresh_token.is_revoked = True
        db.commit()

def revoke_all_user_tokens(db: Session, user_id: int):
    """Revoke all refresh tokens for a user (logout from all devices)"""
    db.query(RefreshToken).filter(RefreshToken.user_id == user_id).update(
        {"is_revoked": True}
    )
    db.commit()
```

### 4. Email Service
```python
# backend/app/services/email.py

import logging
from fastapi_mail import FastMail, MessageSchema, ConnectionConfig
from app.config import settings

logger = logging.getLogger(__name__)

# Email configuration
conf = ConnectionConfig(
    MAIL_USERNAME=settings.SMTP_USERNAME,
    MAIL_PASSWORD=settings.SMTP_PASSWORD,
    MAIL_FROM=settings.SMTP_FROM_EMAIL,
    MAIL_FROM_NAME=settings.SMTP_FROM_NAME,
    MAIL_PORT=settings.SMTP_PORT,
    MAIL_SERVER=settings.SMTP_HOST,
    MAIL_STARTTLS=True,
    MAIL_SSL_TLS=False,
    USE_CREDENTIALS=True,
)

fm = FastMail(conf) if settings.SMTP_HOST else None

async def send_verification_email(email: str, token: str):
    """Send email verification link"""
    if not fm:
        logger.warning(f"Email not configured. Verification link: {settings.FRONTEND_URL}/verify-email?token={token}")
        return

    verification_link = f"{settings.FRONTEND_URL}/verify-email?token={token}"

    message = MessageSchema(
        subject="Verify your OpenGov account",
        recipients=[email],
        body=f"""
        <html>
            <body>
                <h2>Welcome to OpenGov!</h2>
                <p>Please click the link below to verify your email address:</p>
                <a href="{verification_link}">Verify Email</a>
                <p>This link will expire in 24 hours.</p>
                <p>If you didn't create an account, you can safely ignore this email.</p>
            </body>
        </html>
        """,
        subtype="html"
    )

    await fm.send_message(message)
    logger.info(f"Verification email sent to {email}")

async def send_password_reset_email(email: str, token: str):
    """Send password reset link"""
    if not fm:
        logger.warning(f"Email not configured. Reset link: {settings.FRONTEND_URL}/reset-password?token={token}")
        return

    reset_link = f"{settings.FRONTEND_URL}/reset-password?token={token}"

    message = MessageSchema(
        subject="Reset your OpenGov password",
        recipients=[email],
        body=f"""
        <html>
            <body>
                <h2>Password Reset Request</h2>
                <p>Click the link below to reset your password:</p>
                <a href="{reset_link}">Reset Password</a>
                <p>This link will expire in 1 hour.</p>
                <p>If you didn't request a password reset, you can safely ignore this email.</p>
            </body>
        </html>
        """,
        subtype="html"
    )

    await fm.send_message(message)
    logger.info(f"Password reset email sent to {email}")
```

### 5. Authentication Dependency
```python
# backend/app/dependencies/auth.py

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
    db: Session = Depends(get_db)
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
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Invalid token payload"
        )

    user = db.query(User).filter(User.id == int(user_id)).first()
    if not user:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="User not found"
        )

    if not user.is_active:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="User account is inactive"
        )

    return user

async def get_current_verified_user(
    current_user: User = Depends(get_current_user)
) -> User:
    """Get current user and ensure email is verified"""
    if not current_user.is_verified:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="Email not verified. Please verify your email to access this resource."
        )
    return current_user

async def get_current_admin_user(
    current_user: User = Depends(get_current_user)
) -> User:
    """Get current user and ensure they are an admin"""
    if not current_user.is_admin:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="Admin access required"
        )
    return current_user

async def get_current_user_optional(
    credentials: Optional[HTTPAuthorizationCredentials] = Depends(security),
    db: Session = Depends(get_db)
) -> Optional[User]:
    """Get current user if authenticated, otherwise return None"""
    if not credentials:
        return None

    try:
        return await get_current_user(credentials, db)
    except HTTPException:
        return None
```

### 6. Auth Router
```python
# backend/app/routers/auth.py

import logging
from datetime import datetime, timedelta, timezone
from fastapi import APIRouter, Depends, HTTPException, Request, status
from sqlalchemy.orm import Session
from email_validator import validate_email, EmailNotValidError
from app.routers.common import get_db, limiter
from app.models import User
from app.schemas.auth import (
    RegisterRequest, LoginRequest, TokenResponse, RefreshTokenRequest,
    ForgotPasswordRequest, ResetPasswordRequest, VerifyEmailRequest
)
from app.services.auth import (
    create_access_token, create_refresh_token, create_verification_token,
    create_reset_token, verify_refresh_token, revoke_refresh_token,
    revoke_all_user_tokens
)
from app.services.password import hash_password, verify_password, validate_password_strength
from app.services.email import send_verification_email, send_password_reset_email
from app.dependencies.auth import get_current_user
from app.config import settings

logger = logging.getLogger(__name__)
router = APIRouter(prefix="/api/auth", tags=["auth"])


@router.post("/register", response_model=TokenResponse)
@limiter.limit("5/hour")  # Strict limit to prevent spam registrations
async def register(
    request: Request,
    register_data: RegisterRequest,
    db: Session = Depends(get_db)
):
    """Register a new user account"""
    # Validate email
    try:
        email_info = validate_email(register_data.email, check_deliverability=False)
        email = email_info.normalized
    except EmailNotValidError as e:
        raise HTTPException(status_code=400, detail=str(e))

    # Validate password strength
    is_valid, error_msg = validate_password_strength(register_data.password)
    if not is_valid:
        raise HTTPException(status_code=400, detail=error_msg)

    # Check if user already exists
    existing_user = db.query(User).filter(User.email == email).first()
    if existing_user:
        raise HTTPException(status_code=400, detail="Email already registered")

    # Create verification token
    verification_token = create_verification_token()
    verification_expires = datetime.now(timezone.utc) + timedelta(hours=24)

    # Create user
    user = User(
        email=email,
        password_hash=hash_password(register_data.password),
        name=register_data.name,
        is_verified=not settings.REQUIRE_EMAIL_VERIFICATION,  # Auto-verify if not required
        verification_token=verification_token if settings.REQUIRE_EMAIL_VERIFICATION else None,
        verification_token_expires=verification_expires if settings.REQUIRE_EMAIL_VERIFICATION else None
    )

    db.add(user)
    db.commit()
    db.refresh(user)

    # Send verification email if required
    if settings.REQUIRE_EMAIL_VERIFICATION:
        await send_verification_email(email, verification_token)
        logger.info(f"User registered: {email} (verification required)")
    else:
        logger.info(f"User registered: {email} (auto-verified)")

    # Generate tokens
    access_token = create_access_token(data={"sub": str(user.id), "email": user.email})
    refresh_token = create_refresh_token(
        db=db,
        user_id=user.id,
        user_agent=request.headers.get('user-agent'),
        ip_address=request.client.host
    )

    return TokenResponse(
        access_token=access_token,
        refresh_token=refresh_token,
        token_type="bearer",
        expires_in=settings.JWT_ACCESS_TOKEN_EXPIRE_MINUTES * 60
    )


@router.post("/login", response_model=TokenResponse)
@limiter.limit("10/minute")
async def login(
    request: Request,
    login_data: LoginRequest,
    db: Session = Depends(get_db)
):
    """Login with email and password"""
    # Find user
    user = db.query(User).filter(User.email == login_data.email.lower()).first()

    # Verify credentials
    if not user or not verify_password(login_data.password, user.password_hash):
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Incorrect email or password"
        )

    # Check if account is active
    if not user.is_active:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="Account is inactive. Please contact support."
        )

    # Update last login
    user.last_login_at = datetime.now(timezone.utc)
    db.commit()

    # Generate tokens
    access_token = create_access_token(data={"sub": str(user.id), "email": user.email})
    refresh_token = create_refresh_token(
        db=db,
        user_id=user.id,
        user_agent=request.headers.get('user-agent'),
        ip_address=request.client.host
    )

    logger.info(f"User logged in: {user.email}")

    return TokenResponse(
        access_token=access_token,
        refresh_token=refresh_token,
        token_type="bearer",
        expires_in=settings.JWT_ACCESS_TOKEN_EXPIRE_MINUTES * 60
    )


@router.post("/verify-email")
@limiter.limit("10/minute")
async def verify_email(
    request: Request,
    verify_data: VerifyEmailRequest,
    db: Session = Depends(get_db)
):
    """Verify email address with token"""
    user = db.query(User).filter(
        User.verification_token == verify_data.token,
        User.verification_token_expires > datetime.now(timezone.utc)
    ).first()

    if not user:
        raise HTTPException(status_code=400, detail="Invalid or expired verification token")

    # Mark as verified
    user.is_verified = True
    user.verification_token = None
    user.verification_token_expires = None
    db.commit()

    logger.info(f"Email verified: {user.email}")

    return {"message": "Email verified successfully"}


@router.post("/resend-verification")
@limiter.limit("3/hour")
async def resend_verification(
    request: Request,
    current_user: User = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """Resend verification email"""
    if current_user.is_verified:
        raise HTTPException(status_code=400, detail="Email already verified")

    # Generate new token
    verification_token = create_verification_token()
    verification_expires = datetime.now(timezone.utc) + timedelta(hours=24)

    current_user.verification_token = verification_token
    current_user.verification_token_expires = verification_expires
    db.commit()

    # Send email
    await send_verification_email(current_user.email, verification_token)

    return {"message": "Verification email sent"}


@router.post("/forgot-password")
@limiter.limit("3/hour")
async def forgot_password(
    request: Request,
    forgot_data: ForgotPasswordRequest,
    db: Session = Depends(get_db)
):
    """Request password reset email"""
    user = db.query(User).filter(User.email == forgot_data.email.lower()).first()

    # Always return success to prevent email enumeration
    if not user:
        logger.warning(f"Password reset requested for non-existent email: {forgot_data.email}")
        return {"message": "If that email exists, a password reset link has been sent"}

    # Generate reset token
    reset_token = create_reset_token()
    reset_expires = datetime.now(timezone.utc) + timedelta(hours=1)

    user.reset_token = reset_token
    user.reset_token_expires = reset_expires
    db.commit()

    # Send email
    await send_password_reset_email(user.email, reset_token)

    logger.info(f"Password reset requested: {user.email}")

    return {"message": "If that email exists, a password reset link has been sent"}


@router.post("/reset-password")
@limiter.limit("5/hour")
async def reset_password(
    request: Request,
    reset_data: ResetPasswordRequest,
    db: Session = Depends(get_db)
):
    """Reset password with token"""
    user = db.query(User).filter(
        User.reset_token == reset_data.token,
        User.reset_token_expires > datetime.now(timezone.utc)
    ).first()

    if not user:
        raise HTTPException(status_code=400, detail="Invalid or expired reset token")

    # Validate new password
    is_valid, error_msg = validate_password_strength(reset_data.new_password)
    if not is_valid:
        raise HTTPException(status_code=400, detail=error_msg)

    # Update password
    user.password_hash = hash_password(reset_data.new_password)
    user.reset_token = None
    user.reset_token_expires = None
    db.commit()

    # Revoke all refresh tokens for security
    revoke_all_user_tokens(db, user.id)

    logger.info(f"Password reset: {user.email}")

    return {"message": "Password reset successfully"}


@router.post("/refresh", response_model=TokenResponse)
@limiter.limit("20/minute")
async def refresh_access_token(
    request: Request,
    refresh_request: RefreshTokenRequest,
    db: Session = Depends(get_db)
):
    """Refresh access token using refresh token"""
    user = verify_refresh_token(db, refresh_request.refresh_token)

    if not user:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Invalid or expired refresh token"
        )

    # Create new access token
    access_token = create_access_token(data={"sub": str(user.id), "email": user.email})

    return TokenResponse(
        access_token=access_token,
        token_type="bearer",
        expires_in=settings.JWT_ACCESS_TOKEN_EXPIRE_MINUTES * 60
    )


@router.post("/logout")
@limiter.limit("20/minute")
async def logout(
    request: Request,
    refresh_token: RefreshTokenRequest,
    current_user: User = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """Logout current session"""
    revoke_refresh_token(db, refresh_token.refresh_token)
    logger.info(f"User logged out: {current_user.email}")
    return {"message": "Logged out successfully"}


@router.post("/logout-all")
@limiter.limit("10/minute")
async def logout_all_devices(
    request: Request,
    current_user: User = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """Logout from all devices"""
    revoke_all_user_tokens(db, current_user.id)
    logger.info(f"User logged out from all devices: {current_user.email}")
    return {"message": "Logged out from all devices"}


@router.get("/me")
@limiter.limit("100/minute")
async def get_current_user_info(
    request: Request,
    current_user: User = Depends(get_current_user)
):
    """Get current authenticated user info"""
    return {
        "id": current_user.id,
        "email": current_user.email,
        "name": current_user.name,
        "is_verified": current_user.is_verified,
        "is_admin": current_user.is_admin,
        "created_at": current_user.created_at,
        "last_login_at": current_user.last_login_at
    }
```

### 7. Pydantic Schemas
```python
# backend/app/schemas/auth.py

from datetime import datetime
from pydantic import BaseModel, EmailStr, Field

class RegisterRequest(BaseModel):
    email: EmailStr
    password: str = Field(..., min_length=8)
    name: str | None = None

class LoginRequest(BaseModel):
    email: EmailStr
    password: str

class TokenResponse(BaseModel):
    access_token: str
    refresh_token: str | None = None
    token_type: str = "bearer"
    expires_in: int  # seconds

class RefreshTokenRequest(BaseModel):
    refresh_token: str

class ForgotPasswordRequest(BaseModel):
    email: EmailStr

class ResetPasswordRequest(BaseModel):
    token: str
    new_password: str = Field(..., min_length=8)

class VerifyEmailRequest(BaseModel):
    token: str

class UserResponse(BaseModel):
    id: int
    email: str
    name: str | None
    is_verified: bool
    is_admin: bool
    created_at: datetime
    last_login_at: datetime | None

    class Config:
        from_attributes = True
```

---

## Frontend Integration

### 1. Auth Store (Zustand)
```typescript
// frontend/src/stores/authStore.ts

import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface User {
  id: number
  email: string
  name: string | null
  is_verified: boolean
  is_admin: boolean
}

interface AuthState {
  user: User | null
  accessToken: string | null
  refreshToken: string | null
  isAuthenticated: boolean

  setAuth: (accessToken: string, refreshToken: string, user: User) => void
  clearAuth: () => void
  setUser: (user: User) => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      accessToken: null,
      refreshToken: null,
      isAuthenticated: false,

      setAuth: (accessToken, refreshToken, user) => set({
        accessToken,
        refreshToken,
        user,
        isAuthenticated: true
      }),

      clearAuth: () => set({
        user: null,
        accessToken: null,
        refreshToken: null,
        isAuthenticated: false
      }),

      setUser: (user) => set({ user })
    }),
    {
      name: 'auth-storage'
    }
  )
)
```

### 2. API Client with Auth
```typescript
// frontend/src/api/client.ts

import axios from 'axios'
import { useAuthStore } from '@/stores/authStore'

const apiClient = axios.create({
  baseURL: 'http://localhost:8000',
  headers: {
    'Content-Type': 'application/json'
  }
})

// Add auth token to requests
apiClient.interceptors.request.use(
  (config) => {
    const { accessToken } = useAuthStore.getState()
    if (accessToken) {
      config.headers.Authorization = `Bearer ${accessToken}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

// Handle token refresh on 401
apiClient.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config

    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true

      const { refreshToken, setAuth, clearAuth } = useAuthStore.getState()

      if (!refreshToken) {
        clearAuth()
        window.location.href = '/login'
        return Promise.reject(error)
      }

      try {
        const response = await axios.post('http://localhost:8000/api/auth/refresh', {
          refresh_token: refreshToken
        })

        const { access_token } = response.data
        const userResponse = await axios.get('http://localhost:8000/api/auth/me', {
          headers: { Authorization: `Bearer ${access_token}` }
        })

        setAuth(access_token, refreshToken, userResponse.data)
        originalRequest.headers.Authorization = `Bearer ${access_token}`

        return apiClient(originalRequest)
      } catch (refreshError) {
        clearAuth()
        window.location.href = '/login'
        return Promise.reject(refreshError)
      }
    }

    return Promise.reject(error)
  }
)

export default apiClient
```

### 3. Registration Page
```typescript
// frontend/src/pages/RegisterPage.tsx

import { useState } from 'react'
import { useNavigate, Link } from '@tanstack/react-router'
import { useAuthStore } from '@/stores/authStore'
import apiClient from '@/api/client'

export default function RegisterPage() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [name, setName] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const navigate = useNavigate()
  const setAuth = useAuthStore((state) => state.setAuth)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)

    try {
      const response = await apiClient.post('/api/auth/register', {
        email,
        password,
        name
      })

      const { access_token, refresh_token } = response.data

      // Fetch user info
      const userResponse = await apiClient.get('/api/auth/me', {
        headers: { Authorization: `Bearer ${access_token}` }
      })

      setAuth(access_token, refresh_token, userResponse.data)
      navigate({ to: '/feed' })
    } catch (err: any) {
      setError(err.response?.data?.detail || 'Registration failed')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 px-4">
      <div className="max-w-md w-full space-y-8">
        <div>
          <h2 className="text-center text-3xl font-bold text-gray-900">
            Create your account
          </h2>
        </div>
        <form className="mt-8 space-y-6" onSubmit={handleSubmit}>
          {error && (
            <div className="bg-red-50 border border-red-200 text-red-600 px-4 py-3 rounded">
              {error}
            </div>
          )}

          <div className="space-y-4">
            <div>
              <label htmlFor="name" className="block text-sm font-medium text-gray-700">
                Name (optional)
              </label>
              <input
                id="name"
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>

            <div>
              <label htmlFor="email" className="block text-sm font-medium text-gray-700">
                Email address
              </label>
              <input
                id="email"
                type="email"
                required
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>

            <div>
              <label htmlFor="password" className="block text-sm font-medium text-gray-700">
                Password
              </label>
              <input
                id="password"
                type="password"
                required
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
              <p className="mt-1 text-sm text-gray-500">
                Must be at least 8 characters with uppercase, lowercase, digit, and special character
              </p>
            </div>
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50"
          >
            {loading ? 'Creating account...' : 'Sign up'}
          </button>

          <div className="text-center">
            <Link to="/login" className="text-blue-600 hover:text-blue-800">
              Already have an account? Sign in
            </Link>
          </div>
        </form>
      </div>
    </div>
  )
}
```

### 4. Login Page
```typescript
// frontend/src/pages/LoginPage.tsx

import { useState } from 'react'
import { useNavigate, Link } from '@tanstack/react-router'
import { useAuthStore } from '@/stores/authStore'
import apiClient from '@/api/client'

export default function LoginPage() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const navigate = useNavigate()
  const setAuth = useAuthStore((state) => state.setAuth)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)

    try {
      const response = await apiClient.post('/api/auth/login', {
        email,
        password
      })

      const { access_token, refresh_token } = response.data

      // Fetch user info
      const userResponse = await apiClient.get('/api/auth/me', {
        headers: { Authorization: `Bearer ${access_token}` }
      })

      setAuth(access_token, refresh_token, userResponse.data)
      navigate({ to: '/feed' })
    } catch (err: any) {
      setError(err.response?.data?.detail || 'Login failed')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 px-4">
      <div className="max-w-md w-full space-y-8">
        <div>
          <h2 className="text-center text-3xl font-bold text-gray-900">
            Sign in to your account
          </h2>
        </div>
        <form className="mt-8 space-y-6" onSubmit={handleSubmit}>
          {error && (
            <div className="bg-red-50 border border-red-200 text-red-600 px-4 py-3 rounded">
              {error}
            </div>
          )}

          <div className="space-y-4">
            <div>
              <label htmlFor="email" className="block text-sm font-medium text-gray-700">
                Email address
              </label>
              <input
                id="email"
                type="email"
                required
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>

            <div>
              <label htmlFor="password" className="block text-sm font-medium text-gray-700">
                Password
              </label>
              <input
                id="password"
                type="password"
                required
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
          </div>

          <div className="flex items-center justify-between">
            <Link to="/forgot-password" className="text-sm text-blue-600 hover:text-blue-800">
              Forgot your password?
            </Link>
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50"
          >
            {loading ? 'Signing in...' : 'Sign in'}
          </button>

          <div className="text-center">
            <Link to="/register" className="text-blue-600 hover:text-blue-800">
              Don't have an account? Sign up
            </Link>
          </div>
        </form>
      </div>
    </div>
  )
}
```

### 5. Protected Route Component
```typescript
// frontend/src/components/auth/ProtectedRoute.tsx

import { ReactNode } from 'react'
import { Navigate } from '@tanstack/react-router'
import { useAuthStore } from '@/stores/authStore'

interface ProtectedRouteProps {
  children: ReactNode
  requireVerified?: boolean
  requireAdmin?: boolean
}

export default function ProtectedRoute({
  children,
  requireVerified = false,
  requireAdmin = false
}: ProtectedRouteProps) {
  const { isAuthenticated, user } = useAuthStore()

  if (!isAuthenticated) {
    return <Navigate to="/login" />
  }

  if (requireVerified && !user?.is_verified) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-center">
          <h2 className="text-2xl font-bold mb-4">Email Verification Required</h2>
          <p className="text-gray-600">Please verify your email to access this page.</p>
        </div>
      </div>
    )
  }

  if (requireAdmin && !user?.is_admin) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-center">
          <h2 className="text-2xl font-bold mb-4">Access Denied</h2>
          <p className="text-gray-600">You don't have permission to access this page.</p>
        </div>
      </div>
    )
  }

  return <>{children}</>
}
```

---

## Security Considerations

### 1. Password Security
- ✅ Bcrypt hashing (automatically salted, computationally expensive)
- ✅ Password strength requirements (configurable)
- ✅ Never return password hash in API responses
- ✅ Rate limiting on login attempts (10/min)
- ✅ Revoke all sessions on password reset

### 2. Token Security
- ✅ Short access token lifetime (15 minutes)
- ✅ Refresh tokens can be revoked
- ✅ Store refresh tokens securely in database
- ✅ Track device/IP for refresh tokens
- ✅ HTTP-only cookies option (prevent XSS)

### 3. Email Security
- ✅ Normalize emails (case-insensitive)
- ✅ Validate email format
- ✅ Time-limited verification tokens (24 hours)
- ✅ Time-limited reset tokens (1 hour)
- ✅ One-time use tokens
- ✅ Prevent email enumeration (always return success)

### 4. Rate Limiting
| Endpoint | Limit |
|----------|-------|
| Register | 5/hour |
| Login | 10/minute |
| Forgot Password | 3/hour |
| Reset Password | 5/hour |
| Refresh Token | 20/minute |
| Verify Email | 10/minute |
| Resend Verification | 3/hour |

### 5. Input Validation
- ✅ Pydantic schema validation
- ✅ Email format validation
- ✅ Password strength validation
- ✅ SQL injection prevention (SQLAlchemy ORM)
- ✅ XSS prevention (React escapes by default)

### 6. Account Security
- ✅ Account activation/deactivation
- ✅ Email verification (optional in dev)
- ✅ Multi-device session management
- ✅ Logout from all devices
- ✅ Admin role separation

---

## Testing Strategy

### 1. Unit Tests
```python
# backend/tests/test_password.py

def test_hash_password():
    """Test password hashing"""
    password = "TestPass123!"
    hashed = hash_password(password)

    assert hashed != password
    assert verify_password(password, hashed)
    assert not verify_password("WrongPass", hashed)

def test_password_strength_validation():
    """Test password strength requirements"""
    # Valid password
    is_valid, msg = validate_password_strength("TestPass123!")
    assert is_valid

    # Too short
    is_valid, msg = validate_password_strength("Test1!")
    assert not is_valid
    assert "at least 8" in msg

    # No uppercase
    is_valid, msg = validate_password_strength("testpass123!")
    assert not is_valid
    assert "uppercase" in msg
```

```python
# backend/tests/test_auth.py

def test_create_access_token():
    """Test JWT access token creation"""
    token = create_access_token({"sub": "123", "email": "test@example.com"})
    assert token is not None

    payload = verify_access_token(token)
    assert payload["sub"] == "123"
    assert payload["email"] == "test@example.com"
    assert payload["type"] == "access"

def test_register_user(client, db):
    """Test user registration"""
    response = client.post("/api/auth/register", json={
        "email": "newuser@example.com",
        "password": "TestPass123!",
        "name": "Test User"
    })

    assert response.status_code == 200
    data = response.json()
    assert "access_token" in data
    assert "refresh_token" in data

    # Verify user in database
    user = db.query(User).filter(User.email == "newuser@example.com").first()
    assert user is not None
    assert user.name == "Test User"
```

### 2. Integration Tests
```python
def test_complete_auth_flow(client, db):
    """Test registration → login → refresh → logout"""
    # Register
    register_response = client.post("/api/auth/register", json={
        "email": "test@example.com",
        "password": "TestPass123!"
    })
    assert register_response.status_code == 200

    # Login
    login_response = client.post("/api/auth/login", json={
        "email": "test@example.com",
        "password": "TestPass123!"
    })
    assert login_response.status_code == 200
    tokens = login_response.json()

    # Access protected endpoint
    me_response = client.get("/api/auth/me", headers={
        "Authorization": f"Bearer {tokens['access_token']}"
    })
    assert me_response.status_code == 200

    # Refresh token
    refresh_response = client.post("/api/auth/refresh", json={
        "refresh_token": tokens['refresh_token']
    })
    assert refresh_response.status_code == 200

    # Logout
    logout_response = client.post("/api/auth/logout",
        json={"refresh_token": tokens['refresh_token']},
        headers={"Authorization": f"Bearer {tokens['access_token']}"}
    )
    assert logout_response.status_code == 200
```

### 3. Frontend Tests
```typescript
// Test auth store
describe('Auth Store', () => {
  it('should set auth state', () => {
    const { setAuth } = useAuthStore.getState()

    setAuth('access-token', 'refresh-token', {
      id: 1,
      email: 'test@example.com',
      name: 'Test User',
      is_verified: true,
      is_admin: false
    })

    const { isAuthenticated, user } = useAuthStore.getState()
    expect(isAuthenticated).toBe(true)
    expect(user?.email).toBe('test@example.com')
  })
})
```

---

## Migration Path

### Phase 1: Database Setup (Day 1 Morning)
1. Add dependencies to `requirements.txt`
2. Create User and RefreshToken models
3. Generate and run migration (`alembic revision --autogenerate -m "Add users and auth"`)
4. Run `alembic upgrade head`
5. Update `config.py` with auth settings

### Phase 2: Backend Core (Day 1 Afternoon)
1. Implement password utilities (`services/password.py`)
2. Implement JWT utilities (`services/auth.py`)
3. Create email service (`services/email.py`)
4. Create auth dependencies (`dependencies/auth.py`)

### Phase 3: Backend Endpoints (Day 2 Morning)
1. Create auth schemas (`schemas/auth.py`)
2. Build auth router with all endpoints (`routers/auth.py`)
3. Register router in `main.py`
4. Test endpoints with Postman/curl

### Phase 4: Frontend Setup (Day 2 Afternoon)
1. Create auth store (`stores/authStore.ts`)
2. Update API client with interceptors (`api/client.ts`)
3. Build registration page
4. Build login page

### Phase 5: Frontend Pages (Day 3 Morning)
1. Build forgot password page
2. Build reset password page
3. Build verify email page
4. Create ProtectedRoute component
5. Add routes to App.tsx

### Phase 6: Testing & Polish (Day 3 Afternoon)
1. Write and run unit tests
2. Write and run integration tests
3. Test complete flow manually
4. Add error handling and loading states
5. Polish UI/UX

### Phase 7: Deployment Preparation
1. Set up SMTP credentials (Gmail, SendGrid, etc.)
2. Generate strong JWT secret key
3. Configure production environment variables
4. Set `REQUIRE_EMAIL_VERIFICATION=true`
5. Test in staging environment

---

## Email Service Setup

### Gmail Setup
1. Enable 2-factor authentication
2. Generate App Password: https://myaccount.google.com/apppasswords
3. Use app password in `SMTP_PASSWORD`

### SendGrid (Production Recommended)
```bash
SMTP_HOST=smtp.sendgrid.net
SMTP_PORT=587
SMTP_USERNAME=apikey
SMTP_PASSWORD=your-sendgrid-api-key
```

### Mailgun
```bash
SMTP_HOST=smtp.mailgun.org
SMTP_PORT=587
SMTP_USERNAME=postmaster@your-domain.com
SMTP_PASSWORD=your-mailgun-password
```

---

## Future Enhancements

### Phase 3+ Features
1. **Social Login**: Add Google/GitHub OAuth alongside email/password
2. **Two-Factor Authentication**: TOTP via authenticator apps
3. **Account Recovery**: Security questions, backup codes
4. **Session Management UI**: View and revoke active sessions
5. **Password History**: Prevent password reuse
6. **Failed Login Tracking**: Lock account after X failed attempts
7. **Login Notifications**: Email alerts on new device login
8. **User Preferences**: Save feed filters, notification settings
9. **Profile Management**: Update name, email, password
10. **Account Deletion**: GDPR-compliant account removal

---

## Resources

### Documentation
- [FastAPI Security](https://fastapi.tiangolo.com/tutorial/security/)
- [Passlib](https://passlib.readthedocs.io/)
- [python-jose JWT](https://python-jose.readthedocs.io/)
- [Bcrypt](https://github.com/pyca/bcrypt/)
- [FastAPI Mail](https://sabuhish.github.io/fastapi-mail/)

### Security Best Practices
- [OWASP Authentication Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)
- [OWASP Password Storage](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html)
- [JWT Best Practices](https://tools.ietf.org/html/rfc8725)

---

## Conclusion

This implementation provides a complete, production-ready email/password authentication system for the OpenGov platform. The architecture is:

✅ **Secure**: Bcrypt password hashing, JWT tokens, email verification, rate limiting
✅ **Scalable**: Multi-device session management, token refresh, device tracking
✅ **User-friendly**: Password reset, email verification, clear error messages
✅ **Flexible**: Optional email verification in dev, configurable password requirements
✅ **Production-ready**: Email service integration, comprehensive error handling, logging

**Estimated Implementation Time**: 2-3 days
**Priority**: High (Phase 2 requirement)
**Dependencies**: SMTP credentials for production email sending
