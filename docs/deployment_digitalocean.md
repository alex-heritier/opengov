# DigitalOcean Deployment Guide

Complete guide for deploying the OpenGov application to a DigitalOcean droplet.

**Stack:**
- **Backend:** FastAPI + SQLite + APScheduler (background scraper)
- **Frontend:** React + Vite (built static files)
- **Web Server:** Nginx (reverse proxy + static file serving)
- **Process Manager:** systemd (for backend service)
- **SSL/TLS:** Let's Encrypt (via Certbot)

---

## Overview

This guide walks through:
1. Creating and configuring a DigitalOcean droplet
2. Installing system dependencies
3. Deploying backend and frontend
4. Configuring nginx as reverse proxy
5. Setting up SSL with Let's Encrypt
6. Creating systemd service for backend
7. Security hardening and monitoring

**Estimated time:** 45-60 minutes

---

## Prerequisites

- DigitalOcean account
- Domain name pointed to your droplet (for SSL)
- SSH key pair (recommended)
- API keys (Grok, Google OAuth, Twitter OAuth)

---

## Phase 1: DigitalOcean Setup

### 1. Create Droplet

**Recommended specs:**
- **Size:** Basic Droplet - $12/month (2 GB RAM, 1 vCPU, 50 GB SSD)
  - For production with higher traffic: $24/month (4 GB RAM, 2 vCPU)
- **OS:** Ubuntu 24.04 LTS x64
- **Datacenter:** Choose region closest to your users
- **Authentication:** SSH keys (recommended) or password
- **Hostname:** `opengov-prod` (or your preference)

**Optional add-ons:**
- ✅ Monitoring (free)
- ❌ Backups (+20% cost, but recommended for production)

### 2. Configure DNS

Point your domain to the droplet:

```
Type: A Record
Host: @ (or subdomain like "app")
Value: <your-droplet-ip>
TTL: 3600
```

For subdomain (e.g., app.yourdomain.com):
```
Type: A Record
Host: app
Value: <your-droplet-ip>
TTL: 3600
```

Wait 5-15 minutes for DNS propagation.

### 3. Initial SSH Access

```bash
# SSH into your droplet
ssh root@<your-droplet-ip>

# If using SSH key:
ssh -i ~/.ssh/your_key root@<your-droplet-ip>
```

---

## Phase 2: Initial Server Setup

### 1. Update System

```bash
# Update package lists and upgrade packages
apt update && apt upgrade -y

# Install essential tools
apt install -y curl wget git ufw fail2ban unzip
```

### 2. Create Non-Root User

```bash
# Create deployment user
adduser opengov

# Add to sudo group
usermod -aG sudo opengov

# Copy SSH keys to new user (if using SSH keys)
rsync --archive --chown=opengov:opengov ~/.ssh /home/opengov
```

### 3. Configure Firewall

```bash
# Allow SSH, HTTP, and HTTPS
ufw allow OpenSSH
ufw allow 'Nginx Full'

# Enable firewall
ufw --force enable

# Check status
ufw status
```

### 4. Switch to Non-Root User

```bash
# Switch to opengov user
su - opengov

# Or reconnect via SSH:
# ssh opengov@<your-droplet-ip>
```

---

## Phase 3: Install Dependencies

### 1. Install Python 3.11+

Ubuntu 24.04 comes with Python 3.12 by default:

```bash
# Verify Python version
python3 --version  # Should be 3.12+

# Install pip and venv
sudo apt install -y python3-pip python3-venv python3-dev
```

### 2. Install uv (Python Package Manager)

```bash
# Install uv
curl -LsSf https://astral.sh/uv/install.sh | sh

# Add to PATH (add to ~/.bashrc for persistence)
export PATH="$HOME/.cargo/bin:$PATH"
source ~/.bashrc

# Verify installation
uv --version
```

### 3. Install Node.js 20.x

```bash
# Install Node.js 20.x LTS
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt install -y nodejs

# Verify installation
node --version  # Should be v20.x
npm --version   # Should be 10.x
```

### 4. Install Nginx

