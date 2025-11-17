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

Run all commands from the project root. Use `make help` to display all available commands.

### Installation
- `make install` - Install all dependencies
- `make install-backend` - Install Python dependencies only
- `make install-frontend` - Install Node dependencies only

### Development
- `make dev` - Start both backend (localhost:8000) and frontend (localhost:5173) servers
- `make dev-backend` - Start backend dev server
- `make dev-frontend` - Start frontend dev server

### Code Quality
- `make lint` - Check code style
- `make fix` - Auto-fix code style and format

### Testing
- `make test` - Run all tests (backend + frontend)
- `make test-backend` - Run all backend tests
- `make test-backend-fast` - Run backend tests, stop on first failure
- `make test-backend-coverage` - Run backend tests with coverage report
- `make test-frontend` - Run frontend tests

### Database
- `make db-migrate msg="description"` - Create new migration
- `make db-upgrade` - Apply all pending migrations
- `make db-downgrade` - Rollback last migration
- `make db-current` - Show current migration

### Build & Cleanup
- `make build` - Build frontend for production
- `make clean` - Clean up generated files and caches
