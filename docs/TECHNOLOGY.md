# Technology Stack

## Frontend

**Framework**: React 18 + Vite 5

**Video Playback**:
- HLS.js - HTTP Live Streaming
- WebRTC (Native) - Ultra-low latency

**Routing**: React Router v6

**State**: React Context API

**Styling**: Plain CSS với CSS Variables

---

## Backend

**Language**: Go 1.22+

**Framework**: Fiber v2 (Express-like API)

**Database**: SQLite + GORM ORM

**Authentication**: JWT (golang-jwt/jwt) + Bcrypt

**Recording**: FFmpeg (via os/exec)

---

## Streaming

**MediaMTX** (bluenviron/mediamtx)
- RTMP ingestion (port 1935)
- HLS output (port 8888)
- WebRTC/WHEP output (port 8889)

**FFmpeg**
- Recording streams
- Transcoding

---

## Infrastructure

**Docker** - Containerization

**Docker Compose** - Service orchestration

**Nginx** - Reverse proxy (port 8081)
- Routes: `/api/*`, `/hls/*`, `/webrtc/*`, `/*`

---

## Development Tools

**Frontend**:
- Node.js 20+
- npm

**Backend**:
- Go toolchain
- Air (live reload)

**Database**:
- SQLite CLI
- DB Browser for SQLite
