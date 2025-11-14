import os
from dotenv import load_dotenv

load_dotenv()


class Settings:
    """Application configuration"""

    # API Keys
    GROK_API_KEY: str = os.getenv("GROK_API_KEY", "")

    # External APIs
    FEDERAL_REGISTER_API_URL: str = os.getenv(
        "FEDERAL_REGISTER_API_URL", "https://www.federalregister.gov/api/v1"
    )
    GROK_API_URL: str = os.getenv(
        "GROK_API_URL", "https://api.x.ai/v1"
    )

    # Database (set in .env)
    DATABASE_URL: str = os.getenv("DATABASE_URL", "")

    # Scraper settings
    SCRAPER_INTERVAL_MINUTES: int = int(os.getenv("SCRAPER_INTERVAL_MINUTES", "15"))
    SCRAPER_DAYS_LOOKBACK: int = int(os.getenv("SCRAPER_DAYS_LOOKBACK", "1"))

    # CORS
    ALLOWED_ORIGINS: list = os.getenv(
        "ALLOWED_ORIGINS", "http://localhost:5173,http://localhost:3000"
    ).split(",")

    # Timeouts (seconds)
    FEDERAL_REGISTER_TIMEOUT: int = int(os.getenv("FEDERAL_REGISTER_TIMEOUT", "30"))
    GROK_TIMEOUT: int = int(os.getenv("GROK_TIMEOUT", "60"))

    # Limits
    MAX_REQUEST_SIZE_BYTES: int = int(os.getenv("MAX_REQUEST_SIZE_BYTES", "10485760"))  # 10 MB
    FEDERAL_REGISTER_PER_PAGE: int = int(os.getenv("FEDERAL_REGISTER_PER_PAGE", "100"))
    FEDERAL_REGISTER_MAX_PAGES: int = int(os.getenv("FEDERAL_REGISTER_MAX_PAGES", "2"))

    # Environment
    DEBUG: bool = os.getenv("DEBUG", "False").lower() in ("true", "1", "yes")
    BEHIND_PROXY: bool = os.getenv("BEHIND_PROXY", "False").lower() in ("true", "1", "yes")
    USE_MOCK_GROK: bool = os.getenv("USE_MOCK_GROK", "False").lower() in ("true", "1", "yes")

    def validate(self):
        """Validate critical configuration on startup"""
        import sys
        import logging

        # Skip validation during testing
        if "pytest" in sys.modules or "unittest" in sys.modules:
            return True

        # Warn about missing API key but don't fail
        # (grok.py has graceful fallback for missing key)
        if not self.GROK_API_KEY or not self.GROK_API_KEY.strip():
            logging.warning(
                "GROK_API_KEY is not configured. Summaries will be "
                "truncated text instead of AI-generated. "
                "Get your API key from https://x.ai/"
            )
        return True


settings = Settings()
settings.validate()  # Validate on import
