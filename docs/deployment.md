# Deployment Guide

This guide covers deploying OpenGov to DigitalOcean droplets.

## Prerequisites

- DigitalOcean account with droplets (2 recommended: 1 for backend, 1 for frontend)
- Ubuntu 22.04 LTS droplets ($4-6/month)
- Python 3.11+ and Node.js 18+ installed
- Domain name with DNS configured

## Quick Start

### 1. Set Up Droplets

Create two droplets on DigitalOcean:
- **Backend Droplet**: 1GB RAM, $4/month (backend + database)
- **Frontend Droplet**: 512MB RAM, $4/month (nginx + React)

Or use a single droplet for both services.

### 2. Install Dependencies

SSH into droplet and install Node.js and Python:

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install Python
sudo apt install python3.11 python3.11-venv python3-pip -y

# Install Node.js
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt install nodejs -y

# Install uv (Python package manager)
curl -LsSf https://astral.sh/uv/install.sh | sh
```

### 3. Deploy Code

Clone repository:
```bash
git clone https://github.com/yourusername/opengov.git
cd opengov
```

### 4. Backend Setup

```bash
cd backend

# Install dependencies
uv sync

# Create environment file
cp .env.example .env
nano .env
```

Set required variables in `backend/.env`:
```
GROK_API_KEY=your_grok_api_key_here
DATABASE_URL=sqlite:///./opengov.db
DEBUG=false
ENVIRONMENT=production
```

Start backend server:
```bash
# Using systemd for production (recommended)
sudo tee /etc/systemd/system/opengov-backend.service > /dev/null <<EOF
[Unit]
Description=OpenGov Backend
After=network.target

[Service]
Type=simple
User=ubuntu
WorkingDirectory=/home/ubuntu/opengov/backend
Environment="PATH=/home/ubuntu/.cargo/bin"
ExecStart=/home/ubuntu/.local/bin/uv run uvicorn app.main:app --host 0.0.0.0 --port 8000
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable opengov-backend
sudo systemctl start opengov-backend
```

### 5. Frontend Setup

```bash
cd ../frontend

# Install dependencies
npm install

# Build for production
npm run build

# Install nginx
sudo apt install nginx -y

# Create nginx configuration
sudo tee /etc/nginx/sites-available/opengov > /dev/null <<'EOF'
server {
    listen 80;
    server_name yourdomain.com www.yourdomain.com;
    
    location / {
        root /home/ubuntu/opengov/frontend/dist;
        try_files $uri $uri/ /index.html;
        expires 365d;
        add_header Cache-Control "public, immutable";
    }
    
    location /api {
        proxy_pass http://localhost:8000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
EOF

sudo ln -s /etc/nginx/sites-available/opengov /etc/nginx/sites-enabled/
sudo systemctl restart nginx
```

## Production Setup

### Option 1: Single Droplet (Simple)

Run both frontend and backend on one droplet.

Pros: Simple, low cost ($4/month)
Cons: Both services share resources

### Option 2: Separate Droplets (Recommended)

**Backend Droplet** (1GB RAM, $4/month):
```bash
# Follow steps 2 & 4 above
# Expose port 8000
```

**Frontend Droplet** (512MB RAM, $4/month):
```bash
# Follow steps 2 & 5 above
# Update backend API URL in nginx config to point to backend droplet IP
```

## SSL/HTTPS Setup

### Using Let's Encrypt + Certbot

```bash
# Install certbot
sudo apt install certbot python3-certbot-nginx -y

# Obtain certificate
sudo certbot certonly --nginx -d yourdomain.com -d www.yourdomain.com

# Auto-renew (runs daily)
sudo systemctl enable certbot.timer
sudo systemctl start certbot.timer
```

Update nginx config to use SSL:
```bash
server {
    listen 443 ssl http2;
    server_name yourdomain.com www.yourdomain.com;
    
    ssl_certificate /etc/letsencrypt/live/yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/yourdomain.com/privkey.pem;
    
    # ... rest of config
}

# Redirect HTTP to HTTPS
server {
    listen 80;
    server_name yourdomain.com www.yourdomain.com;
    return 301 https://$server_name$request_uri;
}
```

## Database Configuration

### SQLite (Current - Good for MVP)

File-based database stored on droplet:
```bash
# Create backup script
cat > /home/ubuntu/backup.sh <<EOF
#!/bin/bash
sqlite3 /home/ubuntu/opengov/backend/opengov.db ".backup '/backups/opengov-\$(date +%Y%m%d).db'"
EOF

chmod +x /home/ubuntu/backup.sh
```

Add to crontab for daily backups:
```bash
0 2 * * * /home/ubuntu/backup.sh
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
3. Restart backend service:
   ```bash
   sudo systemctl restart opengov-backend
   ```

## Monitoring & Logs

### View Logs

```bash
# Backend logs
sudo journalctl -u opengov-backend -f

# Nginx logs
sudo tail -f /var/log/nginx/access.log
sudo tail -f /var/log/nginx/error.log
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

For SQLite:
```bash
0 2 * * * /home/ubuntu/backup.sh
```

For PostgreSQL via DigitalOcean:
```bash
# Use DO managed database automated backups (free with database)
# Or use pg_dump:
0 2 * * * pg_dump postgresql://user:pass@db.ondigitalocean.com/opengov | gzip > /backups/opengov-$(date +\%Y\%m\%d).sql.gz
```

## Updating Code

Pull updates and restart:
```bash
cd /home/ubuntu/opengov
git pull origin main

# Backend
cd backend && uv sync
sudo systemctl restart opengov-backend

# Frontend
cd ../frontend && npm install && npm run build
sudo systemctl restart nginx
```

## Cost Estimate

- Backend Droplet: $4/month
- Frontend Droplet: $4/month (optional)
- Domain: ~$10/year
- Managed Database: ~$15/month (if using PostgreSQL)
- **Total: ~$8-12/month for SQLite setup, ~$23/month for PostgreSQL**

## Performance Tips

1. **Caching**: Frontend nginx configured with 365d cache for assets
2. **Compression**: Gzip enabled for all text content
3. **Database**: Indexes on `published_at` for fast queries
4. **Rate Limiting**: Configure in FastAPI as needed

## Troubleshooting

### Backend won't start

```bash
sudo journalctl -u opengov-backend -n 50
```

Check that all environment variables are set and dependencies are installed.

### Out of memory

Check resource usage:
```bash
free -h
ps aux | sort -k 3 -r | head
```

Upgrade droplet or optimize queries.

### Database locked

For SQLite:
```bash
# Restart backend service
sudo systemctl restart opengov-backend
```

For PostgreSQL:
```bash
# Check managed DB console in DigitalOcean
```

### Nginx 502 Bad Gateway

Backend is down or not responding on port 8000:
```bash
sudo systemctl status opengov-backend
sudo journalctl -u opengov-backend -n 20
```

## Security Checklist

- [ ] Set `DEBUG=false`
- [ ] Use strong `GROK_API_KEY`
- [ ] Enable SSL/HTTPS
- [ ] Configure firewall (allow only 80, 443, 22)
- [ ] Regular backups
- [ ] Keep packages updated: `sudo apt update && sudo apt upgrade`
- [ ] Monitor logs for suspicious activity
- [ ] Use environment variables (never hardcode secrets)
- [ ] Set up strong SSH keys (disable password auth)

## Next Steps

1. Deploy to test droplet first
2. Monitor performance and logs
3. Set up automated backups
4. Configure custom domain + SSL
5. Monitor with DigitalOcean alerts
