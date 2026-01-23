# AI Agent Guidelines

## Project Overview

Mission Statement: A viral news app that helps Americans stay informed about what their government is doing by publishing live updates on everything the Federal government is doing.

How We Do It: We use the Federal Register API to scrape data, the Grok API to process the data, and then display it on a feed page.

Refer to TODO.md for tasks.

## Tech Stack

- **Backend**: Go + Gin + SQLite
- **Frontend**: React + Vite + TypeScript
- **External APIs**: Federal Register API, Grok API
- **Dependency Management**: `go mod` (Go) and `bun` (Node)

## Project Structure

```
opengov/
├── backend/          # Go API server
├── frontend/         # React web app
├── docs/             # Documentation
├── Makefile
└── TODO.md
```

## Development Rules

- Follow established project structure and patterns
- Keep API integrations functional
- Tests required for all features
- Keep docs/model.md in sync with codebase

## Commands

Run all commands from the project root. Use `make help` to display all available commands.

### Installation
- `make install` - Install all dependencies
- `make install-backend` - Install Go dependencies only
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
- `make test-frontend` - Run frontend tests

### Build & Cleanup
- `make build` - Build backend and frontend for production
- `make clean` - Clean up generated files and caches
