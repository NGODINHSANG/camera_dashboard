# Single Port Deployment Guide

## 🎯 Architecture

```
Internet (Port 54543)
         ↓
    Nginx Proxy (8081) - ONLY exposed port
         ├─ /           → Frontend (React)
         ├─ /api        → Backend (Go)
         ├─ /hls        → MediaMTX HLS streams
         └─ /webrtc     → MediaMTX WebRTC

    All services run internally (no external ports)
```

## 📦 Services

| Service | Internal Port | External Port | Purpose |
|---------|---------------|---------------|---------|
| **Nginx Proxy** | 80 | **8081** (exposed) | Reverse proxy cho tất cả |
| Frontend | 80 | - (internal) | React UI |
| Backend | 8080 | - (internal) | Go API |
| MediaMTX | 8888, 8889 | - (internal) | HLS + WebRTC streams |

## 🚀 Deployment

### 1. Config Router

**Port forwarding:**
```
Port Forwarding:
54543 → 192.168.1.200:8081 (TCP) - Web access
54541 → 192.168.1.200:1935 (TCP) - RTMP stream ingestion
```

### 2. Deploy

```bash
# Edit .env.production
SERVER_IP=YOUR_PUBLIC_IP

# Deploy
./deploy.sh
```

### 3. Access

**Từ Internet:**
```
http://YOUR_PUBLIC_IP:54543
```

**All routes work qua 1 port:**
- Frontend: `http://YOUR_IP:54543/`
- API: `http://YOUR_IP:54543/api/health`
- HLS: `http://YOUR_IP:54543/hls/camera1/index.m3u8`
- WebRTC: `http://YOUR_IP:54543/webrtc/camera1`

## 📹 Camera Setup

### Option 1: RTMP Push (Recommended)

Camera push stream qua RTMP (port 54541):

```bash
# From camera or FFmpeg
ffmpeg -i rtsp://camera-ip/stream \
       -c copy \
       -f flv rtmp://YOUR_PUBLIC_IP:54541/camera_1
```

### Option 2: RTSP Push (Internal)

Nếu camera cùng mạng với server:

```bash
ffmpeg -i rtsp://camera-ip/stream \
       -c copy \
       -f rtsp rtsp://SERVER_LOCAL_IP:8554/camera_1
```

MediaMTX tự động tạo:
- HLS: `/hls/camera_1/index.m3u8`
- WebRTC: `/webrtc/camera_1`

### Option 3: Static Pull

Edit `mediamtx.yml`:

```yaml
paths:
  camera_1:
    source: rtsp://192.168.1.100:554/stream1
    sourceOnDemand: yes
```

MediaMTX tự động pull khi có viewer.

## 🔧 Config Cameras trong Dashboard

Update camera stream URLs trong database:

```sql
-- HLS
UPDATE cameras SET stream_url = '/hls/camera_1/index.m3u8' WHERE id = 1;

-- WebRTC (faster, less bandwidth)
UPDATE cameras SET stream_url = '/webrtc/camera_1/whep' WHERE id = 1;
```

## 📊 Bandwidth Comparison

### HLS (trước):
```
100 cameras × 2Mbps × 10 users = 2Gbps
```

### WebRTC với MediaMTX (sau):
```
100 cameras × 2Mbps = 200Mbps
(MediaMTX forward, không duplicate)
Giảm 90% băng thông!
```

## 🐛 Troubleshooting

### Cannot access from outside

```bash
# Check nginx proxy running
docker ps | grep nginx-proxy

# Check port mapping
netstat -tulpn | grep 8081

# Test from internet (4G)
curl http://YOUR_PUBLIC_IP:54543/health
```

### Stream not working

```bash
# Check MediaMTX logs
docker logs camera-mediamtx -f

# List active paths
curl http://localhost:9997/v3/paths/list

# Test stream
ffplay http://YOUR_IP:54543/hls/camera_1/index.m3u8
```

### High bandwidth still

- Verify frontend using WebRTC (not HLS)
- Check MediaMTX forwarding (not duplicating)
- Monitor: `docker stats`

## 🔐 Security

### Add SSL (Optional)

```nginx
server {
    listen 443 ssl;
    ssl_certificate /etc/nginx/ssl/cert.pem;
    ssl_certificate_key /etc/nginx/ssl/key.pem;
    # ... rest of config
}
```

Then forward port 443 instead of 54543.

### Basic Auth (Optional)

```nginx
location / {
    auth_basic "Camera Dashboard";
    auth_basic_user_file /etc/nginx/.htpasswd;
    proxy_pass http://frontend/;
}
```

## 📝 Summary

✅ **Chỉ 1 port** (54543) exposed ra ngoài
✅ **Tất cả services** internal communication
✅ **Giảm 90% băng thông** với WebRTC
✅ **Scale** được 100+ cameras
✅ **Deploy đơn giản**: `./deploy.sh`

---

**Next steps:**
1. Forward port 54543 → 8081 trên router
2. Run `./deploy.sh`
3. Access http://YOUR_IP:54543
4. Done! 🎉
