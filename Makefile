.PHONY: help install install-backend install-frontend \
        dev dev-backend dev-frontend \
        lint fix \
        test test-backend test-backend-fast test-backend-coverage test-frontend \
        db-migrate db-upgrade db-downgrade db-current \
        build clean

help:
	@echo "OpenGov Development Commands"
	@echo ""
	@echo "INSTALLATION"
	@echo "  make install              - Install all dependencies"
	@echo "  make install-backend      - Install Python dependencies"
	@echo "  make install-frontend     - Install Node dependencies"
	@echo ""
	@echo "DEVELOPMENT"
	@echo "  make dev                  - Start both backend & frontend servers"
	@echo "  make dev-backend          - Start backend dev server (localhost:8000)"
	@echo "  make dev-frontend         - Start frontend dev server (localhost:5173)"
	@echo ""
	@echo "CODE QUALITY"
	@echo "  make lint                 - Check code style (backend)"
	@echo "  make fix                  - Auto-fix code style (backend)"
	@echo ""
	@echo "TESTING"
	@echo "  make test                 - Run all tests (backend + frontend)"
	@echo "  make test-backend         - Run all backend tests"
	@echo "  make test-backend-fast    - Run backend tests (stop on first failure)"
	@echo "  make test-backend-coverage- Run backend tests with coverage report"
	@echo "  make test-frontend        - Run frontend tests"
	@echo ""
	@echo "DATABASE"
	@echo "  make db-migrate msg=DESC  - Create new migration"
	@echo "  make db-upgrade           - Apply all pending migrations"
	@echo "  make db-downgrade         - Rollback last migration"
	@echo "  make db-current           - Show current migration"
	@echo ""
	@echo "BUILD & CLEANUP"
	@echo "  make build                - Build frontend for production"
	@echo "  make clean                - Clean up generated files"

# Installation
install: install-backend install-frontend
	@echo "✓ Installation complete"

install-backend:
	@echo "Installing Python dependencies..."
	cd backend && uv sync

install-frontend:
	@echo "Installing Node dependencies..."
	cd frontend && npm install

# Development
dev: dev-backend dev-frontend

dev-backend:
	@echo "Starting backend dev server..."
	@echo "Backend: http://localhost:8000"
	cd backend && uv run dev

dev-frontend:
	@echo "Starting frontend dev server..."
	@echo "Frontend: http://localhost:5173"
	cd frontend && npm run dev

# Code Quality
lint:
	@echo "Checking code style..."
	cd backend && uv run lint

fix:
	@echo "Auto-fixing code style..."
	cd backend && uv run fix

# Testing
test: test-backend test-frontend

test-backend:
	@echo "Running backend tests..."
	cd backend && uv run pytest

test-backend-fast:
	@echo "Running backend tests (fast mode - stop on first failure)..."
	cd backend && uv run pytest -x --ff

test-backend-coverage:
	@echo "Running backend tests with coverage..."
	cd backend && uv run pytest --cov=app

test-frontend:
	@echo "Running frontend tests..."
	cd frontend && npm test -- --run

# Database
db-migrate:
	@echo "Creating new migration..."
	cd backend && alembic revision --autogenerate -m "$(msg)"

db-upgrade:
	@echo "Applying migrations..."
	cd backend && alembic upgrade head

db-downgrade:
	@echo "Reverting last migration..."
	cd backend && alembic downgrade -1

db-current:
	@echo "Current migration:"
	cd backend && alembic current

# Build & Cleanup
build:
	@echo "Building frontend for production..."
	cd frontend && npm run build
	@echo "✓ Build complete: frontend/dist/"

clean:
	@echo "Cleaning up..."
	find . -type d -name __pycache__ -exec rm -rf {} + 2>/dev/null || true
	find . -type f -name "*.pyc" -delete
	rm -rf backend/.pytest_cache backend/htmlcov
	rm -rf frontend/dist frontend/node_modules/.vite
	@echo "✓ Clean complete"
