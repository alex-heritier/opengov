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

Database backed models should have standard SQLite timestamps even if not explicitly stated in this file.

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

## FRArticle

Unified model combining Federal Register raw data and processed article content. Each Federal Register document becomes one article with both raw API data and AI-processed summary for the public feed.

{
  "id": 1,
  "source": "fedreg",
  "source_id": "2025-01234",
  "document_number": "2025-01234",
  "unique_key": "fedreg:2025-01234",
  "raw_data": { /* complete API response */ },
  "fetched_at": "2025-01-10T10:30:00.000000Z",
  "title": "Notice of Proposed Rulemaking: Food Safety Standards",
  "summary": "The FDA is proposing new food safety standards for processing facilities...",
  "source_url": "https://www.federalregister.gov/documents/2025/01/10/2025-01234",
  "published_at": "2025-01-10T10:00:00.000000Z",
  "document_type": "Notice",
  "pdf_url": "https://www.federalregister.gov/2025-01234.pdf",
  "created_at": "2025-01-10T10:30:00.000000Z",
  "updated_at": "2025-01-10T10:30:00.000000Z"
}

**Fields:**
- `source`: Data source identifier (e.g., "fedreg" for Federal Register)
- `source_id`: Source-specific document ID
- `document_number`: Federal Register document number (unique per source)
- `unique_key`: Composite unique identifier (source:document_number)
- `raw_data`: Complete API response for audit/debugging
- `fetched_at`: When raw data was fetched from API
- `title`: Article headline
- `summary`: AI-generated viral summary
- `source_url`: Link to original document
- `published_at`: Publication date
- `document_type`: Type of Federal Register document (e.g., "Notice", "Rule", "Proposed Rule")
- `pdf_url`: Link to PDF version (nullable)

**Indexes:**
- `unique_key` - Primary deduplication key (unique)
- `document_number` - For Federal Register lookups
- `published_at` - For efficient sorting/filtering by date
- `source` - For filtering by source

## Bookmark

User bookmarks for articles. Allows authenticated users to save articles for later reading.

{
  "id": 1,
  "user_id": 1,
  "frarticle_id": 1,
  "is_bookmarked": 1,
  "created_at": "2025-01-10T10:30:00.000000Z",
  "updated_at": "2025-01-10T10:30:00.000000Z"
}

**Fields:**
- `user_id`: Foreign key to users.id
- `frarticle_id`: Foreign key to frarticles.id
- `is_bookmarked`: Bookmark status (1 = bookmarked, 0 = removed)

**Constraints:**
- Unique constraint on `(user_id, frarticle_id)` - Prevents duplicate bookmarks
- Foreign keys with CASCADE delete

**Indexes:**
- `user_id` - For efficient user bookmark queries
- `frarticle_id` - For article bookmark lookups

## Like

User likes for articles. Allows authenticated users to vote on articles.

{
  "id": 1,
  "user_id": 1,
  "frarticle_id": 1,
  "is_liked": 1,
  "created_at": "2025-01-10T10:30:00.000000Z",
  "updated_at": "2025-01-10T10:30:00.000000Z"
}

**Fields:**
- `user_id`: Foreign key to users.id
- `frarticle_id`: Foreign key to frarticles.id
- `is_liked`: Like status (1 = liked, 0 = disliked)

**Behavior:**
- Clicking the same vote again removes it
- Clicking a different vote updates it

**Constraints:**
- Unique constraint on `(user_id, frarticle_id)` - Prevents duplicate votes
- Foreign keys with CASCADE delete

**Indexes:**
- `user_id` - For efficient user like queries
- `frarticle_id` - For article like lookups
