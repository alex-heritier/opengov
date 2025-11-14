FROM python:3.11-slim

WORKDIR /app

# Install system dependencies and uv
RUN apt-get update && \
    apt-get install -y curl && \
    curl -LsSf https://astral.sh/uv/install.sh | sh && \
    rm -rf /var/lib/apt/lists/*

# Add uv to PATH
ENV PATH="/root/.cargo/bin:$PATH"

# Copy dependency files
COPY backend/pyproject.toml /app/pyproject.toml
COPY backend/uv.lock /app/uv.lock

# Install dependencies with uv
RUN uv sync --frozen --no-dev

# Copy application code
COPY backend/app /app/app
COPY backend/migrations /app/migrations
COPY backend/alembic.ini /app/alembic.ini

# Create database directory
RUN mkdir -p /app/data

# Run migrations and start server
CMD ["sh", "-c", "uv run alembic upgrade head && uv run uvicorn app.main:app --host 0.0.0.0 --port 8000"]
