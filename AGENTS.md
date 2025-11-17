# AI Agent Guidelines

## Project Overview

Mission Statement: A viral news app that helps Americans stay informed about what their government is doing by publishing live updates on everything the Federal government is doing.

How We Do It: We use the Federal Register API to scrape data, the Grok API to process the data, and then display it on a feed page.

Refer to TODO.md for tasks.

## Tech Stack

- **Backend**: FastAPI + SQLite + SQLAlchemy
- **Frontend**: React + Vite + TypeScript
- **External APIs**: Federal Register API, Grok API
- **Dependency Management**: `uv` (Python) and `npm` (Node)

## Project Structure

```
opengov/
├── backend/
│   ├── app/                        # FastAPI application code
│   │   ├── main.py                 # App entry point
│   │   ├── config.py
│   │   ├── database.py
│   │   ├── models/
│   │   ├── services/
│   │   ├── routers/
│   │   ├── schemas/
│   │   └── workers/
│   ├── migrations/                 # Alembic database migrations
│   ├── tests/                      # Test suite
│   ├── pyproject.toml              # Python dependencies & config
│   ├── alembic.ini
│   ├── pytest.ini
│   └── .env.example
│
├── frontend/
│   ├── src/
│   │   ├── main.tsx
│   │   ├── App.tsx
│   │   ├── components/
│   │   ├── pages/
│   │   ├── stores/
│   │   ├── api/
│   │   └── styles/
│   ├── package.json
│   ├── vite.config.ts
│   ├── tsconfig.json
│   └── .env.example
│
├── docs/                           # Documentation (model.md, api.md, auth.md, etc.)
├── Makefile
└── TODO.md
```

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
- Separate service modules for external API integrations
- Environment variables for API keys and configuration
- Background workers for periodic jobs

**Authentication:** fastapi-users with Google OAuth 2.0 fields, JWT tokens in HTTP-only cookies, email/password login

### Frontend
- TypeScript throughout
- Zustand for state management
- TanStack Router + Query for routing and data fetching
- shadcn/ui + Tailwind CSS for styling
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
# Run these from ./frontend/ (cd frontend && ...)

npm run dev        # Dev server
npm install        # Dependencies
npm test           # Tests
npm run build      # Production build
```

### Makefile
- `make build` - Build project
- `make run` - Run development environment
- `make deploy` - Deploy application
- `make test` - Run tests
