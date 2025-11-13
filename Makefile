.PHONY: install run test build clean help

help:
	@echo "OpenGov Development Commands"
	@echo "make install  - Install backend & frontend dependencies"
	@echo "make run      - Start backend & frontend dev servers"
	@echo "make test     - Run backend + frontend tests"
	@echo "make build    - Build frontend for production"
	@echo "make clean    - Clean up generated files"

install:
	@echo "Installing dependencies..."
	cd backend && pip install -r requirements.txt
	cd frontend && npm install
	@echo "Installation complete"

run:
	@echo "Starting development servers..."
	@echo "Backend: http://localhost:8000"
	@echo "Frontend: http://localhost:5173"
	@echo "Press Ctrl+C to stop"
	@(cd backend && uvicorn app.main:app --reload --host 0.0.0.0 --port 8000) &
	@(cd frontend && npm run dev) &
	@wait

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
