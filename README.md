# OpenGov

A viral news app that helps Americans stay informed about what their government is doing by publishing live updates on everything the Federal government is doing.

Future mission: A fully-featured policy intelligence platform to help companies and organizations make better decisions based on government activity. There will also be a customer facing viral news app that will be used to increase brand awareness and engagement.

## Mission

OpenGov publishes real-time updates from the Federal Register with AI-powered summaries to make government activities accessible and shareable.

## Tech Stack

- **Backend**: Go + Gin + PostgreSQL
- **Frontend**: React + Vite + TypeScript + TanStack Router/Query
- **External APIs**: Federal Register API, Grok (xAI)

## Prerequisites

- Go 1.25+
- Node.js 18+
- PostgreSQL 14+
- npm

## Installation

```bash
# Install Go dependencies
cd backend
go mod download
go mod tidy

# Install frontend dependencies
cd ../frontend
npm install
```

## Development

```bash
# Setup database
createdb opengov
psql opengov -f migrations/001_initial_schema.sql

# Set environment variables
cp backend/.env.example backend/.env
# Edit backend/.env with your configuration

# Start backend (from backend directory)
make run

# Start frontend (from frontend directory)
npm run dev

# Backend: http://localhost:8000
# Frontend: http://localhost:5173
```

## Testing

```bash
# Backend tests
cd backend
make test

# Frontend tests
cd frontend
npm test
```

## Building

```bash
# Build backend
cd backend
make build

# Build frontend
cd frontend
npm run build
```

## Project Structure

```
opengov/
├── backend/                    # Go application
│   ├── cmd/server/            # Application entry point
│   ├── internal/
│   │   ├── config/           # Configuration loading
│   │   ├── db/               # Database connection
│   │   ├── handlers/         # HTTP handlers
│   │   ├── middleware/       # Auth, logging middleware
│   │   ├── models/           # Data models
│   │   ├── repository/       # Data access layer
│   │   └── services/         # External API clients
│   ├── migrations/           # SQL migrations
│   └── Makefile
│
├── frontend/                   # React + Vite application
│   ├── src/
│   │   ├── main.tsx          # Entry point
│   │   ├── api/              # API integration
│   │   ├── components/       # React components
│   │   ├── stores/           # Zustand state
│   │   └── styles/           # Tailwind CSS
│   └── package.json
│
└── docs/
    ├── model.md              # Database schema
    ├── api.md                # API documentation
    └── style.md              # UI style guide
```

## API Endpoints

### Health
- `GET /health` - Health check

### Auth
- `POST /api/auth/login` - Login
- `POST /api/auth/register` - Register
- `GET /api/auth/me` - Get current user
- `POST /api/auth/refresh` - Refresh token

### Feed
- `GET /api/feed` - Get paginated articles
- `GET /api/feed/:id` - Get article by ID
- `GET /api/feed/document/:document_number` - Get article by document number

### Bookmarks
- `GET /api/bookmarks` - Get user bookmarks
- `POST /api/bookmarks/:article_id` - Toggle bookmark

### Likes
- `GET /api/likes/:article_id` - Get like counts
- `POST /api/likes/:article_id` - Toggle like

## Configuration

### Backend Environment Variables
Create `backend/.env`:
```
DATABASE_URL=postgres://localhost/opengov?sslmode=disable
JWT_SECRET_KEY=your-secret-key-min-32-chars
GROK_API_KEY=your-key-here
```

### Frontend Environment Variables
Create `frontend/.env.local`:
```
VITE_API_URL=http://localhost:8000
```

## Roadmap

**Phase 1 (Current)**: Basic Federal Register scraping + Grok AI processing + website to create viral buzz

**Phase 2 (Future)**: User accounts, likes, comments, and shares

**Phase 3 (Future)**: OpenGov GaaS product

## Contributing

Follow the patterns established in AGENTS.md.

## License

MIT
