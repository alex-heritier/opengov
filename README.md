# OpenGov

A viral news app that helps Americans stay informed about what their government is doing by publishing live updates on everything the Federal government is doing.

## Mission

OpenGov publishes real-time updates from the Federal Register with AI-powered summaries to make government activities accessible and shareable.

## Tech Stack

- **Backend**: FastAPI + SQLite + SQLAlchemy
- **Frontend**: React + Vite + TypeScript + TanStack Router/Query
- **External APIs**: Federal Register API, Grok (xAI)

## Prerequisites

- Python 3.11+
- Node.js 18+
- pip and npm

## Installation

### Option 1: Docker (Recommended)

```bash
# Build and start all services
docker-compose up --build

# Frontend: http://localhost
# Backend: http://localhost:8000
# API Docs: http://localhost:8000/docs
```

See `DOCKER.md` for detailed Docker setup and troubleshooting.

### Option 2: Local Development

```bash
# Install dependencies for both backend and frontend
make install
```

## Development

```bash
# Start development servers (backend + frontend)
make run

# Backend: http://localhost:8000
# Frontend: http://localhost:5173
# API Docs: http://localhost:8000/docs
```

## Testing

```bash
# Run all tests
make test

# Or individually:
cd backend && pytest
cd frontend && npm test
```

## Building

```bash
# Build frontend for production
make build
```

## Project Structure

```
opengov/
├── backend/                    # FastAPI application
│   ├── app/
│   │   ├── main.py            # App entry point
│   │   ├── config.py          # Configuration
│   │   ├── database.py        # SQLAlchemy setup
│   │   ├── models/            # Data models
│   │   ├── services/          # External API clients
│   │   ├── routers/           # API endpoints
│   │   ├── schemas/           # Pydantic schemas
│   │   └── workers/           # Background tasks
│   ├── tests/
│   └── requirements.txt
│
├── frontend/                   # React + Vite application
│   ├── src/
│   │   ├── main.tsx           # Entry point
│   │   ├── api/               # API integration
│   │   ├── components/        # React components
│   │   ├── stores/            # Zustand state
│   │   └── styles/            # Tailwind CSS
│   └── package.json
│
└── docs/
    ├── model.md               # Database schema
    ├── api.md                 # API documentation
    └── style.md               # UI style guide
```

## Configuration

### Backend Environment Variables
Create `backend/.env`:
```
GROK_API_KEY=your_key_here
FEDERAL_REGISTER_API_URL=https://www.federalregister.gov/api/v1
SCRAPER_INTERVAL_MINUTES=15
SCRAPER_DAYS_LOOKBACK=1
```

### Frontend Environment Variables
Create `frontend/.env.local`:
```
VITE_API_URL=http://localhost:8000
```

## API Documentation

Full API documentation available at `http://localhost:8000/docs` when running the backend.

See `docs/api.md` for detailed endpoint information.

## Database Schema

See `docs/model.md` for complete data model documentation.

## UI Style Guide

See `docs/style.md` for component patterns and design decisions.

## Roadmap

**Phase 1 (Current)**: Basic Federal Register scraping + Grok AI processing + website to create viral buzz

**Phase 2 (Future)**: User accounts, likes, comments, and shares

**Phase 3 (Future)**: OpenGov GaaS product

## Contributing

Follow the patterns established in AGENTS.md.

## License

MIT
