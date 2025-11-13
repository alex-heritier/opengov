# AI Agent Guidelines

## Overview

This is a government as a service (GaaS) project for corrupt countries with bad governance.

## Roadmap

1. Federal register scraper to create a viral buzz around us
2. Create the opengov GaaS product

## Tech Stack

- **Backend**: FastAPI
- **Frontend**: React with Vite
- **APIs**:
  - Federal Register API (periodically polled)
  - Grok's grok-4-fast API (for summarization and analysis)

## How It Works

The backend periodically hits the Federal Register API and runs the results through Grok's grok-4-fast API to summarize and analyze the register details. It then creates interesting, digestible little blurbs and articles on the `/feed` page for people to read and share.

## Project Structure

```
opengov/
├── backend/
│   ├── app/
│   │   ├── __init__.py
│   │   ├── main.py                 # FastAPI app entry point
│   │   ├── config.py               # Configuration and environment variables
│   │   ├── database.py             # SQLAlchemy setup and session management
│   │   ├── models/                 # SQLAlchemy models
│   │   │   ├── __init__.py
│   │   │   ├── user.py
│   │   │   ├── article.py
│   │   │   └── federal_register.py
│   │   ├── services/               # Business logic and external API integrations
│   │   │   ├── __init__.py
│   │   │   ├── federal_register.py # Federal Register API client
│   │   │   ├── grok.py            # Grok API client
│   │   │   └── auth.py            # Google OAuth logic
│   │   ├── routers/               # API endpoints
│   │   │   ├── __init__.py
│   │   │   ├── auth.py            # Authentication routes
│   │   │   ├── feed.py            # Feed/articles routes
│   │   │   └── users.py
│   │   ├── schemas/               # Pydantic schemas
│   │   │   ├── __init__.py
│   │   │   ├── user.py
│   │   │   └── article.py
│   │   └── workers/               # Background tasks
│   │       ├── __init__.py
│   │       └── scraper.py         # Periodic Federal Register scraper
│   ├── requirements.txt
│   ├── .env.example
│   └── opengov.db                 # SQLite database file
│
├── frontend/
│   ├── src/
│   │   ├── main.tsx               # App entry point
│   │   ├── App.tsx
│   │   ├── routes/                # TanStack Router routes
│   │   │   ├── __root.tsx
│   │   │   ├── index.tsx
│   │   │   └── feed.tsx
│   │   ├── components/
│   │   │   ├── ui/                # shadcn components
│   │   │   ├── feed/              # Feed-specific components
│   │   │   │   ├── ArticleCard.tsx
│   │   │   │   └── FeedList.tsx
│   │   │   ├── auth/              # Auth-specific components
│   │   │   │   └── GoogleLogin.tsx
│   │   │   └── layout/
│   │   │       ├── Header.tsx
│   │   │       └── Footer.tsx
│   │   ├── stores/                # Zustand stores
│   │   │   ├── authStore.ts
│   │   │   └── feedStore.ts
│   │   ├── api/                   # TanStack Query hooks and API calls
│   │   │   ├── client.ts          # Axios/fetch client
│   │   │   ├── queries.ts         # Query hooks
│   │   │   └── mutations.ts       # Mutation hooks
│   │   ├── lib/                   # Utilities
│   │   │   └── utils.ts
│   │   └── styles/
│   │       └── globals.css
│   ├── public/
│   ├── index.html
│   ├── package.json
│   ├── vite.config.ts
│   ├── tsconfig.json
│   ├── components.json            # shadcn config
│   └── .env.example
│
├── Makefile
├── AGENTS.md
├── CLAUDE.md -> AGENTS.md
├── .gitignore
└── README.md
```

## API Structure

### Authentication Endpoints
- `POST /api/auth/google/login` - Initiate Google OAuth flow
- `GET /api/auth/google/callback` - Handle Google OAuth callback
- `POST /api/auth/logout` - Logout current user
- `GET /api/auth/me` - Get current authenticated user info

### Feed Endpoints
- `GET /api/feed` - Get paginated list of articles/blurbs
  - Query params: `page`, `limit`, `sort` (newest, trending, etc.)
- `GET /api/feed/{article_id}` - Get specific article details
- `POST /api/feed/{article_id}/share` - Track article sharing (analytics)

### User Endpoints
- `GET /api/users/me` - Get current user profile
- `PATCH /api/users/me` - Update user profile settings

### Admin Endpoints (Optional)
- `POST /api/admin/scrape` - Manually trigger Federal Register scrape
- `GET /api/admin/articles` - Get all articles with metadata for admin dashboard

### Response Format

All API responses follow this structure:
```json
{
  "success": true,
  "data": { ... },
  "error": null
}
```

Error responses:
```json
{
  "success": false,
  "data": null,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable error message"
  }
}
```

## Implementation Notes

### Backend

- Use FastAPI's async/await for all API endpoints and external calls
- Implement periodic tasks with background workers (e.g., APScheduler or Celery)
- Use SQLite as the database with SQLAlchemy ORM for data persistence
- Store Federal Register data and Grok summaries in the database
- Implement Google OAuth for user authentication and login
- Create separate service modules for Federal Register API and Grok API integrations
- Include proper error handling and retry logic for external API calls
- Use environment variables for API keys and configuration
- Add rate limiting to prevent API quota exhaustion

### Frontend

- Structure components by feature (e.g., `/components/feed`, `/components/article`)
- Use Zustand for global state management
- Use TanStack Router for routing and navigation
- Use TanStack Query for data fetching, caching, and server state management
- Use shadcn/ui components for consistent UI design
- Implement infinite scroll or pagination for the `/feed` page
- Make articles shareable with meta tags for social media previews
- Use Vite's environment variables for API endpoint configuration
- Keep UI responsive and mobile-friendly
- Add loading states and error boundaries for better UX

## Development Guidelines

When working on this project:
- Maintain the FastAPI backend structure
- Follow React/Vite conventions for frontend development
- Ensure API integrations remain functional
- Keep the feed content engaging and digestible

### Makefile Commands

Use the Makefile for common development tasks:
- `make build` - Build the project (frontend and backend)
- `make run` - Run the development environment
- `make deploy` - Deploy the application
