# Camera Dashboard

Hệ thống giám sát camera đa dự án với streaming HLS và WebRTC.

---

## Tính năng

**Streaming đa giao thức**
- HLS (HTTP Live Streaming) - độ trễ thấp
- WebRTC (WHEP) - độ trễ cực thấp <500ms
- RTMP Ingestion từ camera

**Quản lý đa dự án**
- Nhiều dự án độc lập
- Mỗi dự án có danh sách camera riêng
- Phân quyền Admin/User

**Giao diện linh hoạt**
- Grid layout: 2x2, 3x3, 4x4, 5x5, 6x6
- Fullscreen mode cho từng camera
- Tùy chỉnh aspect ratio (16:9, 4:3, 1:1, auto)

**Ghi hình**
- Ghi trực tiếp từ stream
- Quản lý recordings

---

## Kiến trúc

```
Internet → Router (:8081, :1935)
    ↓
Nginx Reverse Proxy (:8081)
    ├─→ Frontend (React)
    ├─→ Backend (Go API)
    └─→ MediaMTX (Streaming)
```

---

## Tech Stack

**Frontend**: React 18 + Vite + HLS.js + WebRTC

**Backend**: Go + Fiber + SQLite + GORM + JWT

**Streaming**: MediaMTX (RTMP/HLS/WebRTC) + FFmpeg

**Infra**: Docker + Docker Compose + Nginx

---

## Quick Start

### Development

```bash
# Frontend
npm install
npm run dev

# Backend
cd backend
go run cmd/api/main.go
```

### Production

```bash
# Build images
build-images.bat

# Deploy
docker-compose -f docker-compose.prod.yml up -d
```

---

## Stream URLs

**HLS**: `/hls/{stream_name}/index.m3u8`

**WebRTC**: `/webrtc/{stream_name}/whep`

**RTMP Push**: `rtmp://host:1935/{stream_name}`

---

## Default Login

```
Email: admin@example.com
Password: admin123
```

---

## Cấu trúc Project

```
Dashboard/
├── src/                # Frontend React
├── backend/            # Backend Go
├── nginx/              # Nginx config
├── docs/               # Documentation
├── mediamtx.yml        # MediaMTX config
└── docker-compose.prod.yml
```

---

## Tài liệu

**API Documentation** - [API.md](./API.md) - REST API endpoints

**Architecture** - [ARCHITECTURE.md](./ARCHITECTURE.md) - Sơ đồ kiến trúc

**Database** - [DATABASE.md](./DATABASE.md) - Schema database

**Deployment** - [DEPLOYMENT.md](./DEPLOYMENT.md) - Hướng dẫn triển khai

**Technology** - [TECHNOLOGY.md](./TECHNOLOGY.md) - Stack công nghệ

---

## Ports

| Service | Port | Description |
|---------|------|-------------|
| Nginx | 8081 | Web access |
| MediaMTX RTMP | 1935 | Stream ingestion |
| Backend | 3000 | API (internal) |
| MediaMTX HLS | 8888 | HLS (internal) |
| MediaMTX WebRTC | 8889 | WebRTC (internal) |

---

## License

Private - Camera Monitoring System