```bash
# Install nginx
sudo apt install -y nginx

# Start and enable nginx
sudo systemctl start nginx
sudo systemctl enable nginx

# Check status
sudo systemctl status nginx
```

### 5. Install Certbot (Let's Encrypt)

```bash
# Install Certbot for nginx
sudo apt install -y certbot python3-certbot-nginx
```

---

## Phase 4: Deploy Application

### 1. Clone Repository

```bash
# Navigate to home directory
cd ~

# Clone repository (use HTTPS or SSH)
git clone https://github.com/your-username/opengov.git
cd opengov

# Or if using private repo with SSH key:
# git clone git@github.com:your-username/opengov.git
```

### 2. Create Application Directory Structure

```bash
# Create logs directory
mkdir -p ~/opengov/logs

# Create database directory (for SQLite)
mkdir -p ~/opengov/backend/data
```

### 3. Set Up Backend

```bash
cd ~/opengov/backend

# Install dependencies with uv
uv sync

# Create .env file from example
cp .env.example .env

# Edit .env with production values (see Phase 5)
nano .env
```

### 4. Build Frontend

```bash
cd ~/opengov/frontend

# Install dependencies
npm install

# Build for production
npm run build

# Frontend build output is in: frontend/dist/
```

### 5. Set Up Database

```bash
cd ~/opengov/backend

# Run database migrations
uv run alembic upgrade head

# Verify database was created
ls -la data/  # Should see opengov.db
```

---

## Phase 5: Configure Environment Variables

### Backend Environment (.env)

Edit `~/opengov/backend/.env`:

```bash
# Database Configuration
DATABASE_URL=sqlite:///./data/opengov.db

# API Keys
GROK_API_KEY=your-actual-grok-api-key

# External APIs
FEDERAL_REGISTER_API_URL=https://www.federalregister.gov/api/v1
GROK_API_URL=https://api.x.ai/v1

# Google OAuth Configuration
GOOGLE_CLIENT_ID=your-production-google-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-production-google-client-secret
GOOGLE_REDIRECT_URI=https://yourdomain.com/api/auth/google/callback

# Twitter OAuth Configuration
TWITTER_CLIENT_ID=your-production-twitter-client-id
TWITTER_CLIENT_SECRET=your-production-twitter-client-secret
TWITTER_REDIRECT_URI=https://yourdomain.com/api/auth/twitter/callback

# JWT Configuration
# IMPORTANT: Generate a NEW secret key for production!
# Generate with: python3 -c "import secrets; print(secrets.token_urlsafe(32))"
JWT_SECRET_KEY=<generate-new-secret-min-32-chars>
JWT_ALGORITHM=HS256
JWT_ACCESS_TOKEN_EXPIRE_MINUTES=60

# Frontend URL (for OAuth redirects)
FRONTEND_URL=https://yourdomain.com

# Scraper Configuration
SCRAPER_INTERVAL_MINUTES=15
SCRAPER_DAYS_LOOKBACK=1

# CORS Configuration
ALLOWED_ORIGINS=https://yourdomain.com

# API Timeouts (seconds)
FEDERAL_REGISTER_TIMEOUT=30
GROK_TIMEOUT=60

# Request Limits
MAX_REQUEST_SIZE_BYTES=10485760
FEDERAL_REGISTER_PER_PAGE=100
FEDERAL_REGISTER_MAX_PAGES=2

# Environment Settings
DEBUG=False
ENVIRONMENT=production
BEHIND_PROXY=True
USE_MOCK_GROK=False

# Authentication Security
# CRITICAL: Must be True in production with HTTPS
COOKIE_SECURE=True
```

**Important notes:**
- Generate a NEW `JWT_SECRET_KEY` for production
- Update OAuth redirect URIs to use your domain
- Set `COOKIE_SECURE=True` (requires HTTPS)
- Set `BEHIND_PROXY=True` (nginx reverse proxy)
- Update `ALLOWED_ORIGINS` and `FRONTEND_URL` to your domain

### Generate JWT Secret Key

```bash
# Generate secure JWT secret
python3 -c "import secrets; print(secrets.token_urlsafe(32))"

# Copy output and paste into .env as JWT_SECRET_KEY
```

---

