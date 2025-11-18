from datetime import datetime, timezone
from typing import Optional
from sqlalchemy import Boolean, DateTime, Integer, String
from sqlalchemy.orm import Mapped, mapped_column
from fastapi_users.db import SQLAlchemyBaseUserTable
from app.database import Base


class User(SQLAlchemyBaseUserTable[int], Base):
    """User model for authentication with fastapi-users"""

    __tablename__ = "users"

    # Primary key (inherited from SQLAlchemyBaseUserTable but we define it explicitly)
    id: Mapped[int] = mapped_column(Integer, primary_key=True, index=True)

    # Required by fastapi-users
    email: Mapped[str] = mapped_column(String(255), unique=True, nullable=False, index=True)
    hashed_password: Mapped[str] = mapped_column(String(1024), nullable=False)
    is_active: Mapped[bool] = mapped_column(Boolean, default=True, nullable=False)
    is_superuser: Mapped[bool] = mapped_column(Boolean, default=False, nullable=False)
    is_verified: Mapped[bool] = mapped_column(Boolean, default=False, nullable=False)

    # Additional fields for Google OAuth (optional)
    google_id: Mapped[Optional[str]] = mapped_column(
        String(255), unique=True, nullable=True, index=True
    )
    name: Mapped[Optional[str]] = mapped_column(
        String(255), nullable=True
    )
    picture_url: Mapped[Optional[str]] = mapped_column(String(500), nullable=True)

    # Political leaning
    political_leaning: Mapped[Optional[str]] = mapped_column(
        String(50), nullable=True
    )

    # Timestamps (use timezone-aware UTC)
    created_at: Mapped[datetime] = mapped_column(
        DateTime, default=lambda: datetime.now(timezone.utc), nullable=False
    )
    updated_at: Mapped[datetime] = mapped_column(
        DateTime,
        default=lambda: datetime.now(timezone.utc),
        onupdate=lambda: datetime.now(timezone.utc),
        nullable=False,
    )
    last_login_at: Mapped[Optional[datetime]] = mapped_column(
        DateTime, nullable=True
    )

    def __repr__(self):
        return f"<User(id={self.id}, email={self.email}, name={self.name})>"
