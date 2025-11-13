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

    # Database
    DATABASE_URL: str = os.getenv("DATABASE_URL", "sqlite:///./opengov.db")

    # Scraper settings
    SCRAPER_INTERVAL_MINUTES: int = int(os.getenv("SCRAPER_INTERVAL_MINUTES", "15"))
    SCRAPER_DAYS_LOOKBACK: int = int(os.getenv("SCRAPER_DAYS_LOOKBACK", "0"))

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
    FEDERAL_REGISTER_MAX_PAGES: int = int(os.getenv("FEDERAL_REGISTER_MAX_PAGES", "100"))

    # Environment
    DEBUG: bool = os.getenv("DEBUG", "False").lower() == "true"


settings = Settings()
