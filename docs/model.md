# OpenGov Data Model

## Formatting Guide

This document documents all database models. When adding or updating models:

1. **JSON Example First**: Start each model section with a JSON example showing all fields with realistic values
2. **Field Descriptions**: After the JSON, list each field with:
   - Field name in backticks
   - Description of what the field stores
   - Constraints: `(unique)`, `(nullable)`, `(indexed)`
3. **Timestamps**: Don't list `created_at` and `updated_at` in field descriptions unless they have special behavior
4. **Constraints Section**: For tables with foreign keys or unique constraints, add a Constraints section
5. **Indexes Section**: Document all indexes including composite indexes
6. **One Model Per Section**: Each model gets its own `## ModelName` heading

**JSON Conventions:**
- Use `null` for nullable fields
- Use realistic example values
- Wrap complex objects with comment: `{ /* ... */ }`
- Use ISO 8601 timestamps: `2025-01-01T10:30:00.000000Z`

Database backed models should have standard PostgreSQL timestamps even if not explicitly stated in this file.

## User

Users of the application. Can authenticate via email/password or Google OAuth.

{
  "id": 1,
  "email": "user@example.com",
  "is_active": 1,
  "is_superuser": 0,
  "is_verified": 0,
  "google_id": null,
  "name": "John Doe",
  "picture_url": "https://example.com/avatar.png",
  "political_leaning": null,
  "state": "CA",
  "created_at": "2025-01-01T10:30:00.000000Z",
  "updated_at": "2025-01-01T10:30:00.000000Z",
  "last_login_at": "2025-01-10T14:30:00.000000Z"
}

**Auth Fields:**
- `email`: User's email address (unique, used for login)
- `hashed_password`: bcrypt hash of user's password (not exposed in API)
- `google_id`: Google OAuth user ID (nullable, set when signed up via Google)
- `is_active`: Whether the account is active (1 = active, 0 = disabled)
- `is_superuser`: Admin flag for superuser privileges (1 = superuser)
- `is_verified`: Email verification status (1 = verified)

**Profile Fields:**
- `name`: User's display name (nullable)
- `picture_url`: Profile picture URL from Google OAuth (nullable)
- `political_leaning`: User's political leaning for personalized feed (nullable)
- `state`: User's US state (2-letter code, e.g., "CA", "NY", nullable)

**Timestamps:**
- `created_at`: When the user account was created
- `updated_at`: When the user account was last updated
- `last_login_at`: When the user last logged in (nullable)

## Agency

Federal government agencies from Federal Register API.

{
  "id": 1,
  "fr_agency_id": 1,
  "raw_name": "Department of Agriculture",
  "name": "Department of Agriculture",
  "short_name": "USDA",
  "slug": "department-of-agriculture",
  "description": "The USDA provides leadership on food, agriculture, natural resources, and related issues.",
  "url": "https://www.usda.gov",
  "json_url": "https://www.federalregister.gov/api/v1/agencies/1.json",
  "parent_id": null,
  "raw_data": { /* complete API response */ },
  "created_at": "2025-01-01T10:30:00.000000Z",
  "updated_at": "2025-01-01T10:30:00.000000Z"
}

**Fields:**
- `fr_agency_id`: Federal Register agency ID (unique)
- `raw_name`: Original agency name from API
- `name`: Display name for the agency
- `short_name`: Abbreviated agency name (nullable)
- `slug`: URL-friendly identifier (unique)
- `description`: Agency description (nullable)
- `url`: Agency website URL (nullable)
- `json_url`: Federal Register API URL for this agency (nullable)
- `parent_id`: Parent agency ID if applicable (nullable)
- `raw_data`: Complete API response as JSON

**Indexes:**
- `fr_agency_id` - For deduplication
- `slug` - For lookups by slug
- `name` - For searching/filtering by name

## FeedEntry

Unified feed entries table. Contains denormalized data for fast feed retrieval.

{
  "id": 1,
  "source_type": "federal_register",
  "title": "Notice of Proposed Rulemaking: Food Safety Standards",
  "short_text": "The FDA is proposing new food safety standards for processing facilities...",
  "key_points": [
    "New safety requirements for food processors",
    "Public comment period opens",
    "Implementation deadline in 2026"
  ],
  "political_score": -15,
  "impact_score": "medium",
  "source_url": "https://www.federalregister.gov/documents/2025/01/10/2025-01234",
  "published_at": "2025-01-10T10:00:00.000000Z",
  "created_at": "2025-01-10T10:30:00.000000Z",
  "updated_at": "2025-01-10T10:30:00.000000Z"
}

**Fields:**
- `source_type`: Type of source (e.g., "federal_register" for Federal Register documents)
- `title`: Entry headline
- `short_text`: AI-generated summary (1-2 sentences)
- `key_points`: JSON array of key takeaways (nullable)
- `political_score`: AI-generated political leaning from -100 (left) to 100 (right), 0 = neutral (nullable)
- `impact_score`: AI-generated impact level: "low" (routine), "medium" (notable), "high" (major news) (nullable)
- `source_url`: Link to original document
- `published_at`: Publication date

**Indexes:**
- `published_at DESC` - For efficient sorting/filtering by date
- `source_type` - For filtering by source type

## PolicyDocument

