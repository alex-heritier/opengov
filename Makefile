.PHONY: install run test build clean help

help:
	@echo "OpenGov Development Commands"
	@echo "make install  - Install backend & frontend dependencies"
	@echo "make runb    - Start backend dev server"
	@echo "make runf    - Start frontend dev server"
	@echo "make test     - Run backend + frontend tests"
	@echo "make build    - Build frontend for production"
	@echo "make clean    - Clean up generated files"

install:
	@echo "Installing dependencies..."
	cd backend && uv sync
	cd frontend && npm install
	@echo "Installation complete"

runb:
	@echo "Starting backend dev server..."
	@echo "Backend: http://localhost:8000"
	cd backend && uvicorn app.main:app --reload --host 0.0.0.0 --port 8000

runf:
	@echo "Starting frontend dev server..."
	@echo "Frontend: http://localhost:5173"
	cd frontend && npm run dev

test:
	@echo "Running tests..."
	cd backend && pytest
	cd frontend && npm test

build:
	@echo "Building frontend for production..."
	cd frontend && npm run build
	@echo "Build complete: frontend/dist/"

clean:
	@echo "Cleaning up..."
	find . -type d -name __pycache__ -exec rm -rf {} + 2>/dev/null || true
	find . -type f -name "*.pyc" -delete
	rm -rf backend/.pytest_cache backend/htmlcov
	rm -rf frontend/dist frontend/node_modules/.vite
	@echo "Clean complete"
