# Code Cleanup Implementation Guide

This guide provides specific code examples and before/after comparisons for each issue.

## Issue #1: Extract Article Response Builder

### Current Code (3 locations)
**backend/app/routers/feed.py - Lines 62-67 (get_feed)**
```python
article_dict = ArticleResponse.from_orm(article).model_dump()
if article.federal_register_entry:
    article_dict["document_number"] = article.federal_register_entry.document_number
article_responses.append(ArticleResponse(**article_dict))
```

**Lines 103-106 (get_article_by_document_number)**
```python
article_dict = ArticleDetail.from_orm(article).model_dump()
article_dict["document_number"] = article.federal_register_entry.document_number
return ArticleDetail(**article_dict)
```

**Lines 125-129 (get_article)**
```python
article_dict = ArticleDetail.from_orm(article).model_dump()
if article.federal_register_entry:
    article_dict["document_number"] = article.federal_register_entry.document_number
return ArticleDetail(**article_dict)
```

### Improved Code

Add this helper function at the top of feed.py:
```python
from typing import Union

def _build_article_response(
    article: Article, detail: bool = False
) -> Union[ArticleResponse, ArticleDetail]:
    """
    Build article response with document_number if available.
    
    Args:
        article: Article ORM model
        detail: If True, return ArticleDetail; otherwise ArticleResponse
    
    Returns:
        ArticleResponse or ArticleDetail with document_number populated
    """
    schema_class = ArticleDetail if detail else ArticleResponse
    article_dict = schema_class.from_orm(article).model_dump()
    
    # Add document_number if federal_register_entry exists
    if article.federal_register_entry:
        article_dict["document_number"] = article.federal_register_entry.document_number
    
    return schema_class(**article_dict)


# Updated get_feed (lines 62-67)
article_responses = [_build_article_response(article) for article in articles]

# Updated get_article_by_document_number (lines 103-106)
return _build_article_response(article, detail=True)

# Updated get_article (lines 125-129)
return _build_article_response(article, detail=True)
```

---

## Issue #2: Extract Text Truncation Helper

### Current Code
**backend/app/services/grok.py - Lines 59, 89-91, 96, 99, 102**

Pattern repeated 4 times:
```python
# Location 1 (line 59)
return text[:SUMMARY_MAX_FALLBACK] + "..." if len(text) > SUMMARY_MAX_FALLBACK else text

# Location 2 (lines 89-91)
return (
    text[:SUMMARY_MAX_FALLBACK] + "..."
    if len(text) > SUMMARY_MAX_FALLBACK else text
)

# Locations 3-4 (lines 96, 99, 102) - same pattern
```

### Improved Code

Add helper function after constants (around line 33):
```python
def _truncate_fallback(text: str) -> str:
    """
    Truncate text to fallback max length with ellipsis.
    
    Args:
        text: Text to truncate
    
    Returns:
        Original text if <= max length, otherwise truncated with "..."
    """
    if not text or len(text) <= SUMMARY_MAX_FALLBACK:
        return text
    return text[:SUMMARY_MAX_FALLBACK] + "..."
```

Replace all occurrences:
```python
# Old (line 59)
return text[:SUMMARY_MAX_FALLBACK] + "..." if len(text) > SUMMARY_MAX_FALLBACK else text
# New
return _truncate_fallback(text)

# Old (lines 89-91)
return (
    text[:SUMMARY_MAX_FALLBACK] + "..."
    if len(text) > SUMMARY_MAX_FALLBACK else text
)
# New
return _truncate_fallback(text)

# Same for lines 96, 99, 102
```

---

## Issue #3: Refactor Agency Field Updates

### Current Code
**backend/app/services/federal_register.py - Lines 185-206**

```python
updated = False
if existing_agency.name != agency_data.get("name"):
    existing_agency.name = agency_data.get("name")
    updated = True
if existing_agency.short_name != agency_data.get("short_name"):
    existing_agency.short_name = agency_data.get("short_name")
    updated = True
if existing_agency.slug != agency_data.get("slug"):
    existing_agency.slug = agency_data.get("slug")
    updated = True
if existing_agency.description != agency_data.get("description"):
    existing_agency.description = agency_data.get("description")
    updated = True
if existing_agency.url != agency_data.get("url"):
    existing_agency.url = agency_data.get("url")
    updated = True
if existing_agency.json_url != agency_data.get("json_url"):
    existing_agency.json_url = agency_data.get("json_url")
    updated = True
if existing_agency.parent_id != agency_data.get("parent_id"):
    existing_agency.parent_id = agency_data.get("parent_id")
    updated = True
```

