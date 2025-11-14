# Code Cleanup Quick Reference Guide

## Overview
This document provides a quick lookup for all 9 code simplification opportunities identified.

## At-a-Glance Issues

### 1. Article Response Building (HIGH)
- **File**: `backend/app/routers/feed.py`
- **Lines**: 62-67, 103-106, 125-129
- **Problem**: 3 identical response building blocks
- **Fix**: Extract to helper function
- **Time**: 30 min

### 2. Text Truncation Repetition (MEDIUM)
- **File**: `backend/app/services/grok.py`
- **Lines**: 59, 89-91, 96, 99, 102
- **Problem**: Same pattern repeated 4 times
- **Fix**: Extract `_truncate_fallback()` helper
- **Time**: 15 min

### 3. Agency Field Updates (HIGH)
- **File**: `backend/app/services/federal_register.py`
- **Lines**: 185-206
- **Problem**: 22 lines of repetitive field comparisons
- **Fix**: Use loop with AGENCY_FIELDS list
- **Time**: 45 min

### 4. Magic Numbers Scattered (MEDIUM)
- **Files**: `scraper.py`, `grok.py`, `feed.py`, `federal_register.py`
- **Problem**: Hardcoded values throughout codebase
- **Fix**: Move to `config.py` with environment variables
- **Time**: 45 min
- **Examples**:
  - `FEED_RATE_LIMIT = "100/minute"` (3 occurrences)
  - `max-age=300` (cache TTL)
  - `[:1000]` (text truncation)
  - `0.5` (API delay)

### 5. Duplicate Schema Fields (MEDIUM)
- **File**: `backend/app/schemas/article.py`
- **Lines**: 12, 20, 22-23
- **Problem**: `document_number` redeclared in child class
- **Fix**: Remove duplicate field and Config class from child
- **Time**: 10 min

### 6. Missing Type Hints (MEDIUM)
- **Files**: `federal_register.py`, `grok.py`, `feed.py`
- **Problem**: Generic `list` instead of `list[dict]`, missing Callable hints
- **Fix**: Add specific type hints throughout
- **Time**: 30 min
- **Examples**:
  - Line 19: `async def fetch_recent_documents(...) -> list[dict]`
  - Line 109: `async def fetch_agencies() -> list[dict]`
  - Line 1 (schemas): Use `list[]` instead of `List[]`

### 7. Unused Imports (LOW)
- **File 1**: `backend/app/routers/admin.py:4`
  - Remove: `from sqlalchemy import desc`
- **File 2**: `backend/app/routers/common.py:9`
  - Move nested import to module top
- **Time**: 5 min

### 8. Inconsistent Error Handling (MEDIUM)
- **File**: `backend/app/services/federal_register.py`
- **Lines**: 91-100 vs 247-256
- **Problem**: Some functions silently fail, others raise
- **Fix**: Standardize with custom exceptions
- **Time**: 30 min

### 9. Inefficient Database Queries (MEDIUM)
- **File**: `backend/app/workers/scraper.py`
- **Lines**: 81-92
- **Problem**: Two queries per document (200 for 100 documents)
- **Fix**: Single joined query or bulk fetch
- **Time**: 25 min

## Implementation Checklist

### Phase 1: Quick Wins (30 min)
- [ ] Remove unused `desc` import from admin.py
- [ ] Move import in common.py to top
- [ ] Remove duplicate `document_number` from ArticleDetail schema
- [ ] Remove duplicate Config class from ArticleDetail

### Phase 2: Helpers & Type Safety (75 min)
- [ ] Create `_truncate_fallback()` in grok.py
- [ ] Add type hints to function signatures
- [ ] Update schema typing imports

### Phase 3: Refactoring (120 min)
- [ ] Extract article response builder
- [ ] Add AGENCY_FIELDS constant and refactor update logic
- [ ] Centralize magic numbers in config.py
- [ ] Create consistent error handling strategy
- [ ] Optimize database queries in scraper

## File Impact Summary

| File | Issues | Total LOC Changed |
|------|--------|-------------------|
| feed.py | #1 | ~30 |
| grok.py | #2, #6 | ~20 |
| federal_register.py | #3, #4, #6, #8 | ~50 |
| article.py (schemas) | #5, #6 | ~10 |
| scraper.py | #4, #9 | ~20 |
| admin.py | #7 | ~1 |
| common.py | #7 | ~1 |
| config.py | #4 | ~30 |
| **Total** | | **~162 LOC** |

## Testing Recommendations

1. **Unit Tests**: Add tests for new helper functions
   - `test_truncate_fallback()`
   - `test_build_article_response()`
   - `test_update_agency_fields()`

2. **Integration Tests**: Verify scraper behavior
   - Duplicate detection still works
   - Agency sync still works

3. **Performance Tests**: Validate optimizations
   - Query count reduction
   - Response time improvements

## Configuration Changes Example

```python
# In config.py - add these new settings

# Grok API configuration
GROK_TEMPERATURE: float = float(os.getenv("GROK_TEMPERATURE", "0.7"))
GROK_MAX_TOKENS: int = int(os.getenv("GROK_MAX_TOKENS", "300"))
GROK_SUMMARY_MAX_FALLBACK: int = int(os.getenv("GROK_SUMMARY_MAX_FALLBACK", "200"))

# Feed API configuration
FEED_RATE_LIMIT: str = os.getenv("FEED_RATE_LIMIT", "100/minute")
FEED_CACHE_TTL_SECONDS: int = int(os.getenv("FEED_CACHE_TTL_SECONDS", "300"))
FEED_DEFAULT_PAGE_SIZE: int = int(os.getenv("FEED_DEFAULT_PAGE_SIZE", "20"))

# Scraper configuration
SCRAPER_BATCH_SIZE: int = int(os.getenv("SCRAPER_BATCH_SIZE", "50"))
SCRAPER_FULL_TEXT_TRUNCATE: int = int(os.getenv("SCRAPER_FULL_TEXT_TRUNCATE", "1000"))
SCRAPER_API_REQUEST_DELAY: float = float(os.getenv("SCRAPER_API_REQUEST_DELAY", "0.5"))

# Federal Register API configuration
FEDERAL_REGISTER_API_REQUEST_DELAY: float = float(os.getenv("FEDERAL_REGISTER_API_REQUEST_DELAY", "0.5"))
```

## Review Checklist Before Committing

- [ ] All tests pass (`pytest backend/`)
- [ ] No new linting errors (`black --check backend/`)
- [ ] Type hints complete (`mypy backend/`)
- [ ] API behavior unchanged
- [ ] Performance tests validate improvements
- [ ] Documentation updated if needed

---

**Total Estimated Time to Implement All Issues: 3.5-4 hours**

See `CODE_CLEANUP_ANALYSIS.md` for detailed explanations.
