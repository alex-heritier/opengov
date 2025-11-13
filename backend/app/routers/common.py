from fastapi import Request
from sqlalchemy.orm import Session
from slowapi import Limiter
from slowapi.util import get_remote_address
from app.database import SessionLocal


def get_ip_from_x_forwarded_for(request: Request):
    """Extract client IP from X-Forwarded-For header if present (for proxy setups)"""
    from app.config import settings
    if settings.BEHIND_PROXY:
        x_forwarded_for = request.headers.get("X-Forwarded-For")
        if x_forwarded_for:
            # Use rightmost IP (added by our trusted proxy)
            return x_forwarded_for.split(",")[-1].strip()
    return get_remote_address(request)


# Rate limiter (use X-Forwarded-For for proxies)
limiter = Limiter(key_func=get_ip_from_x_forwarded_for)


def get_db():
    """Dependency for database session"""
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()
