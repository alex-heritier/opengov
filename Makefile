.PHONY: help \
        install install-backend install-frontend \
        dev dev-backend dev-frontend \
        test test-backend test-frontend \
        lint lint-backend lint-frontend \
        fix fix-backend fix-frontend \
        build clean

help:
	@echo "OpenGov Development Commands"
	@echo ""
	@echo "INSTALLATION"
	@echo "  make install              - Install all dependencies"
	@echo "  make install-backend      - Install Go dependencies"
	@echo "  make install-frontend     - Install Node dependencies"
	@echo ""
	@echo "DEVELOPMENT"
	@echo "  make dev                  - Start both backend & frontend servers"
	@echo "  make dev-backend          - Start backend dev server (localhost:8000)"
	@echo "  make dev-frontend         - Start frontend dev server (localhost:5173)"
	@echo ""
	@echo "TESTING"
	@echo "  make test                 - Run all tests (backend + frontend)"
	@echo "  make test-backend         - Run all backend tests"
	@echo "  make test-frontend        - Run frontend tests"
	@echo ""
	@echo "CODE QUALITY"
	@echo "  make lint                 - Check code style"
	@echo "  make fix                  - Auto-fix code style and format"
	@echo ""
	@echo "BUILD & CLEANUP"
	@echo "  make build                - Build backend and frontend for production"
	@echo "  make clean                - Clean up generated files"

install: install-backend install-frontend
	@echo "✓ Installation complete"

install-backend:
	@echo "Installing Go dependencies..."
	cd backend && make deps
	@echo "Installing air (live reload)..."
	cd backend && make install-air

install-frontend:
	@echo "Installing Node dependencies..."
	cd frontend && bun install

dev: dev-backend dev-frontend

dev-backend:
	@echo "Starting backend dev server with auto-reload..."
	@echo "Backend: http://localhost:8000"
	cd backend && make dev

dev-frontend:
	@echo "Starting frontend dev server..."
	@echo "Frontend: http://localhost:5173"
	cd frontend && bun run dev

test: test-backend test-frontend

test-backend:
	@echo "Running backend tests..."
	cd backend && make test

test-frontend:
	@echo "Running frontend tests..."
	cd frontend && bun run test

build:
	@echo "Building backend..."
	cd backend && make build
	@echo "Building frontend..."
	cd frontend && bun run build
	@echo "✓ Build complete: backend/bin/, frontend/dist/"

clean:
	@echo "Cleaning up..."
	cd backend && make clean
	rm -rf frontend/dist frontend/node_modules/.vite
	@echo "✓ Clean complete"

lint: lint-backend lint-frontend

lint-backend:
	@echo "Linting backend..."
	cd backend && make lint

lint-frontend:
	@echo "Linting frontend..."
	cd frontend && bun run lint

fix: fix-backend fix-frontend

fix-backend:
	@echo "Fixing backend..."
	cd backend && make fmt

fix-frontend:
	@echo "Fixing frontend..."
	cd frontend && bun run fix
