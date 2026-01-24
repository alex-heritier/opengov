# Backend Rules

## Project Structure

```
backend/
├── cmd/
│   └── server/                   # Application entry point
├── internal/
│   ├── config/                   # Configuration
│   ├── db/                       # Database connection
│   ├── handlers/                 # HTTP handlers
│   ├── middleware/               # Auth middleware
│   ├── models/                   # Data models
│   ├── repository/               # Data access layer
│   └── services/                 # Business logic & external APIs
├── bin/                          # Compiled binaries
├── scripts/                      # Utility scripts
├── go.mod
└── go.sum
```

## Implementation Guidelines

- Use standard library + Gin for HTTP routing
- PostgreSQL + database/sql for data persistence
- Separate service modules for external API integrations
- Environment variables for API keys and configuration
- Background goroutines for periodic jobs

**Authentication:** JWT Bearer tokens in Authorization header, email/password login + Google OAuth

**Auth Flow:**
- Login/Register: Returns `{access_token, user}` in JSON response
- OAuth: Redirects to `/auth/callback#access_token=<token>`
- API requests: `Authorization: Bearer <token>` header required

## Architecture layer overview

Follow this general layer pattern: Handler -> Service -> Repository -> DB

Prefer embedding or composing structs as to avoid schema duplication. Ex. LoginResponse should embed a User struct NOT duplicate the same fields as User.

## Testing

- Tests required for all features
- Backend: built-in `go test`
- Mock external API integrations
- Run tests before commits

## Documentation

- `docs/model.md` - Keep data models in sync

## Commands

### Installation
- `make install-backend` - Install Go dependencies

### Development
- `make dev-backend` - Start backend dev server

### Testing
- `make test-backend` - Run all backend tests

### Build
- `make build` - Build backend for production