## Phase 6: Configure Systemd Service

Create systemd service to run the backend automatically.

### 1. Create Service File

```bash
sudo nano /etc/systemd/system/opengov.service
```

Add the following content:

```ini
[Unit]
Description=OpenGov FastAPI Application
After=network.target

[Service]
Type=simple
User=opengov
Group=opengov
WorkingDirectory=/home/opengov/opengov/backend
Environment="PATH=/home/opengov/.cargo/bin:/home/opengov/.local/bin:/usr/local/bin:/usr/bin:/bin"

# Run with uvicorn via uv
ExecStart=/home/opengov/.cargo/bin/uv run uvicorn app.main:app --host 127.0.0.1 --port 8000 --workers 2

# Restart policy
Restart=always
RestartSec=10

# Resource limits
LimitNOFILE=4096

# Logging
StandardOutput=append:/home/opengov/opengov/logs/opengov.log
StandardError=append:/home/opengov/opengov/logs/opengov-error.log

[Install]
WantedBy=multi-user.target
```

**Notes:**
- Runs as `opengov` user (non-root)
- Binds to `127.0.0.1:8000` (nginx will proxy to this)
- Uses 2 workers (adjust based on CPU cores)
- Automatically restarts on failure
- Logs to `~/opengov/logs/`

### 2. Enable and Start Service

```bash
# Reload systemd to recognize new service
sudo systemctl daemon-reload

# Enable service to start on boot
sudo systemctl enable opengov

# Start service
sudo systemctl start opengov

# Check status
sudo systemctl status opengov

# View logs
sudo journalctl -u opengov -f
```

### 3. Verify Backend is Running

```bash
# Test backend locally
curl http://127.0.0.1:8000/api/feed
```

---

## Phase 7: Configure Nginx

### 1. Create Nginx Configuration

```bash
sudo nano /etc/nginx/sites-available/opengov
```

Add the following configuration:

```nginx
# Rate limiting zone
limit_req_zone $binary_remote_addr zone=api_limit:10m rate=10r/s;

# Upstream backend
upstream opengov_backend {
    server 127.0.0.1:8000 fail_timeout=0;
}

server {
    listen 80;
    server_name yourdomain.com www.yourdomain.com;

    # Redirect HTTP to HTTPS (after SSL is configured)
    # Uncomment after running certbot:
    # return 301 https://$server_name$request_uri;

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    # Serve frontend static files
    location / {
        root /home/opengov/opengov/frontend/dist;
        try_files $uri $uri/ /index.html;

        # Cache static assets
        location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
            expires 1y;
            add_header Cache-Control "public, immutable";
        }
    }

    # Proxy API requests to backend
    location /api/ {
        # Rate limiting
        limit_req zone=api_limit burst=20 nodelay;

        proxy_pass http://opengov_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # Timeouts
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;

        # Disable buffering for SSE/streaming
        proxy_buffering off;
    }

    # Health check endpoint
    location /health {
        proxy_pass http://opengov_backend/api/feed?limit=1;
        access_log off;
    }

    # Deny access to hidden files
    location ~ /\. {
        deny all;
        access_log off;
        log_not_found off;
    }

    # Custom error pages
    error_page 502 503 504 /50x.html;
    location = /50x.html {
        root /usr/share/nginx/html;
    }
}
```

**Important:** Replace `yourdomain.com` with your actual domain.

### 2. Enable Site and Test Configuration

```bash
# Create symbolic link to enable site
sudo ln -s /etc/nginx/sites-available/opengov /etc/nginx/sites-enabled/

# Remove default site (optional)
sudo rm /etc/nginx/sites-enabled/default

# Test nginx configuration
sudo nginx -t

# If test passes, reload nginx
sudo systemctl reload nginx
```

### 3. Verify Nginx is Serving the Site

```bash
# Test from server
curl http://localhost

# Test from your computer
# Open browser: http://yourdomain.com
```

---

## Phase 8: SSL/TLS with Let's Encrypt

### 1. Obtain SSL Certificate