### Improved Code

Add constant and helper at top of federal_register.py:
```python
# Fields to sync from agency data to model
AGENCY_UPDATEABLE_FIELDS = [
    "name",
    "short_name",
    "slug",
    "description",
    "url",
    "json_url",
    "parent_id",
]


def _update_agency_fields(agency: Agency, data: dict) -> bool:
    """
    Update agency fields from data dictionary.
    
    Args:
        agency: Agency ORM model to update
        data: Dictionary with new values
    
    Returns:
        True if any fields were updated, False otherwise
    """
    updated = False
    for field in AGENCY_UPDATEABLE_FIELDS:
        old_value = getattr(agency, field)
        new_value = data.get(field)
        if old_value != new_value:
            setattr(agency, field, new_value)
            updated = True
    return updated
```

Then replace lines 185-206 with:
```python
updated = _update_agency_fields(existing_agency, agency_data)
```

---

## Issue #4: Centralize Magic Numbers

### Current Code (scattered throughout)

In `backend/app/routers/feed.py`:
```python
@limiter.limit("100/minute")  # Line 17, 79, 110 - repeated 3 times
limit: int = Query(20, ge=1, le=100)  # Line 22
response.headers["Cache-Control"] = "public, max-age=300"  # Line 45
```

In `backend/app/services/federal_register.py`:
```python
await asyncio.sleep(0.5)  # Line 83
```

In `backend/app/services/grok.py`:
```python
GROK_TEMPERATURE = 0.7  # Line 10
GROK_MAX_TOKENS = 300  # Line 11
SUMMARY_MAX_FALLBACK = 200  # Line 12
```

In `backend/app/workers/scraper.py`:
```python
summary_text = abstract or doc.get("full_text", "")[:1000]  # Line 105
```

### Improved Code

**Add to backend/app/config.py** (in Settings class):
```python
# API Rate Limiting
FEED_RATE_LIMIT: str = os.getenv("FEED_RATE_LIMIT", "100/minute")
ADMIN_RATE_LIMIT: str = os.getenv("ADMIN_RATE_LIMIT", "50/minute")
ADMIN_SCRAPE_RATE_LIMIT: str = os.getenv("ADMIN_SCRAPE_RATE_LIMIT", "10/minute")
ADMIN_SYNC_AGENCIES_RATE_LIMIT: str = os.getenv("ADMIN_SYNC_AGENCIES_RATE_LIMIT", "5/minute")

# Feed Pagination
FEED_DEFAULT_PAGE_SIZE: int = int(os.getenv("FEED_DEFAULT_PAGE_SIZE", "20"))
FEED_MAX_PAGE_SIZE: int = int(os.getenv("FEED_MAX_PAGE_SIZE", "100"))

# Feed Caching
FEED_CACHE_TTL_SECONDS: int = int(os.getenv("FEED_CACHE_TTL_SECONDS", "300"))

# Grok API Configuration
GROK_TEMPERATURE: float = float(os.getenv("GROK_TEMPERATURE", "0.7"))
GROK_MAX_TOKENS: int = int(os.getenv("GROK_MAX_TOKENS", "300"))
GROK_SUMMARY_MAX_FALLBACK: int = int(os.getenv("GROK_SUMMARY_MAX_FALLBACK", "200"))

# Scraper Configuration
SCRAPER_FULL_TEXT_TRUNCATE: int = int(os.getenv("SCRAPER_FULL_TEXT_TRUNCATE", "1000"))
SCRAPER_API_REQUEST_DELAY: float = float(os.getenv("SCRAPER_API_REQUEST_DELAY", "0.5"))

# API Request Delays
FEDERAL_REGISTER_REQUEST_DELAY: float = float(os.getenv("FEDERAL_REGISTER_REQUEST_DELAY", "0.5"))
```

Then update usage:

**In feed.py:**
```python
# Old
@limiter.limit("100/minute")

# New
@limiter.limit(settings.FEED_RATE_LIMIT)

# Old
limit: int = Query(20, ge=1, le=100)

# New
limit: int = Query(settings.FEED_DEFAULT_PAGE_SIZE, ge=1, le=settings.FEED_MAX_PAGE_SIZE)

# Old
response.headers["Cache-Control"] = "public, max-age=300"

# New
response.headers["Cache-Control"] = f"public, max-age={settings.FEED_CACHE_TTL_SECONDS}"
```

**In federal_register.py (line 83):**
```python
# Old
await asyncio.sleep(0.5)

# New
await asyncio.sleep(settings.FEDERAL_REGISTER_REQUEST_DELAY)
```

