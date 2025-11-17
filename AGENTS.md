# AI Agent Guidelines

## Project Overview
÷«
Mission Statement: A viral news app that helps Americans stay informed about what their government is doing by publishing live updates on everything the Federal government is doing.

How We Do It: We use the Federal Register API to scrape data, the Grok API to process the data, and then display it on a feed page.

Refer to TODO.md for tasks.

## Roadmap

1. ✅ Basic Federal register scraping + Grok AI processing + website to create viral buzz
2. 🚧 **PHASE 2 (CURRENT)**: User accounts with Google OAuth authentication
3. 🔮 Allow likes, comments, and shares (Phase 3, do not implement yet)
4. 🔮 Create the opengov GaaS product (Phase 4, do not implement yet)

**Current Status:** Phase 2 - Implementing Google OAuth authentication with JWT tokens for user accounts. Authentication system includes login, token management, and protected endpoints.

## Tech Stack

- **Backend**: FastAPI + SQLite + SQLAlchemy
- **Frontend**: React + Vite + TypeScript
- **External APIs**: Federal Register API, Grok grok-4-fast API
- **Dependency Management**: `uv` (Python) and `npm` (Node)

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
│   ├── pyproject.toml             # Python dependencies (uv)
│   ├── uv.lock                    # Locked dependency versions
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
├── docs/
│   ├── model.md                   # Data model documentation (MUST be kept in sync)
│   └── style.md                   # UI technical style guide
├── Makefile
├── AGENTS.md
├── CLAUDE.md -> AGENTS.md
├── .gitignore
└── README.md
```

## API Structure

### Authentication Endpoints
**[PHASE 2 - IMPLEMENTED]**
- `GET /api/auth/google/login` - Initiate Google OAuth flow
- `GET /api/auth/google/callback` - Handle Google OAuth callback
- `POST /api/auth/renew` - Renew JWT access token
- `POST /api/auth/logout` - Logout (client-side token removal)
- `GET /api/auth/me` - Get current authenticated user info

### Feed Endpoints
- `GET /api/feed` - Get paginated list of articles/blurbs
  - Query params: `page`, `limit`, `sort` (newest, trending, etc.)
- `GET /api/feed/{article_id}` - Get specific article details
- `POST /api/feed/{article_id}/share` - Track article sharing (analytics)

### User Endpoints
**[PHASE 3 - NOT YET IMPLEMENTED]**
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

## Implementation Guidelines

### Backend
- Use FastAPI async/await for all endpoints and external calls
- SQLite + SQLAlchemy for data persistence
- **[PHASE 2]** Google OAuth authentication with JWT tokens (stateless, 1-hour expiration)
- Separate service modules for external API integrations
- Environment variables for API keys and configuration
- Background workers for periodic Federal Register scraping

**Key Libraries:** fastapi, uvicorn, sqlalchemy, alembic, httpx, apscheduler, slowapi, tenacity, authlib, python-jose, cachetools

**Authentication:** Google OAuth 2.0 with JWT tokens, automatic token renewal, token caching for performance

### Frontend
- TypeScript throughout
- Component structure by feature
- Zustand for state management
- TanStack Router + Query for routing and data fetching
- shadcn/ui + Tailwind CSS for styling
- Infinite scroll/pagination for feed
- Responsive design with loading states and error boundaries

## Testing

- Tests required for all features
- Backend: pytest
- Frontend: Vitest + React Testing Library
- Mock external API integrations
- Run tests before commits

## Documentation

- `docs/model.md` - Keep data models in sync (SQLAlchemy, Pydantic)
- `docs/style.md` - UI component patterns and design decisions

## Development Rules

- Follow established project structure and patterns
- Keep API integrations functional
- Tests required for all features
- Keep docs/model.md in sync with codebase

## Commands

### Backend (using **uv**)

```bash
# Run these from ./backend/ (cd backend && ...)

# 1. Install / update dependencies + create/use .venv automatically
uv sync                                   # Install all dependencies (including dev)
uv sync --frozen                          # Install without updating uv.lock (perfect for CI)

# 2. Add a new dependency
uv add fastapi uvicorn[standard]          # Production dependency
uv add ruff pytest alembic --dev          # Dev / test dependencies
uv add sqlalchemy psycopg2-binary         # Example regular deps

# 3. Development server
uv run dev                                # Run development server

# 4. Linting & formatting
uv run lint                               # Check the whole project
uv run fix                                # Auto-fix and format code

# 5. Tests
uv run test                               # Run all tests
uv run pytest -x --ff                     # Stop on first failure + run failed first
uv run pytest --cov=app                   # With coverage

# 6. Database migrations (Alembic)
alembic revision --autogenerate -m "description"   # Create new migration
alembic upgrade head                                   # Apply migrations to latest
alembic downgrade -1                                   # Rollback last migration
alembic current                                        # Show current revision
```

### Frontend
```bash
cd frontend
npm run dev        # Dev server
npm install        # Dependencies
npm test           # Tests
npm run build      # Production build
```

### Makefile
- `make build` - Build project
- `make run` - Run development environment
- `make deploy` - Deploy application
