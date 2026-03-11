# Deployment Guide

## Requirements

- **Server**: Ubuntu 20.04+ / Windows Server
- **CPU**: 4 cores
- **RAM**: 16GB
- **Storage**: 256GB SSD
- **Docker**: 24.0+
- **Docker Compose**: 2.20+

---

## Deployment Steps

### 1. Build Images (Development Machine)

Chạy script build:
```bash
# Windows
build-images.bat

# Linux/Mac
./build-images.sh
```

Output:
- `deploy/frontend.tar`
- `deploy/backend.tar`

---

### 2. Transfer to Server

```bash
# Đóng gói toàn bộ project
zip -r dashboard.zip . -x "node_modules/*" -x ".git/*"

# Copy lên server
scp dashboard.zip user@server:/home/user/
```

---

### 3. Setup on Server

```bash
# Unzip
cd /home/user
unzip dashboard.zip -d camera-dashboard
cd camera-dashboard

# Load Docker images
docker load -i deploy/frontend.tar
docker load -i deploy/backend.tar

# Verify images loaded
docker images | grep camera-dashboard
```

---

### 4. Configure Environment

Tạo file `.env`:
```bash
JWT_SECRET=your-super-secret-key-change-this-minimum-32-chars
JWT_EXPIRY=24h
```

---

### 5. Start Services

```bash
# Start all services
docker-compose -f docker-compose.prod.yml up -d

# Check status
docker-compose -f docker-compose.prod.yml ps

# View logs
docker-compose -f docker-compose.prod.yml logs -f
```

---

### 6. Router Configuration

Cấu hình port forwarding trên router:
```
External Port 8081 → Server IP:8081 (HTTP)
External Port 1935 → Server IP:1935 (RTMP)
```

---

### 7. Firewall

```bash
sudo ufw allow 8081/tcp
sudo ufw allow 1935/tcp
sudo ufw enable
```

---

### 8. Verify

```bash
# Check API health
curl http://localhost:8081/api/health

# Access dashboard
http://server-ip:8081

# Default admin login:
Email: admin@example.com
Password: admin123
```

---

## Update

```bash
# Build new images
build-images.bat

# Transfer to server
scp deploy/*.tar user@server:/tmp/

# On server:
cd camera-dashboard
docker-compose -f docker-compose.prod.yml down
docker load -i /tmp/frontend.tar
docker load -i /tmp/backend.tar
docker-compose -f docker-compose.prod.yml up -d
```

---

## Backup

```bash
# Backup database
docker cp camera-backend:/app/data/camera_dashboard.db \
  ./backups/db_$(date +%Y%m%d).db

# Backup recordings
tar -czf backups/recordings_$(date +%Y%m%d).tar.gz recordings/
```

---

## Rollback

```bash
docker-compose -f docker-compose.prod.yml down
docker load -i backups/frontend_v1.0.tar
docker load -i backups/backend_v1.0.tar
cp backups/db_20240101.db backend/data/camera_dashboard.db
docker-compose -f docker-compose.prod.yml up -d
```
