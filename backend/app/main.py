import logging
import time
from contextlib import asynccontextmanager
from logging.handlers import RotatingFileHandler
from fastapi import FastAPI, Request
from fastapi.middleware.cors import CORSMiddleware
from slowapi import Limiter
from slowapi.util import get_remote_address
from slowapi.errors import RateLimitExceeded
from fastapi.exceptions import RequestValidationError
from fastapi.responses import JSONResponse
from apscheduler.schedulers.asyncio import AsyncIOScheduler
from app.config import settings
from app.routers import feed, admin
from app.workers.scraper import fetch_and_process
from app.database import engine, Base

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
    handlers=[
        logging.StreamHandler(),  # Console
        RotatingFileHandler("scraper.log", maxBytes=10*1024*1024, backupCount=5)  # File with rotation
    ]
)
logger = logging.getLogger(__name__)

# NOTE: Database tables are managed by Alembic migrations.
# Run migrations separately using: alembic upgrade head
# Do not use Base.metadata.create_all() as it bypasses migration tracking and causes schema drift.

# Scheduler instance
scheduler = AsyncIOScheduler()

# Import limiter from common to avoid circular imports
from app.routers.common import limiter


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Handle app startup and shutdown"""
    # Startup
    logger.info("Starting OpenGov API")
    scheduler.add_job(
        fetch_and_process,
        "interval",
        minutes=settings.SCRAPER_INTERVAL_MINUTES,
        id="federal_register_scraper",
    )
    scheduler.start()
    logger.info(f"Scraper scheduled to run every {settings.SCRAPER_INTERVAL_MINUTES} minutes")

    yield

    # Shutdown
    logger.info("Shutting down OpenGov API")
    scheduler.shutdown(wait=True)  # Wait for running jobs to complete


app = FastAPI(
    title="OpenGov API",
    version="0.1.0",
    lifespan=lifespan,
)

# Add rate limiter
app.state.limiter = limiter
app.add_exception_handler(RateLimitExceeded, lambda r, exc: JSONResponse(
    status_code=429,
    content={"detail": "Rate limit exceeded. Maximum 100 requests per minute."}
))

# Add request size limit middleware
@app.middleware("http")
async def request_size_limit_middleware(request: Request, call_next):
    """Limit request body size"""
    content_length = request.headers.get("content-length")
    if content_length and int(content_length) > settings.MAX_REQUEST_SIZE_BYTES:
        return JSONResponse(
            status_code=413,
            content={"detail": f"Request body too large (max {settings.MAX_REQUEST_SIZE_BYTES // (1024*1024)} MB)"}
        )
    return await call_next(request)

# Add request/response logging middleware
@app.middleware("http")
async def logging_middleware(request: Request, call_next):
    """Log request/response details"""
    start_time = time.time()
    
    # Log request
    logger.debug(f"→ {request.method} {request.url.path}")
    
    # Process request
    response = await call_next(request)
    
    # Log response
    process_time = time.time() - start_time
    logger.info(
        f"← {request.method} {request.url.path} {response.status_code} ({process_time:.3f}s)"
    )
    
    return response

# Add CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=settings.ALLOWED_ORIGINS,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Add custom exception handler for validation errors
@app.exception_handler(RequestValidationError)
async def validation_exception_handler(request: Request, exc: RequestValidationError):
    """Handle validation errors"""
    logger.warning(f"Validation error: {exc}")
    return JSONResponse(
        status_code=422,
        content={
            "detail": "Invalid request parameters",
            "errors": [{"field": str(err["loc"]), "message": err["msg"]} for err in exc.errors()]
        }
    )

# Include routers
app.include_router(feed.router)
app.include_router(admin.router)


@app.get("/health")
async def health_check():
    """Health check endpoint"""
    return {"status": "ok"}


@app.get("/health/db")
async def health_check_db():
    """Database health check endpoint"""
    db = None
    try:
        from app.database import SessionLocal
        from sqlalchemy import text
        db = SessionLocal()
        # Simple query to verify database connection
        db.execute(text("SELECT 1"))
        return {
            "status": "ok",
            "database": "connected"
        }
    except Exception as e:
        logger.error(f"Database health check failed: {e}")
        return {
            "status": "error",
            "database": "disconnected",
            "error": str(e)
        }
    finally:
        if db:
            db.close()