```bash
# Run Certbot to obtain and configure SSL
sudo certbot --nginx -d yourdomain.com -d www.yourdomain.com

# Follow prompts:
# - Enter email address
# - Agree to Terms of Service
# - Choose whether to redirect HTTP to HTTPS (recommended: Yes)
```

Certbot will:
- Obtain SSL certificate from Let's Encrypt
- Automatically modify nginx config to use HTTPS
- Set up HTTP to HTTPS redirect
- Configure SSL security settings

### 2. Verify SSL Certificate

```bash
# Check certificate status
sudo certbot certificates

# Test HTTPS in browser
# Open: https://yourdomain.com
```

### 3. Auto-Renewal Setup

Certbot automatically sets up a systemd timer for renewal:

```bash
# Check renewal timer
sudo systemctl status certbot.timer

# Test renewal (dry run)
sudo certbot renew --dry-run
```

Certificates will auto-renew 30 days before expiration.

### 4. Updated Nginx Config (Post-SSL)

After running Certbot, your nginx config will be updated automatically. Final config should look like:

```nginx
# /etc/nginx/sites-available/opengov (after Certbot)

limit_req_zone $binary_remote_addr zone=api_limit:10m rate=10r/s;

upstream opengov_backend {
    server 127.0.0.1:8000 fail_timeout=0;
}

server {
    listen 80;
    server_name yourdomain.com www.yourdomain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name yourdomain.com www.yourdomain.com;

    # SSL certificates (managed by Certbot)
    ssl_certificate /etc/letsencrypt/live/yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/yourdomain.com/privkey.pem;
    include /etc/letsencrypt/options-ssl-nginx.conf;
    ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem;

    # Security headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    # Serve frontend static files
    location / {
        root /home/opengov/opengov/frontend/dist;
        try_files $uri $uri/ /index.html;

        location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
            expires 1y;
            add_header Cache-Control "public, immutable";
        }
    }

    # Proxy API requests to backend
    location /api/ {
        limit_req zone=api_limit burst=20 nodelay;

        proxy_pass http://opengov_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;

        proxy_buffering off;
    }

    location /health {
        proxy_pass http://opengov_backend/api/feed?limit=1;
        access_log off;
    }

    location ~ /\. {
        deny all;
        access_log off;
        log_not_found off;
    }

    error_page 502 503 504 /50x.html;
    location = /50x.html {
        root /usr/share/nginx/html;
    }
}
```

---

## Phase 9: Update OAuth Settings

After SSL is configured, update OAuth redirect URIs in provider consoles.

### Google OAuth

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Select your project
3. Navigate to "APIs & Services" > "Credentials"
4. Edit your OAuth 2.0 Client ID
5. Add authorized redirect URI:
   - `https://yourdomain.com/api/auth/google/callback`
6. Save changes

### Twitter OAuth

