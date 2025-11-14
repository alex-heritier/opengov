# OpenGov Codebase Code Simplification & Cleanup Analysis

## Executive Summary
Analysis of 14+ backend Python files across routers, services, models, schemas, and workers reveals **9 major code simplification opportunities** affecting code maintainability, performance, and consistency.

---

## 1. REDUNDANT CODE - Article Response Building (HIGH PRIORITY)

### Issue
Three near-identical code blocks build article responses differently across `feed.py`:

**File:** `/home/user/opengov/backend/app/routers/feed.py`

**Problem Pattern (Lines 62-67, 103-106, 125-129):**
```python
# Pattern 1 (get_feed, lines 62-67)
article_dict = ArticleResponse.from_orm(article).model_dump()
if article.federal_register_entry:
    article_dict["document_number"] = article.federal_register_entry.document_number
article_responses.append(ArticleResponse(**article_dict))

# Pattern 2 (get_article_by_document_number, lines 103-106)
article_dict = ArticleDetail.from_orm(article).model_dump()
article_dict["document_number"] = article.federal_register_entry.document_number
return ArticleDetail(**article_dict)

# Pattern 3 (get_article, lines 125-129)
article_dict = ArticleDetail.from_orm(article).model_dump()
if article.federal_register_entry:
    article_dict["document_number"] = article.federal_register_entry.document_number
return ArticleDetail(**article_dict)
```