**In grok.py:**
```python
# Old
GROK_MODEL = "grok-4-fast"
GROK_TEMPERATURE = 0.7
GROK_MAX_TOKENS = 300
SUMMARY_MAX_FALLBACK = 200

# New (use settings instead)
response = await client.post(
    ...
    json={
        "model": "grok-4-fast",
        "temperature": settings.GROK_TEMPERATURE,
        "max_tokens": settings.GROK_MAX_TOKENS,
    },
)
```

**In scraper.py (line 105):**
```python
# Old
summary_text = abstract or doc.get("full_text", "")[:1000]

# New
summary_text = abstract or doc.get("full_text", "")[:settings.SCRAPER_FULL_TEXT_TRUNCATE]
```

---

## Issue #5: Remove Duplicate Schema Fields

### Current Code
**backend/app/schemas/article.py - Lines 1-24**

```python
from datetime import datetime
from pydantic import BaseModel


class ArticleResponse(BaseModel):
    id: int
    title: str
    summary: str
    source_url: str
    published_at: datetime
    created_at: datetime
    document_number: str | None = None

    class Config:
        from_attributes = True


class ArticleDetail(ArticleResponse):
    updated_at: datetime
    document_number: str | None = None  # DUPLICATE!

    class Config:
        from_attributes = True  # DUPLICATE!
```

### Improved Code

```python
from datetime import datetime
from pydantic import BaseModel


class ArticleResponse(BaseModel):
    id: int
    title: str
    summary: str
    source_url: str
    published_at: datetime
    created_at: datetime
    document_number: str | None = None

    class Config:
        from_attributes = True


class ArticleDetail(ArticleResponse):
    updated_at: datetime
    # document_number is inherited from ArticleResponse
    # Config is inherited from ArticleResponse
```

---

## Issue #6: Add Complete Type Hints

### Current Code
**backend/app/services/federal_register.py**

```python
# Line 19 - Generic type
async def fetch_recent_documents(days: int = 1) -> list:

# Line 109 - Generic type
async def fetch_agencies() -> list:

# Line 152 - Vague parameter type
def store_agencies(db: Session, agencies_data: list) -> dict:
```

**backend/app/services/grok.py**

```python
# Line 105 - Missing return type
def _get_summarizer():
    if settings.USE_MOCK_GROK:
        from app.services.grok_mock import summarize_text as mock_summarize
        return mock_summarize
    else:
        return _summarize_text_real
```

**backend/app/schemas/feed.py**

```python
# Line 1 - Legacy typing
from typing import List

class FeedResponse(BaseModel):
    articles: List[ArticleResponse]  # Should use list[]
```

### Improved Code

**In federal_register.py:**
```python
from typing import Any

# Line 19
async def fetch_recent_documents(days: int = 1) -> list[dict[str, Any]]:

# Line 109
async def fetch_agencies() -> list[dict[str, Any]]:

# Line 152
def store_agencies(db: Session, agencies_data: list[dict[str, Any]]) -> dict[str, int]:
```

**In grok.py:**
```python
from typing import Callable

# Line 105
def _get_summarizer() -> Callable[[str], str]:
    if settings.USE_MOCK_GROK:
        from app.services.grok_mock import summarize_text as mock_summarize
        return mock_summarize
    else:
        return _summarize_text_real
```

**In schemas/feed.py:**
```python
# Remove: from typing import List
# Use Python 3.9+ syntax

class FeedResponse(BaseModel):
    articles: list[ArticleResponse]  # Remove List
```

---

## Issue #7: Remove Unused Imports

### Current Code

**backend/app/routers/admin.py - Line 4**
```python
from sqlalchemy import desc  # UNUSED!
```

**backend/app/routers/common.py - Lines 1-15**
```python
from fastapi import Request
from slowapi import Limiter
from slowapi.util import get_remote_address
from app.database import SessionLocal


def get_ip_from_x_forwarded_for(request: Request):
    """Extract client IP from X-Forwarded-For header if present"""
    from app.config import settings  # Nested import! Bad for performance
```

### Improved Code

**In admin.py:**
```python
# Just remove line 4
# from sqlalchemy import desc
```

**In common.py:**
```python
from fastapi import Request
from slowapi import Limiter
from slowapi.util import get_remote_address
from app.database import SessionLocal
from app.config import settings  # Move to top


def get_ip_from_x_forwarded_for(request: Request):
    """Extract client IP from X-Forwarded-For header if present"""
    # Remove: from app.config import settings
    if settings.BEHIND_PROXY:
        # ... rest of function
```

