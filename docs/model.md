# OpenGov Data Model

## Database Schema

### FRArticle
Unified model combining Federal Register raw data and processed article content. Each Federal Register document becomes one article with both raw API data and AI-processed summary for the public feed.

| Field | Type | Notes |
|-------|------|-------|
| id | Integer | Primary key |
| document_number | String(50) | Unique Federal Register ID (unique, indexed) |
| raw_data | JSON | Complete API response (for audit/debugging) |
| fetched_at | DateTime | When raw data was fetched from API (indexed) |
| title | String(500) | Article headline |
| summary | Text | AI-generated viral summary |
| source_url | String(500) | Link to Federal Register (unique, indexed) |
| published_at | DateTime | Publication date (indexed) |
| created_at | DateTime | When inserted into database |
| updated_at | DateTime | Last update time |

**Indexes:**
- `document_number` - For deduplication and lookups (unique)
- `source_url` - Enforces uniqueness, prevents duplicate articles (unique)
- `published_at` - For efficient sorting/filtering by date
- `fetched_at` - For tracking scraper runs

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

### Bookmark
User bookmarks for articles. Allows authenticated users to save articles for later reading.

| Field | Type | Notes |
|-------|------|-------|
| id | Integer | Primary key |
| user_id | Integer | Foreign key to users.id (indexed, cascade delete) |
| frarticle_id | Integer | Foreign key to frarticles.id (indexed, cascade delete) |
| is_bookmarked | Boolean | Bookmark status (default: True) |
| created_at | DateTime | When bookmark was created |
| updated_at | DateTime | Last update time |

**Indexes:**
- `user_id` - For efficient user bookmark queries
- `frarticle_id` - For article bookmark lookups
- `(user_id, is_bookmarked)` - Composite index for filtering active bookmarks
- Unique constraint on `(user_id, frarticle_id)` - Prevents duplicate bookmarks

### Like
User likes and dislikes for articles. Allows authenticated users to vote on articles.

| Field | Type | Notes |
|-------|------|-------|
| id | Integer | Primary key |
| user_id | Integer | Foreign key to users.id (indexed, cascade delete) |
| frarticle_id | Integer | Foreign key to frarticles.id (indexed, cascade delete) |
| is_positive | Boolean | True for like, False for dislike |
| created_at | DateTime | When like/dislike was created |
| updated_at | DateTime | Last update time |

**Indexes:**
- `user_id` - For efficient user like queries
- `frarticle_id` - For article like lookups
- `(user_id, is_positive)` - Composite index for filtering likes/dislikes
- Unique constraint on `(user_id, frarticle_id)` - Prevents duplicate votes

## Entity Relationship

**FRArticle** is a standalone entity with no foreign key relationships to other tables. Each Federal Register document maps to exactly one FRArticle.

**Bookmark** creates a many-to-many relationship between Users and FRArticles:
- One user can bookmark many articles
- One article can be bookmarked by many users
- The unique constraint ensures each user can only bookmark an article once

**Like** creates a many-to-many relationship between Users and FRArticles:
- One user can like/dislike many articles
- One article can be liked/disliked by many users
- The unique constraint ensures each user can only have one vote per article
- Clicking the same vote again removes it; clicking a different vote updates it

**Duplicate Prevention:**
- `document_number` has a unique constraint - prevents duplicate Federal Register documents
- `source_url` has a unique constraint - prevents duplicate articles
- The scraper checks both fields before creating new FRArticles
- Bookmarks have a unique constraint on `(user_id, frarticle_id)` to prevent duplicate bookmarks
- Likes have a unique constraint on `(user_id, frarticle_id)` to prevent duplicate votes

**API Usage:**
- Articles can be retrieved by ID: `GET /api/feed/{article_id}`
- Articles can be retrieved by Federal Register document_number: `GET /api/feed/document/{document_number}`
- Toggle bookmark: `POST /api/bookmarks/toggle` with `{frarticle_id: <id>}`
- Get user bookmarks: `GET /api/bookmarks`
- Remove bookmark: `DELETE /api/bookmarks/{frarticle_id}`
- Toggle like/dislike: `POST /api/likes/toggle` with `{frarticle_id: <id>, is_positive: <true/false>}`
- Get like status: `GET /api/likes/status/{frarticle_id}`
- Get like counts: `GET /api/likes/counts/{frarticle_id}`
- Remove like: `DELETE /api/likes/{frarticle_id}`

## Pydantic Schemas

### ArticleResponse
Used for API responses listing articles (based on FRArticle model).
- id, document_number, title, summary, source_url, published_at, created_at
- is_bookmarked (optional, indicates if current user has bookmarked)
- user_like_status (optional, null = no vote, true = liked, false = disliked)
- likes_count (total number of likes)
- dislikes_count (total number of dislikes)

### ArticleDetail
Extended response for single article views.
- All of ArticleResponse + updated_at, fetched_at

### FeedResponse
Paginated feed of articles.
- articles: List[ArticleResponse]
- page, limit, total, has_next

### BookmarkToggle
Request schema for toggling bookmark status.
- frarticle_id: int

### BookmarkResponse
Response schema for bookmark operations.
- id, user_id, frarticle_id, is_bookmarked, created_at, updated_at

### BookmarkedArticleResponse
Response schema for bookmarked articles with article details.
- id, document_number, title, summary, source_url, published_at, created_at, bookmarked_at

### LikeToggle
Request schema for toggling like/dislike status.
- frarticle_id: int
- is_positive: bool (True for like, False for dislike)

### LikeResponse
Response schema for like operations.
- id, user_id, frarticle_id, is_positive, created_at, updated_at
