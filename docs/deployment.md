# Deployment Guide

This guide covers deploying OpenGov to DigitalOcean droplets using Docker.

## Prerequisites

- DigitalOcean account with droplets (2 recommended: 1 for backend, 1 for frontend)
- Ubuntu 22.04 LTS droplets ($4-6/month)
- Docker and Docker Compose installed
- Domain name with DNS configured

## Quick Start

### 1. Set Up Droplets

Create two droplets on DigitalOcean:
- **Backend Droplet**: 1GB RAM, $4/month (backend + database)
- **Frontend Droplet**: 512MB RAM, $4/month (nginx + React)

Or single droplet with both containers.

### 2. Install Docker

SSH into droplet and run:

```bash
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER
newgrp docker
```

Install Docker Compose:
```bash
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
```

### 3. Deploy Code

Clone repository:
```bash
git clone https://github.com/yourusername/opengov.git
cd opengov
```

Create environment file:
```bash
cp backend/.env.example backend/.env
nano backend/.env
```

Set required variables:
```
GROK_API_KEY=your_grok_api_key_here
DATABASE_URL=sqlite:///./opengov.db
DEBUG=false
ENVIRONMENT=production
```

### 4. Build and Run

Start services:
```bash
docker-compose up -d
```

Verify health:
```bash
curl http://localhost:8000/health
curl http://localhost/health
```

## Production Setup

### Option 1: Single Droplet (Simple)

Run both frontend and backend on one droplet using docker-compose.

Pros: Simple, low cost ($4/month)
Cons: Both services share resources

### Option 2: Separate Droplets (Recommended)

**Backend Droplet** (1GB RAM, $4/month):
```bash
# Only backend service
docker-compose up -d backend
# Expose port 8000
```

**Frontend Droplet** (512MB RAM, $4/month):
```bash
# Only frontend service, point to backend URL
VITE_API_URL=https://api.yourdomain.com docker-compose up -d frontend
```

### Option 3: Load Balanced (Enterprise)

Use DigitalOcean Load Balancer + multiple backend droplets.

## SSL/HTTPS Setup

### Using Let's Encrypt + Certbot

For frontend (nginx):
```bash
docker exec -it opengov-frontend-1 apk add certbot certbot-nginx
certbot certonly --nginx -d yourdomain.com -d www.yourdomain.com
```

For backend (if exposing directly):
```bash
# Better: Use reverse proxy with nginx in front of uvicorn
```

### Using DigitalOcean Spaces + CDN

1. Create Spaces bucket
2. Build frontend: `npm run build`
3. Upload `frontend/dist/*` to Spaces
4. Enable CDN on bucket
5. Point domain to Spaces CDN

## Database Configuration

### SQLite (Current - Good for MVP)

File-based database stored in container volume:
```yaml
volumes:
  - ./backend/opengov.db:/app/opengov.db
```

Pros: No setup, automatic backups with code
Cons: Not scalable beyond ~10K users

### PostgreSQL (Recommended for Scale)

Use DigitalOcean Managed Database:
1. Create managed PostgreSQL in DO console
2. Update environment:
   ```
   DATABASE_URL=postgresql://user:pass@db.ondigitalocean.com:25060/opengov
   ```
3. Run migrations:
   ```bash
   docker-compose exec backend alembic upgrade head
   ```

## Monitoring & Logs

### View Logs

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f backend
docker-compose logs -f frontend
```

### Health Checks

Monitor endpoints:
- Backend: `GET http://api.yourdomain.com/health`
- Backend DB: `GET http://api.yourdomain.com/health/db`
- Frontend: `GET http://yourdomain.com/health`

### Set Up Monitoring

Using DigitalOcean Monitoring:
1. Enable Monitoring on droplets
2. Create alerts for CPU, memory, disk
3. Configure notification to email/Slack

## Backup Strategy

### Automated Daily Backups

```bash
#!/bin/bash
# backup.sh
docker-compose exec -T backend sqlite3 /app/opengov.db ".backup '/backups/opengov-$(date +%Y%m%d).db'"
```

Add to crontab:
```bash
0 2 * * * /home/ubuntu/backup.sh
```

Or use DigitalOcean Spaces:
```bash
docker-compose exec -T backend aws s3 cp /app/opengov.db s3://your-backup-bucket/
```

## Updating Code

Pull updates and restart:
```bash
git pull origin main
docker-compose down
docker-compose up -d --build
docker-compose exec backend alembic upgrade head  # if migrations exist
```

## Cost Estimate

- Backend Droplet: $4/month
- Frontend Droplet: $4/month (optional)
- Domain: ~$10/year
- Backups: Free with git + managed DB
- **Total: ~$8-12/month for complete setup**

## Performance Tips

1. **Caching**: Frontend nginx already configured with 365d cache for assets
2. **Compression**: Gzip enabled for all text content
3. **Database**: Indexes on `published_at` for fast queries
4. **Rate Limiting**: 100 req/min for feed (adjust in config)

## Troubleshooting

### Container won't start

```bash
docker-compose logs backend  # Check error message
docker-compose down -v       # Remove volumes and restart fresh
```

### Out of memory

Check resource usage:
```bash
docker stats
```

Upgrade droplet or optimize queries.

### Database locked

For SQLite:
```bash
# Restart containers
docker-compose restart backend
```

For PostgreSQL:
```bash
# Check managed DB console in DigitalOcean
```

## Security Checklist

- [ ] Set `DEBUG=false`
- [ ] Use strong `GROK_API_KEY`
- [ ] Enable SSL/HTTPS
- [ ] Configure firewall (allow only 80, 443, 22)
- [ ] Regular backups
- [ ] Keep Docker images updated
- [ ] Monitor logs for suspicious activity
- [ ] Use environment variables (never hardcode secrets)

## Next Steps

1. Deploy to test droplet first
2. Monitor performance and logs
3. Set up automated backups
4. Configure custom domain + SSL
5. Monitor with DigitalOcean alerts