1. Go to [Twitter Developer Portal](https://developer.twitter.com/)
2. Select your app
3. Navigate to "Settings" > "User authentication settings"
4. Update callback URI:
   - `https://yourdomain.com/api/auth/twitter/callback`
5. Save changes

### Update Backend .env

Ensure your `backend/.env` has production URLs:

```bash
GOOGLE_REDIRECT_URI=https://yourdomain.com/api/auth/google/callback
TWITTER_REDIRECT_URI=https://yourdomain.com/api/auth/twitter/callback
FRONTEND_URL=https://yourdomain.com
ALLOWED_ORIGINS=https://yourdomain.com
COOKIE_SECURE=True
```

Restart backend after changes:

```bash
sudo systemctl restart opengov
```

---

## Phase 10: Security Hardening

### 1. Configure Fail2Ban

Protect against brute force attacks:

```bash
# Install fail2ban (already installed in Phase 2)
# Configure SSH jail
sudo nano /etc/fail2ban/jail.local
```

Add:

```ini
[sshd]
enabled = true
port = ssh
logpath = %(sshd_log)s
maxretry = 3
bantime = 3600
findtime = 600
```

Restart fail2ban:

```bash
sudo systemctl restart fail2ban
sudo systemctl status fail2ban
```

### 2. Secure SSH

```bash
# Edit SSH config
sudo nano /etc/ssh/sshd_config
```

Recommended settings:

```
PermitRootLogin no
PasswordAuthentication no
PubkeyAuthentication yes
X11Forwarding no
```

Restart SSH:

```bash
sudo systemctl restart sshd
```

### 3. Set File Permissions

```bash
# Secure .env file
chmod 600 ~/opengov/backend/.env

# Secure database directory
chmod 700 ~/opengov/backend/data
chmod 600 ~/opengov/backend/data/opengov.db

# Secure log directory
chmod 750 ~/opengov/logs
```

### 4. Configure Automatic Security Updates

```bash
# Install unattended-upgrades
sudo apt install -y unattended-upgrades

# Enable automatic security updates
sudo dpkg-reconfigure -plow unattended-upgrades
```

---

## Phase 11: Monitoring and Logging

### 1. View Application Logs

```bash
# Backend logs (systemd)
sudo journalctl -u opengov -f

# Backend logs (file)
tail -f ~/opengov/logs/opengov.log
tail -f ~/opengov/logs/opengov-error.log

# Scraper logs
tail -f ~/opengov/backend/scraper.log

# Nginx access logs
sudo tail -f /var/log/nginx/access.log

# Nginx error logs
sudo tail -f /var/log/nginx/error.log
```

### 2. Monitor System Resources

```bash
# Check disk usage
df -h

# Check memory usage
free -h

# Check running processes
htop  # (install with: sudo apt install htop)

# Check service status
sudo systemctl status opengov
sudo systemctl status nginx
```

### 3. Set Up Log Rotation

Backend logs are rotated automatically by systemd. For custom logs:

```bash
sudo nano /etc/logrotate.d/opengov
```

Add:

```
/home/opengov/opengov/backend/scraper.log {
    daily
    missingok
    rotate 14
    compress
    delaycompress
    notifempty
    create 0640 opengov opengov
}
```

Test:

```bash
sudo logrotate -f /etc/logrotate.d/opengov
```

### 4. DigitalOcean Monitoring

Enable monitoring in DigitalOcean dashboard:
- CPU usage
- Memory usage
- Disk I/O
- Network traffic
- Alerts for high usage

---

## Phase 12: Database Management

### 1. Database Backups

Create backup script:

```bash
nano ~/opengov/scripts/backup-db.sh
```

Add:

```bash
#!/bin/bash

# Database backup script
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/home/opengov/backups"
DB_PATH="/home/opengov/opengov/backend/data/opengov.db"
BACKUP_FILE="$BACKUP_DIR/opengov_db_$TIMESTAMP.db"

# Create backup directory if it doesn't exist
mkdir -p "$BACKUP_DIR"

# Copy database
cp "$DB_PATH" "$BACKUP_FILE"

# Compress backup
gzip "$BACKUP_FILE"

# Keep only last 7 days of backups
find "$BACKUP_DIR" -name "opengov_db_*.db.gz" -mtime +7 -delete

echo "Backup completed: $BACKUP_FILE.gz"
```

Make executable:

```bash
chmod +x ~/opengov/scripts/backup-db.sh
```

### 2. Automated Daily Backups

Add to crontab:

```bash
crontab -e
```

Add line:

```cron
# Daily database backup at 3 AM
0 3 * * * /home/opengov/opengov/scripts/backup-db.sh >> /home/opengov/opengov/logs/backup.log 2>&1
```

### 3. Run Migrations

When deploying updates with database changes:

```bash
cd ~/opengov/backend

# Pull latest code
git pull

# Run migrations
uv run alembic upgrade head

# Restart backend
sudo systemctl restart opengov
```

---

## Phase 13: Deployment Workflow

### 1. Manual Deployment Script

Create deployment script for updates:

```bash
nano ~/opengov/scripts/deploy.sh
```

Add:

```bash
#!/bin/bash
set -e

echo "Starting deployment..."

# Navigate to project directory
cd /home/opengov/opengov

# Pull latest code
echo "Pulling latest code..."
git pull origin main

# Backend updates
echo "Updating backend..."
cd backend
uv sync
uv run alembic upgrade head
cd ..

# Frontend updates
echo "Building frontend..."
cd frontend
npm install
npm run build
cd ..

# Restart backend service
echo "Restarting backend service..."
sudo systemctl restart opengov

# Reload nginx
echo "Reloading nginx..."
sudo systemctl reload nginx

echo "Deployment complete!"
echo "Checking service status..."
sudo systemctl status opengov --no-pager
```

Make executable:

```bash
chmod +x ~/opengov/scripts/deploy.sh
```

### 2. Deploy Updates

```bash
# Run deployment script
~/opengov/scripts/deploy.sh

# Monitor logs during deployment
sudo journalctl -u opengov -f
```

### 3. Rollback Strategy

If deployment fails:

```bash
# Navigate to project
cd ~/opengov

# Revert to previous commit
git log --oneline -n 5  # Find previous commit hash
git checkout <previous-commit-hash>

# Rebuild and restart
cd frontend && npm run build && cd ..
sudo systemctl restart opengov
```

---

## Phase 14: Performance Optimization

### 1. Adjust Worker Count

Edit systemd service to match CPU cores:

```bash
sudo nano /etc/systemd/system/opengov.service
```

Change `--workers` to `(2 * CPU_cores) + 1`:

```ini
# For 1 CPU: --workers 2
# For 2 CPU: --workers 4
ExecStart=/home/opengov/.cargo/bin/uv run uvicorn app.main:app --host 127.0.0.1 --port 8000 --workers 4
```

Reload and restart:

```bash
sudo systemctl daemon-reload
sudo systemctl restart opengov
```

### 2. Nginx Caching (Optional)

Add caching for API responses:

```nginx
# Add to nginx config
proxy_cache_path /var/cache/nginx/opengov levels=1:2 keys_zone=api_cache:10m max_size=100m inactive=60m;

# In location /api/ block:
proxy_cache api_cache;
proxy_cache_valid 200 5m;
proxy_cache_key "$scheme$request_method$host$request_uri";
add_header X-Cache-Status $upstream_cache_status;
```

### 3. Enable Gzip Compression

Edit nginx config:

```bash
sudo nano /etc/nginx/nginx.conf
```

Ensure gzip settings are enabled:

```nginx
gzip on;
gzip_vary on;
gzip_min_length 1024;
gzip_types text/plain text/css text/xml text/javascript application/x-javascript application/xml+rss application/json application/javascript;
```

---

## Troubleshooting

### Backend Service Won't Start

```bash
# Check service status
sudo systemctl status opengov

# View detailed logs
sudo journalctl -u opengov -n 50

# Common issues:
# - Missing .env file
# - Database not migrated
# - Wrong file permissions
# - Python dependencies not installed
```

### Nginx 502 Bad Gateway

```bash
# Check if backend is running
curl http://127.0.0.1:8000/api/feed

# Check nginx error logs
sudo tail -f /var/log/nginx/error.log

# Restart backend
sudo systemctl restart opengov
```

### SSL Certificate Issues

```bash
# Check certificate status
sudo certbot certificates

# Renew certificate manually
sudo certbot renew

# Test nginx config
sudo nginx -t
```

### Database Locked

```bash
# Check if multiple processes are accessing database
ps aux | grep uvicorn

# Stop service and restart
sudo systemctl stop opengov
sudo systemctl start opengov
```

### OAuth Not Working

Check:
- Redirect URIs match exactly in provider console
- `COOKIE_SECURE=True` with HTTPS
- `ALLOWED_ORIGINS` includes your domain
- `FRONTEND_URL` is correct

### High Memory Usage

```bash
# Check memory usage
free -h

# Identify processes using most memory
ps aux --sort=-%mem | head

# Reduce worker count if needed
sudo nano /etc/systemd/system/opengov.service
# Reduce --workers value
sudo systemctl daemon-reload
sudo systemctl restart opengov
```

---

## Maintenance Tasks

### Daily
- Monitor error logs: `sudo journalctl -u opengov --since today`
- Check disk space: `df -h`

### Weekly
- Review access logs for unusual activity
- Check service uptime: `sudo systemctl status opengov`
- Verify backups exist: `ls -lh ~/backups/`

### Monthly
- Review and rotate logs
- Update system packages: `sudo apt update && sudo apt upgrade`
- Test backup restoration
- Review security patches

### Quarterly
- Review and update dependencies
- Security audit (check for vulnerabilities)
- Performance review (optimize if needed)
- Review and update SSL certificates (auto-renewed by Certbot)

---

## Cost Estimate

### DigitalOcean Costs

**Basic Setup:**
- Droplet (2 GB RAM): $12/month
- Backups (+20%): $2.40/month
- Bandwidth: Included (1 TB)
- **Total: ~$15/month**

**Production Setup:**
- Droplet (4 GB RAM): $24/month
- Backups (+20%): $4.80/month
- Bandwidth: Included (2 TB)
- **Total: ~$29/month**

**Additional Costs:**
- Domain name: $10-15/year
- SSL certificate: Free (Let's Encrypt)

---

## Security Checklist

- [x] Firewall configured (ufw)
- [x] Fail2Ban installed and configured
- [x] SSH secured (no root login, key-only auth)
- [x] SSL/TLS enabled (HTTPS)
- [x] Environment variables secured (600 permissions)
- [x] Database file secured (600 permissions)
- [x] Non-root user for app (opengov)
- [x] Security headers configured in nginx
- [x] Rate limiting enabled
- [x] Automatic security updates enabled
- [x] Strong JWT secret key generated
- [x] COOKIE_SECURE=True
- [x] BEHIND_PROXY=True

---

## Post-Deployment Verification

### 1. Functionality Tests

```bash
# Test frontend loads
curl -I https://yourdomain.com
# Should return 200 OK

# Test API
curl https://yourdomain.com/api/feed
# Should return JSON response

# Test OAuth redirects
curl -I https://yourdomain.com/api/auth/google/login
# Should return 302 redirect

# Test health check
curl https://yourdomain.com/health
# Should return 200 OK
```

### 2. Browser Tests

- [ ] Homepage loads correctly
- [ ] Login with email/password works
- [ ] Login with Google OAuth works
- [ ] Login with Twitter OAuth works
- [ ] Feed displays articles
- [ ] Article details page works
- [ ] Responsive design works on mobile
- [ ] HTTPS certificate shows valid (green padlock)

### 3. Service Tests

```bash
# Check all services are running
sudo systemctl status opengov
sudo systemctl status nginx
sudo systemctl status certbot.timer

# Check logs for errors
sudo journalctl -u opengov --since "1 hour ago" | grep -i error
```

---

## Resources

### Documentation
- [DigitalOcean Docs](https://docs.digitalocean.com/)
- [Nginx Docs](https://nginx.org/en/docs/)
- [Let's Encrypt Docs](https://letsencrypt.org/docs/)
- [Systemd Documentation](https://www.freedesktop.org/software/systemd/man/)

### Monitoring Tools
- [DigitalOcean Monitoring](https://www.digitalocean.com/docs/monitoring/)
- [Uptime Robot](https://uptimerobot.com/) - Free uptime monitoring
- [Better Uptime](https://betteruptime.com/) - Status page + monitoring

### Performance Testing
- [GTmetrix](https://gtmetrix.com/) - Page speed analysis
- [WebPageTest](https://www.webpagetest.org/) - Performance testing
- [SSL Labs](https://www.ssllabs.com/ssltest/) - SSL configuration test

---

## Summary

You've successfully deployed OpenGov to DigitalOcean! Your application now has:

✅ **Secure Infrastructure**
- Ubuntu 24.04 LTS server
- Firewall and fail2ban protection
- Non-root user setup

✅ **Application Stack**
- FastAPI backend with systemd service
- React frontend built and served by nginx
- SQLite database with automated backups

✅ **Web Server**
- Nginx reverse proxy
- SSL/TLS with Let's Encrypt
- Rate limiting and security headers

✅ **Production Features**
- OAuth authentication (Google + Twitter)
- Automated scraper with APScheduler
- Logging and monitoring
- Automated backups and updates

✅ **Security**
- HTTPS enforced
- Secure cookies
- Protected credentials
- Automatic security updates

Your site is now live at **https://yourdomain.com** 🚀

For support, check the troubleshooting section or review logs for detailed error messages.