**Impact:** 
- Code duplication makes maintenance difficult
- Inconsistent null checking (Pattern 1 has guard, Pattern 2 doesn't)
- Inefficient: converts ORM → dict → back to Pydantic model

**Recommendation:** Extract helper function:
```python
def _build_article_response(article: Article, use_detail: bool = False) -> Union[ArticleResponse, ArticleDetail]:
    """Build article response with optional document_number"""
    schema = ArticleDetail if use_detail else ArticleResponse
    article_dict = schema.from_orm(article).model_dump()
    if article.federal_register_entry:
        article_dict["document_number"] = article.federal_register_entry.document_number
    return schema(**article_dict)
```

---

## 2. REPETITIVE TEXT TRUNCATION (MEDIUM PRIORITY)

### Issue
Same fallback truncation pattern repeated 4 times in `grok.py`:

**File:** `/home/user/opengov/backend/app/services/grok.py`

**Pattern (Lines 59, 89-91, 96, 99, 102):**
```python
# Line 59
return text[:SUMMARY_MAX_FALLBACK] + "..." if len(text) > SUMMARY_MAX_FALLBACK else text

# Line 89-91
return (
    text[:SUMMARY_MAX_FALLBACK] + "..."
    if len(text) > SUMMARY_MAX_FALLBACK else text
)

# Line 96, 99, 102 - same pattern repeated
```

**Impact:** 
- Duplicated business logic across 4 exception handlers
- Hard-to-maintain pattern in multiple places

**Recommendation:** Extract helper function:
```python
def _truncate_fallback(text: str) -> str:
    """Truncate text to fallback max length with ellipsis"""
    if not text or len(text) <= SUMMARY_MAX_FALLBACK:
        return text
    return text[:SUMMARY_MAX_FALLBACK] + "..."
```

---

## 3. VERBOSE AGENCY FIELD UPDATES (HIGH PRIORITY)

### Issue
Repetitive field-by-field comparison and update pattern in agency sync:

**File:** `/home/user/opengov/backend/app/services/federal_register.py`

**Problem (Lines 185-206):**
16 repetitive lines checking each field individually:
```python
updated = False
if existing_agency.name != agency_data.get("name"):
    existing_agency.name = agency_data.get("name")
    updated = True
if existing_agency.short_name != agency_data.get("short_name"):
    existing_agency.short_name = agency_data.get("short_name")
    updated = True
# ... 5 more nearly identical blocks ...
if existing_agency.parent_id != agency_data.get("parent_id"):
    existing_agency.parent_id = agency_data.get("parent_id")
    updated = True
```

**Impact:**
- 22 lines of boilerplate code
- Error-prone when adding new fields
- Hard to understand intent

**Recommendation:** Create reusable update helper:
```python
AGENCY_FIELDS = ["name", "short_name", "slug", "description", "url", "json_url", "parent_id"]

def _update_agency_fields(agency: Agency, data: dict) -> bool:
    """Update agency fields from data dict, return True if any updated"""
    updated = False
    for field in AGENCY_FIELDS:
        old_value = getattr(agency, field)
        new_value = data.get(field)
        if old_value != new_value:
            setattr(agency, field, new_value)
            updated = True
    return updated
```

---

## 4. MAGIC NUMBERS & STRINGS - Constants Scattered (MEDIUM PRIORITY)

### Issue
Important numeric values are hardcoded throughout codebase without clear centralization:

**File:** `/home/user/opengov/backend/app/workers/scraper.py`

**Lines with magic numbers:**
- Line 13: `BATCH_SIZE = 50` (good - defined)
- Line 105: `[:1000]` - Magic number for full_text truncation (undocumented)

**File:** `/home/user/opengov/backend/app/services/grok.py`

**Lines with magic numbers:**
- Line 10: `GROK_TEMPERATURE = 0.7`
- Line 11: `GROK_MAX_TOKENS = 300`
- Line 12: `SUMMARY_MAX_FALLBACK = 200`
- Line 26: `280 characters` (in prompt, hardcoded as comment)

**File:** `/home/user/opengov/backend/app/routers/feed.py`

**Lines with magic numbers:**
- Line 17, 79, 110: `@limiter.limit("100/minute")` (3 times)
- Line 22: `limit: int = Query(20, ge=1, le=100)` - max 100
- Line 45: `max-age=300` (5 minutes - undocumented)

**File:** `/home/user/opengov/backend/app/services/federal_register.py`

**Lines with magic numbers:**
- Line 83: `await asyncio.sleep(0.5)` - API delay (undocumented)

**Recommendation:** Centralize in `config.py`:
```python
# Grok API settings
GROK_TEMPERATURE: float = float(os.getenv("GROK_TEMPERATURE", "0.7"))
GROK_MAX_TOKENS: int = int(os.getenv("GROK_MAX_TOKENS", "300"))
GROK_SUMMARY_MAX_FALLBACK: int = int(os.getenv("GROK_SUMMARY_MAX_FALLBACK", "200"))
GROK_SUMMARY_TARGET_CHARS: int = int(os.getenv("GROK_SUMMARY_TARGET_CHARS", "280"))

# Scraper settings
SCRAPER_FULL_TEXT_TRUNCATE: int = int(os.getenv("SCRAPER_FULL_TEXT_TRUNCATE", "1000"))
SCRAPER_API_REQUEST_DELAY: float = float(os.getenv("SCRAPER_API_REQUEST_DELAY", "0.5"))
SCRAPER_BATCH_SIZE: int = int(os.getenv("SCRAPER_BATCH_SIZE", "50"))

# Feed API settings
FEED_RATE_LIMIT: str = os.getenv("FEED_RATE_LIMIT", "100/minute")
FEED_DEFAULT_PAGE_SIZE: int = int(os.getenv("FEED_DEFAULT_PAGE_SIZE", "20"))
FEED_MAX_PAGE_SIZE: int = int(os.getenv("FEED_MAX_PAGE_SIZE", "100"))
FEED_CACHE_TTL_SECONDS: int = int(os.getenv("FEED_CACHE_TTL_SECONDS", "300"))
```

---

## 5. DUPLICATE SCHEMA FIELDS (MEDIUM PRIORITY)

### Issue
`document_number` field is duplicated across schema hierarchy:

**File:** `/home/user/opengov/backend/app/schemas/article.py`

**Lines 12, 20:**
```python
class ArticleResponse(BaseModel):
    # ... other fields ...
    document_number: str | None = None  # Line 12

class ArticleDetail(ArticleResponse):
    updated_at: datetime
    document_number: str | None = None  # Line 20 - DUPLICATE

    class Config:
        from_attributes = True  # Lines 22-23 also duplicated
```

**Impact:**
- Field is inherited, so re-declaring it is redundant
- Both classes have duplicate `Config` classes
- Violates DRY principle

**Recommendation:** Simplify:
```python
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
```

---

## 6. INCONSISTENT TYPE HINTS (MEDIUM PRIORITY)

### Issue
Missing or incomplete type hints in services and schemas:

**File:** `/home/user/opengov/backend/app/services/federal_register.py`

**Line 19:** Missing return type annotation
```python
async def fetch_recent_documents(days: int = 1) -> list:  # Generic 'list'
    # Should be -> list[dict] or more specific
```

**Line 109:** Same issue
```python
async def fetch_agencies() -> list:  # Should be -> list[dict]
```

**Line 152:** Parameter type is vague
```python
def store_agencies(db: Session, agencies_data: list) -> dict:
    # agencies_data should be -> list[dict]
```

**File:** `/home/user/opengov/backend/app/services/grok.py`

**Lines 41, 105:** Missing type hints for factory function
```python
def _get_summarizer():  # No return type
    # ...
    return _summarize_text_real  # Should declare Callable type
```

**File:** `/home/user/opengov/backend/app/schemas/feed.py`

**Line 1:** Using legacy typing import for Python 3.9+
```python
from typing import List  # Should use list[ArticleResponse] in Python 3.9+
```

**Recommendation:** 
```python
from typing import Callable

# In services/federal_register.py
async def fetch_recent_documents(days: int = 1) -> list[dict]:
async def fetch_agencies() -> list[dict]:
def store_agencies(db: Session, agencies_data: list[dict]) -> dict:

# In services/grok.py
def _get_summarizer() -> Callable[[str], str]:

# In schemas/feed.py
articles: list[ArticleResponse]  # Python 3.9+ syntax
```

---

## 7. UNUSED IMPORTS (LOW PRIORITY)

### Issue
Imported but unused modules:

**File:** `/home/user/opengov/backend/app/routers/admin.py`

**Line 4:** `from sqlalchemy import desc`
- Imported but never used in the file

**File:** `/home/user/opengov/backend/app/routers/common.py`

**Line 9:** Nested import inside function
```python
def get_ip_from_x_forwarded_for(request: Request):
    from app.config import settings  # Should be at top
```

**Impact:** 
- Import at line 9 of common.py is inefficient for repeated calls
- Creates temporary import overhead on every request

**Recommendation:**
- Remove `desc` import from `admin.py:4`
- Move import to top of `common.py`

---

## 8. INCONSISTENT ERROR HANDLING (MEDIUM PRIORITY)

### Issue
Different error handling strategies across similar functions:

**File:** `/home/user/opengov/backend/app/services/federal_register.py`

**fetch_recent_documents (Lines 91-100):** Catches exceptions silently, returns empty list
```python
except httpx.TimeoutException:
    logger.error(...)
    return []  # Silent failure
except httpx.HTTPError as e:
    logger.error(...)
    return []  # Silent failure
```

**store_agencies (Lines 247-256):** Raises exception after rollback
```python
except Exception as e:
    db.rollback()
    logger.error(...)
    raise  # Explicit failure
```

**Impact:**
- Caller can't distinguish between "no data" and "error occurred"
- API scrape silently fails, but admin endpoint explicitly fails
- Inconsistent user experience

**Recommendation:** Use custom exception or consistent pattern:
```python
class FederalRegisterAPIError(Exception):
    """Raised when Federal Register API call fails"""
    pass

async def fetch_recent_documents(...) -> list[dict]:
    try:
        # ... implementation ...
    except httpx.TimeoutException as e:
        raise FederalRegisterAPIError(f"Timeout: {e}") from e
    except httpx.HTTPError as e:
        raise FederalRegisterAPIError(f"HTTP error: {e}") from e
```

---

## 9. INEFFICIENT DATABASE QUERIES (MEDIUM PRIORITY)

### Issue
Redundant database round-trips in scraper:

**File:** `/home/user/opengov/backend/app/workers/scraper.py`

**Lines 81-92:** Two separate queries for duplicate detection
```python
existing_article = db.query(Article).filter(
    Article.source_url == source_url
).first()

existing_fed_entry = db.query(FederalRegister).filter(
    FederalRegister.document_number == doc_number
).first()

if existing_article or existing_fed_entry:
    # Skip
```

**Impact:**
- Two database calls per document when one could suffice
- For 100 documents = 200 queries (in worst case)
- No indexing strategy clarified

**Recommendation:** Combine into single query:
```python
# Check both in one query using OR condition
existing = db.query(Article).join(
    FederalRegister, Article.federal_register_id == FederalRegister.id
).filter(
    (Article.source_url == source_url) | 
    (FederalRegister.document_number == doc_number)
).first()

if existing:
    skipped_count += 1
    continue
```

Or use bulk check for performance:
```python
# Fetch all document_numbers we're about to process
existing_doc_numbers = {
    row[0] for row in db.query(FederalRegister.document_number).filter(
        FederalRegister.document_number.in_([doc.get("document_number") for doc in documents])
    )
}

for doc in documents:
    if doc.get("document_number") in existing_doc_numbers:
        skipped_count += 1
        continue
```

---

## Summary Table

| Issue | Severity | Files | Lines | Effort |
|-------|----------|-------|-------|--------|
| 1. Redundant article response building | HIGH | feed.py | 62-67, 103-106, 125-129 | 30 min |
| 2. Repetitive text truncation | MEDIUM | grok.py | 59, 89-91, 96, 99, 102 | 15 min |
| 3. Verbose agency field updates | HIGH | federal_register.py | 185-206 | 45 min |
| 4. Magic numbers scattered | MEDIUM | Multiple | Various | 45 min |
| 5. Duplicate schema fields | MEDIUM | article.py | 12, 20, 22-23 | 10 min |
| 6. Incomplete type hints | MEDIUM | Multiple | Various | 30 min |
| 7. Unused imports | LOW | admin.py, common.py | 4, 9 | 5 min |
| 8. Inconsistent error handling | MEDIUM | federal_register.py | 91-100, 247-256 | 30 min |
| 9. Inefficient queries | MEDIUM | scraper.py | 81-92 | 25 min |

**Total Estimated Cleanup Time: 3.5-4 hours**

---

## Implementation Priority

1. **Week 1 (Quick wins):**
   - #7: Remove unused imports (5 min)
   - #5: Clean duplicate schema fields (10 min)
   - #2: Extract text truncation helper (15 min)

2. **Week 2 (Medium effort):**
   - #1: Extract article response builder (30 min)
   - #6: Add complete type hints (30 min)
   - #4: Centralize magic numbers in config (45 min)

3. **Week 3 (High impact):**
   - #3: Refactor agency field updates (45 min)
   - #8: Implement consistent error handling (30 min)
   - #9: Optimize database queries (25 min)

---

## Testing Impact

- All changes should maintain backward compatibility at API level
- Unit tests should be added for new helper functions
- Integration tests should verify scraper behavior with refactored code
- Performance tests should validate query optimizations
