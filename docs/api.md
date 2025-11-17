# OpenGov API Documentation

## Base URL
`http://localhost:8000` (dev)

## Security

### Request Size Limits
- Maximum request body size: 10 MB
- Exceeding limit returns 413 status with message: "Request body too large"

### Input Validation
- All query parameters are validated with type constraints
- Invalid parameters return 422 status with detailed error messages

### SQL Injection Protection
- All database queries use SQLAlchemy ORM which prevents SQL injection
- User input is never interpolated into SQL

## Response Format
All responses follow this structure:
```json
{
  "success": true,
  "data": {},
  "error": null
}
```

## Endpoints

### Health Check
`GET /health`

Returns server status.

**Response:**
```json
{
  "status": "ok"
}
```

### Database Health Check
`GET /health/db`

Returns database connection status.

**Response (success):**
```json
{
  "status": "ok",
  "database": "connected"
}
```

**Response (error):**
```json
{
  "status": "error",
  "database": "disconnected",
  "error": "Connection timeout"
}
```

### Get Feed
`GET /api/feed`

Retrieve paginated list of articles.

**Query Parameters:**
- `page` (int, default 1) - Page number
- `limit` (int, default 20, max 100) - Articles per page
- `sort` (string, default "newest") - Sort order: `newest` or `oldest`

**Response:**
```json
{
  "articles": [
    {
      "id": 1,
      "title": "Federal Register Update",
      "summary": "Engaging summary here",
      "source_url": "https://federalregister.gov/...",
      "published_at": "2024-01-15T10:30:00",
      "created_at": "2024-01-15T11:00:00",
      "document_number": "2024-01234"
    }
  ],
  "page": 1,
  "limit": 20,
  "total": 150,
  "has_next": true
}
```

**Status Codes:**
- `200` - Success
- `400` - Invalid parameters
- `500` - Server error

### Get Article by ID
`GET /api/feed/{article_id}`

Get detailed view of a single article by ID.

**Parameters:**
- `article_id` (int) - Article ID

**Response:**
```json
{
  "id": 1,
  "title": "Federal Register Update",
  "summary": "Summary text",
  "source_url": "https://federalregister.gov/...",
  "published_at": "2024-01-15T10:30:00",
  "created_at": "2024-01-15T11:00:00",
  "updated_at": "2024-01-15T11:00:00",
  "document_number": "2024-01234"
}
```

**Status Codes:**
- `200` - Success
- `404` - Article not found
- `500` - Server error

### Get Article by Document Number
`GET /api/feed/document/{document_number}`

Get article by Federal Register document number.

**Parameters:**
- `document_number` (string) - Federal Register document number (e.g., "2024-01234")

**Response:**
```json
{
  "id": 1,
  "title": "Federal Register Update",
  "summary": "Summary text",
  "source_url": "https://federalregister.gov/...",
  "published_at": "2024-01-15T10:30:00",
  "created_at": "2024-01-15T11:00:00",
  "updated_at": "2024-01-15T11:00:00",
  "document_number": "2024-01234"
}
```

**Status Codes:**
- `200` - Success
- `404` - Article not found
- `500` - Server error

### Manual Scrape
`POST /api/admin/scrape`

Manually trigger Federal Register scraping job.

**Response:**
```json
{
  "status": "success",
  "message": "Scrape job completed"
}
```

### Admin Stats
`GET /api/admin/stats`

Get scraper statistics.

**Response:**
```json
{
  "total_articles": 145,
  "last_scrape_time": "2024-01-15T11:15:00",
  "last_scrape_human": "5 minutes ago"
}
```

### Scraper Runs
`GET /api/admin/scraper-runs`

Get list of recent scraper job executions for monitoring and observability.

**Query Parameters:**
- `limit` (int, default 10, max 50) - Number of runs to return

**Response:**
```json
{
  "runs": [
    {
      "id": 42,
      "started_at": "2024-01-15T11:15:00",
      "completed_at": "2024-01-15T11:15:45",
      "processed_count": 25,
      "skipped_count": 18,
      "error_count": 0,
      "success": true,
      "error_message": null,
      "duration_seconds": 45.2
    }
  ],
  "total": 156
}
```

**Status Codes:**
- `200` - Success
- `429` - Rate limit exceeded
- `500` - Server error

## Error Responses

### 404 Not Found
```json
{
  "detail": "Article not found"
}
```

### 500 Server Error
```json
{
  "detail": "Internal server error"
}
```

## Rate Limiting

**Applied to all endpoints:**
- Feed endpoints (`GET /api/feed`, `GET /api/feed/{id}`): 100 requests per minute
- Admin stats endpoint (`GET /api/admin/stats`): 50 requests per minute
- Manual scrape endpoint (`POST /api/admin/scrape`): 10 requests per minute
- Scraper runs endpoint (`GET /api/admin/scraper-runs`): 50 requests per minute

When rate limit exceeded, receives 429 status with message: "Rate limit exceeded. Maximum X requests per minute."

## Pagination
- Default limit: 20 articles
- Maximum limit: 100 articles
- Use `has_next` to determine if more pages exist
