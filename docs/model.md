# OpenGov Data Model

## Database Schema

### Article
Represents a processed government update.

| Field | Type | Notes |
|-------|------|-------|
| id | Integer | Primary key |
| federal_register_id | Integer | Foreign key to FederalRegister (indexed) |
| title | String(500) | Article headline |
| summary | Text | AI-generated viral summary |
| source_url | String(500) | Link to Federal Register (unique) |
| published_at | DateTime | Publication date (indexed) |
| created_at | DateTime | When inserted into database |
| updated_at | DateTime | Last update time |

**Indexes:**
- `published_at` - For efficient sorting/filtering
- `source_url` - Enforces uniqueness, prevents duplicate articles

### FederalRegister
Raw entries from Federal Register API.

| Field | Type | Notes |
|-------|------|-------|
| id | Integer | Primary key |
| document_number | String(50) | Unique Federal Register ID (indexed) |
| raw_data | JSON | Complete API response |
| fetched_at | DateTime | When fetched (indexed) |
| processed | Boolean | Whether Article was created (indexed) |

**Indexes:**
- `document_number` - For deduplication
- `(processed, fetched_at)` - For finding unprocessed entries

### ScraperRun
Execution records for scraper jobs (monitoring/observability).

| Field | Type | Notes |
|-------|------|-------|
| id | Integer | Primary key |
| started_at | DateTime | When job started (indexed) |
| completed_at | DateTime | When job finished |
| processed_count | Integer | Number of articles created |
| skipped_count | Integer | Number of duplicates skipped |
| error_count | Integer | Number of processing errors |
| success | Boolean | Whether job completed successfully |
| error_message | String(500) | Error details if failed |

**Computed Fields** (in API response):
- `duration_seconds` - Calculated as (completed_at - started_at).total_seconds()

**Indexes:**
- `started_at` - For querying recent runs

### Agency
Federal government agencies from Federal Register API.

| Field | Type | Notes |
|-------|------|-------|
| id | Integer | Primary key |
| fr_agency_id | Integer | Federal Register agency ID (unique, indexed) |
| name | String(500) | Full agency name (indexed) |
| short_name | String(200) | Abbreviated agency name |
| slug | String(200) | URL-friendly identifier (unique, indexed) |
| description | Text | Agency description (optional) |
| url | String(500) | Agency website URL (optional) |
| json_url | String(500) | Federal Register API URL for this agency |
| parent_id | Integer | Parent agency ID if applicable |
| raw_data | JSON | Complete API response |
| created_at | DateTime | When inserted into database |
| updated_at | DateTime | Last update time |

**Indexes:**
- `fr_agency_id` - For deduplication
- `slug` - For lookups by slug
- `name` - For searching/filtering by name

## Entity Relationship

```
FederalRegister (1) ----> (many) Article
        (id)           federal_register_id
```

Each Federal Register entry produces exactly one Article. The `federal_register_id` foreign key ensures traceability from Article back to its source document.

## Pydantic Schemas

### ArticleResponse
Used for API responses listing articles.
- id, title, summary, source_url, published_at, created_at

### ArticleDetail
Extended response for single article views.
- All of ArticleResponse + updated_at

### FeedResponse
Paginated feed of articles.
- articles: List[ArticleResponse]
- page, limit, total, has_next
