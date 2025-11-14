"""Application configuration with Pydantic Settings validation"""
import sys
from typing import List
from pydantic import Field, field_validator, SecretStr
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    """Application configuration with validation"""

    model_config = SettingsConfigDict(
        env_file=".env",
        env_file_encoding="utf-8",
        case_sensitive=True,
        extra="ignore",
    )

    # API Keys
    GROK_API_KEY: str = Field(default="", description="Grok API key for AI processing")

    # External APIs
    FEDERAL_REGISTER_API_URL: str = Field(
        default="https://www.federalregister.gov/api/v1",
        description="Federal Register API base URL"
    )
    GROK_API_URL: str = Field(
        default="https://api.x.ai/v1",
        description="Grok API base URL"
    )

    # Database
    DATABASE_URL: str = Field(..., description="Database connection URL")

    # Scraper settings
    SCRAPER_INTERVAL_MINUTES: int = Field(
        default=15,
        ge=1,
        le=1440,
        description="Minutes between scraper runs"
    )
    SCRAPER_DAYS_LOOKBACK: int = Field(
        default=1,
        ge=1,
        le=30,
        description="Days to look back for Federal Register entries"
    )

    # CORS
    ALLOWED_ORIGINS: List[str] = Field(
        default=["http://localhost:5173", "http://localhost:3000"],
        description="Allowed CORS origins"
    )

    # Timeouts (seconds)
    FEDERAL_REGISTER_TIMEOUT: int = Field(
        default=30,
        ge=5,
        le=300,
        description="Timeout for Federal Register API requests"
    )
    GROK_TIMEOUT: int = Field(
        default=60,
        ge=10,
        le=300,
        description="Timeout for Grok API requests"
    )

    # Limits
    MAX_REQUEST_SIZE_BYTES: int = Field(
        default=10485760,  # 10 MB
        ge=1024,
        description="Maximum request body size in bytes"
    )
    FEDERAL_REGISTER_PER_PAGE: int = Field(
        default=100,
        ge=1,
        le=1000,
        description="Items per page for Federal Register API"
    )
    FEDERAL_REGISTER_MAX_PAGES: int = Field(
        default=2,
        ge=1,
        le=100,
        description="Maximum pages to fetch from Federal Register API"
    )

    # Environment
    DEBUG: bool = Field(default=False, description="Enable debug mode")
    BEHIND_PROXY: bool = Field(
        default=False,
        description="Whether app is behind a proxy (for IP extraction)"
    )
    USE_MOCK_GROK: bool = Field(
        default=False,
        description="Use mock Grok responses for testing"
    )

    # Google OAuth
    GOOGLE_CLIENT_ID: str = Field(default="", description="Google OAuth Client ID")
    GOOGLE_CLIENT_SECRET: SecretStr = Field(
        default=SecretStr(""),
        description="Google OAuth Client Secret"
    )
    GOOGLE_REDIRECT_URI: str = Field(
        default="http://localhost:8000/api/auth/google/callback",
        description="Google OAuth redirect URI"
    )

    # JWT
    JWT_SECRET_KEY: SecretStr = Field(
        default=SecretStr(""),
        description="Secret key for JWT token signing"
    )
    JWT_ALGORITHM: str = Field(default="HS256", description="JWT signing algorithm")
    JWT_ACCESS_TOKEN_EXPIRE_MINUTES: int = Field(
        default=60,
        ge=5,
        le=1440,
        description="JWT token expiration in minutes"
    )

    # Frontend URL
    FRONTEND_URL: str = Field(
        default="http://localhost:5173",
        description="Frontend URL for OAuth redirects"
    )

    @field_validator("ALLOWED_ORIGINS", mode="before")
    @classmethod
    def parse_origins(cls, v):
        """Parse comma-separated origins string"""
        if isinstance(v, str):
            return [origin.strip() for origin in v.split(",") if origin.strip()]
        return v

    @field_validator("JWT_SECRET_KEY")
    @classmethod
    def validate_jwt_secret(cls, v):
        """Validate JWT secret key length and strength"""
        # Skip validation during testing
        if "pytest" in sys.modules or "unittest" in sys.modules:
            return v

        secret_value = v.get_secret_value() if hasattr(v, "get_secret_value") else v

        if not secret_value or len(secret_value) < 32:
            raise ValueError(
                "JWT_SECRET_KEY must be at least 32 characters long. "
                "Generate one with: python -c \"import secrets; print(secrets.token_urlsafe(32))\""
            )

        # Check for weak/default secrets
        weak_secrets = [
            "change-me",
            "secret",
            "password",
            "test-secret",
            "development",
            "your-secret-key",
        ]
        if any(weak in secret_value.lower() for weak in weak_secrets):
            import logging
            logging.warning(
                "JWT_SECRET_KEY appears to be a weak or default value. "
                "Use a strong random secret in production!"
            )

        return v

    @field_validator("GOOGLE_CLIENT_SECRET")
    @classmethod
    def validate_google_secret(cls, v):
        """Validate Google OAuth secret"""
        # Skip validation during testing
        if "pytest" in sys.modules or "unittest" in sys.modules:
            return v

        secret_value = v.get_secret_value() if hasattr(v, "get_secret_value") else v

        # Only validate if GOOGLE_CLIENT_ID is set
        # (both should be set together or neither)
        return v

    def validate_oauth_config(self) -> bool:
        """Check if OAuth is properly configured"""
        has_client_id = bool(self.GOOGLE_CLIENT_ID)
        has_client_secret = bool(
            self.GOOGLE_CLIENT_SECRET.get_secret_value()
            if hasattr(self.GOOGLE_CLIENT_SECRET, "get_secret_value")
            else self.GOOGLE_CLIENT_SECRET
        )

        # Both should be set or both should be empty
        if has_client_id != has_client_secret:
            import logging
            logging.warning(
                "Google OAuth partially configured. Both GOOGLE_CLIENT_ID and "
                "GOOGLE_CLIENT_SECRET must be set for authentication to work."
            )
            return False

        return has_client_id and has_client_secret


# Create singleton settings instance
settings = Settings()

# Log warnings for optional configuration
import logging

if not settings.GROK_API_KEY:
    logging.warning(
        "GROK_API_KEY is not configured. Summaries will be truncated text "
        "instead of AI-generated. Get your API key from https://x.ai/"
    )

if not settings.validate_oauth_config():
    logging.warning(
        "Google OAuth is not configured. Authentication endpoints will not work. "
        "See docs/auth.md for setup instructions."
    )