Unified model combining Federal Register raw data and processed document content. Each Federal Register document becomes one entry with both raw API data and AI-processed summary for the public feed.

{
  "id": 1,
  "feed_entry_id": 1,
  "source": "federal_register",
  "source_id": "2025-01234",
  "document_number": "2025-01234",
  "unique_key": "federal_register:2025-01234",
  "raw_data": { /* complete API response */ },
  "fetched_at": "2025-01-10T10:30:00.000000Z",
  "title": "Notice of Proposed Rulemaking: Food Safety Standards",
  "agency": "Food and Drug Administration",
  "summary": "The FDA is proposing new food safety standards for processing facilities...",
  "keypoints": [
    "New safety requirements for food processors",
    "Public comment period opens",
    "Implementation deadline in 2026"
  ],
  "impact_score": "medium",
  "political_score": -15,
  "source_url": "https://www.federalregister.gov/documents/2025/01/10/2025-01234",
  "published_at": "2025-01-10T10:00:00.000000Z",
  "document_type": "Notice",
  "pdf_url": "https://www.federalregister.gov/2025-01234.pdf",
  "created_at": "2025-01-10T10:30:00.000000Z",
  "updated_at": "2025-01-10T10:30:00.000000Z"
}

**Fields:**
- `feed_entry_id`: Foreign key to feed_entries.id
- `source`: Data source identifier (e.g., "federal_register" for Federal Register)
- `source_id`: Source-specific document ID
- `document_number`: Federal Register document number (unique per source)
- `unique_key`: Composite unique identifier (source:document_number)
- `raw_data`: Complete API response for audit/debugging
- `fetched_at`: When raw data was fetched from API
- `title`: Document headline
- `agency`: Government agency name from Federal Register (nullable)
- `summary`: AI-generated viral summary (1-2 sentences)
- `keypoints`: JSON array of key takeaways (nullable)
- `impact_score`: AI-generated impact level: "low" (routine), "medium" (notable), "high" (major news) (nullable)
- `political_score`: AI-generated political leaning from -100 (left) to 100 (right), 0 = neutral (nullable)
- `source_url`: Link to original document
- `published_at`: Publication date
- `document_type`: Type of Federal Register document (e.g., "Notice", "Rule", "Proposed Rule")
- `pdf_url`: Link to PDF version (nullable)

**Indexes:**
- `unique_key` - Primary deduplication key (unique)
- `document_number` - For Federal Register lookups
- `published_at` - For efficient sorting/filtering by date
- `source` - For filtering by source

## RawEntry

Ingestion log storing raw upstream data for each document. One row per upstream document.

{
  "id": 1,
  "source_key": "federal_register",
  "external_id": "2025-01234",
  "raw_data": { /* complete API response */ },
  "fetched_at": "2025-01-10T10:30:00.000000Z",
  "policy_document_id": 1,
  "created_at": "2025-01-10T10:30:00.000000Z"
}

**Fields:**
- `source_key`: Data source identifier (e.g., "federal_register")
- `external_id`: Source-specific document ID (e.g., document_number for Federal Register)
- `raw_data`: Complete API response JSON
- `fetched_at`: When data was fetched from upstream API
- `policy_document_id`: Foreign key to policy_documents.id
- `created_at`: When the raw entry was created

**Constraints:**
- `UNIQUE (source_key, external_id)` - One raw entry per upstream document
- `FK policy_document_id → policy_documents(id) ON DELETE CASCADE`

**Indexes:**
- `policy_document_id` - For looking up raw data by document

## Bookmark

User bookmarks for feed entries. Allows authenticated users to save entries for later reading.

{
  "id": 1,
  "user_id": 1,
  "feed_entry_id": 1,
  "created_at": "2025-01-10T10:30:00.000000Z",
  "updated_at": "2025-01-10T10:30:00.000000Z"
}

**Fields:**
- `user_id`: Foreign key to users.id
- `feed_entry_id`: Foreign key to feed_entries.id

**Behavior:**
- Row presence means bookmarked
- Unbookmarking deletes the row

**Constraints:**
- Primary key on `(user_id, feed_entry_id)` - Prevents duplicate bookmarks
- Foreign keys with CASCADE delete

**Indexes:**
- `user_id` - For efficient user bookmark queries
- `feed_entry_id` - For entry bookmark lookups

## Like

User likes for feed entries. Allows authenticated users to vote on entries.

{
  "id": 1,
  "user_id": 1,
  "feed_entry_id": 1,
  "value": 1,
  "created_at": "2025-01-10T10:30:00.000000Z",
  "updated_at": "2025-01-10T10:30:00.000000Z"
}

**Fields:**
- `user_id`: Foreign key to users.id
- `feed_entry_id`: Foreign key to feed_entries.id
- `value`: Vote value (1 = like, -1 = dislike)

**Behavior:**
- Clicking the same vote again removes it
- Clicking a different vote updates it

**Constraints:**
- Primary key on `(user_id, feed_entry_id)` - Prevents duplicate votes
- Foreign keys with CASCADE delete
- Check constraint: `value IN (1, -1)`

**Indexes:**
- `user_id` - For efficient user like queries
- `feed_entry_id` - For entry like lookups
- `(feed_entry_id, value)` - For counting likes/dislikes
