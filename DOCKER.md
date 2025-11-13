# Docker Setup Guide

## Prerequisites

- Docker Desktop installed and running (or Docker + Docker Compose on Linux)
- GROK_API_KEY environment variable set

## Quick Start

### 1. Set up environment variables

```bash
cp backend/.env.example backend/.env
# Edit backend/.env and add your GROK_API_KEY
export GROK_API_KEY="your-api-key-here"
```

### 2. Build and start containers

```bash
docker-compose build
docker-compose up
```

### 3. Access the application

- **Frontend**: http://localhost
- **Backend API**: http://localhost:8000
- **API docs**: http://localhost:8000/docs

### 4. Verify services are healthy

```bash
docker-compose ps
```

All services should show `healthy` status after ~30 seconds.

## Available Services

### Backend (FastAPI)
- Port: 8000
- Health check: GET http://localhost:8000/health
- Swagger docs: http://localhost:8000/docs

### Frontend (React + Vite)
- Port: 80
- Health check: GET http://localhost/health
- Routes all API requests to backend via nginx proxy

## Environment Variables

### Backend
Set in `backend/.env`:
- `GROK_API_KEY` - Required for article summarization
- `ENVIRONMENT` - Set to "production" in docker
- `DEBUG` - Set to "false" in production
- `DATABASE_URL` - SQLite path (auto-configured)
- `FEDERAL_REGISTER_API_URL` - Federal Register API endpoint
- `SCRAPER_INTERVAL_MINUTES` - Scraper frequency (default: 15)

### Frontend
Set in `frontend/.env`:
- `VITE_API_URL` - Backend API URL (auto-configured to http://localhost:8000)

## Common Commands

```bash
# Start containers
docker-compose up -d

# Stop containers
docker-compose down

# View logs
docker-compose logs -f

# View specific service logs
docker-compose logs -f backend
docker-compose logs -f frontend

# Run migrations
docker-compose exec backend alembic upgrade head

# Restart a service
docker-compose restart backend
docker-compose restart frontend

# Remove all containers and volumes
docker-compose down -v
```

## Troubleshooting

### Backend won't start
1. Check GROK_API_KEY is set: `echo $GROK_API_KEY`
2. View logs: `docker-compose logs backend`
3. Ensure database migrations run: `docker-compose exec backend alembic upgrade head`

### Frontend shows blank page
1. Check nginx is serving files: `docker-compose logs frontend`
2. Verify backend is accessible from frontend container
3. Check VITE_API_URL is correct in environment

### Services not healthy
- Wait ~30 seconds for services to fully start
- Check `docker-compose ps` status
- View logs for specific errors

### Database issues
- The database persists at `./backend/opengov.db`
- To reset: `rm backend/opengov.db` then restart container
- Migrations run automatically on startup

## Production Deployment

For production deployment:

1. Use environment-specific configurations
2. Add SSL/TLS with reverse proxy (Nginx, Traefik)
3. Use PostgreSQL instead of SQLite
4. Configure proper backup strategies
5. Set up monitoring and logging
6. Use Docker secrets for sensitive environment variables

Refer to `docs/deployment.md` for detailed production setup.