---

## Issue #8: Standardize Error Handling

### Current Code

**In fetch_recent_documents (Lines 91-100):**
```python
except httpx.TimeoutException:
    logger.error(f"Federal Register API timeout...")
    return []  # Silent failure
except httpx.HTTPError as e:
    logger.error(f"Federal Register API HTTP error...")
    return []  # Silent failure
```

**In store_agencies (Lines 247-256):**
```python
try:
    db.commit()
    # ...
except Exception as e:
    db.rollback()
    logger.error(f"Error committing agencies...")
    raise  # Explicit failure!
```

### Improved Code

**Add custom exception at top of federal_register.py:**
```python
class FederalRegisterAPIError(Exception):
    """Raised when Federal Register API call fails"""
    pass
```

**Update fetch_recent_documents:**
```python
async def fetch_recent_documents(days: int = 1) -> list[dict[str, Any]]:
    # ... existing code ...
    try:
        # ... implementation ...
    except httpx.TimeoutException as e:
        logger.error(f"Federal Register API timeout...")
        raise FederalRegisterAPIError(
            f"Timeout after {settings.FEDERAL_REGISTER_TIMEOUT}s"
        ) from e
    except httpx.HTTPError as e:
        logger.error(f"Federal Register API HTTP error...")
        raise FederalRegisterAPIError(f"HTTP error: {e}") from e
    except Exception as e:
        logger.error(f"Unexpected error fetching Federal Register...")
        raise FederalRegisterAPIError(f"Unexpected error: {e}") from e
```

**Update store_agencies:**
```python
def store_agencies(db: Session, agencies_data: list[dict[str, Any]]) -> dict[str, int]:
    # ... implementation ...
    try:
        db.commit()
        logger.info(f"Stored agencies: ...")
    except Exception as e:
        db.rollback()
        logger.error(f"Error committing agencies: {e}", exc_info=True)
        raise FederalRegisterAPIError(f"Failed to store agencies: {e}") from e
```

---

## Issue #9: Optimize Database Queries

### Current Code
**backend/app/workers/scraper.py - Lines 81-92**

```python
# Two separate queries - inefficient!
existing_article = db.query(Article).filter(
    Article.source_url == source_url
).first()

existing_fed_entry = db.query(FederalRegister).filter(
    FederalRegister.document_number == doc_number
).first()

if existing_article or existing_fed_entry:
    logger.debug("  → Already in database, skipping")
    skipped_count += 1
    continue
```

### Improved Code - Option 1: Single Query with JOIN

```python
# Single query checking both conditions
existing = (
    db.query(Article)
    .join(FederalRegister, Article.federal_register_id == FederalRegister.id, isouter=True)
    .filter(
        (Article.source_url == source_url) | 
        (FederalRegister.document_number == doc_number)
    )
    .first()
)

if existing:
    logger.debug("  → Already in database, skipping")
    skipped_count += 1
    continue
```

### Improved Code - Option 2: Bulk Fetch (Better for 100+ documents)

```python
# Add this at start of fetch_and_process, after fetching documents
doc_numbers = [doc.get("document_number") for doc in documents if doc.get("document_number")]
source_urls = [doc.get("html_url") for doc in documents if doc.get("html_url")]

existing_doc_numbers = {
    row[0] for row in db.query(FederalRegister.document_number).filter(
        FederalRegister.document_number.in_(doc_numbers)
    ).all()
}

existing_source_urls = {
    row[0] for row in db.query(Article.source_url).filter(
        Article.source_url.in_(source_urls)
    ).all()
}

# Then in the loop:
for doc in documents:
    doc_number = doc.get("document_number", "UNKNOWN")
    source_url = doc.get("html_url", "")
    
    if source_url in existing_source_urls or doc_number in existing_doc_numbers:
        logger.debug("  → Already in database, skipping")
        skipped_count += 1
        continue
    
    # ... rest of processing ...
```

---

## Summary

These 9 issues affect:
- **Code maintainability**: Duplicated code is harder to maintain
- **Performance**: Extra database queries, nested imports, inefficient conversions
- **Type safety**: Missing type hints reduce IDE support and catch fewer bugs
- **Consistency**: Different error handling strategies confuse developers

Implementing these changes will result in:
- 20-30% less code duplication
- Faster scraper (fewer queries)
- Better IDE support (complete type hints)
- Consistent error handling
- Easier to extend (magic numbers centralized)

Estimated time: 3.5-4 hours for all issues.
