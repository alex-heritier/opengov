.PHONY: help install install-backend install-frontend \
        dev dev-backend dev-frontend \
        lint fix \
        test test-backend test-frontend \
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
	@echo "CODE QUALITY"
	@echo "  make lint                 - Check code style (backend)"
	@echo "  make fix                  - Auto-fix code style (backend)"
	@echo ""
	@echo "TESTING"
	@echo "  make test                 - Run all tests (backend + frontend)"
	@echo "  make test-backend         - Run all backend tests"
	@echo "  make test-frontend        - Run frontend tests"
	@echo ""
	@echo "BUILD & CLEANUP"
	@echo "  make build                - Build backend and frontend for production"
	@echo "  make clean                - Clean up generated files"

# Installation
install: install-backend install-frontend
	@echo "✓ Installation complete"

install-backend:
	@echo "Installing Go dependencies..."
	cd backend && go mod download && go mod tidy
	@echo "Installing air (live reload)..."
	cd backend && go install github.com/air-verse/air@latest

install-frontend:
	@echo "Installing Node dependencies..."
	cd frontend && bun install

# Development
dev: dev-backend dev-frontend

dev-backend:
	@echo "Starting backend dev server with auto-reload..."
	@echo "Backend: http://localhost:8000"
	cd backend && make dev

dev-frontend:
	@echo "Starting frontend dev server..."
	@echo "Frontend: http://localhost:5173"
	cd frontend && bun run dev

# Code Quality
lint:
	@echo "Checking code style..."
	cd backend && go vet ./...

fix:
	@echo "Auto-fixing code style..."
	cd backend && go fmt ./...

# Testing
test: test-backend test-frontend

test-backend:
	@echo "Running backend tests..."
	cd backend && go test -v ./...

test-frontend:
	@echo "Running frontend tests..."
	cd frontend && bun test --run

# Build & Cleanup
build:
	@echo "Building backend..."
	cd backend && go build -o bin/server ./cmd/server
	@echo "Building frontend for production..."
	cd frontend && npm run build
	@echo "✓ Build complete: backend/bin/server, frontend/dist/"

clean:
	@echo "Cleaning up..."
	rm -rf backend/bin backend/coverage.out backend/coverage.html
	rm -rf frontend/dist frontend/node_modules/.vite
	find . -type d -name __pycache__ -exec rm -rf {} + 2>/dev/null || true
	@echo "✓ Clean complete"
