# Architecture

## System Architecture

```
                        Internet
                           │
                    ┌──────▼──────┐
                    │   Router    │
                    │ :8081→8081  │
                    │ :1935→1935  │
                    └──────┬──────┘
                           │
              ┌────────────▼────────────┐
              │    Nginx Proxy :8081    │
              │                         │
              │  /api/*    → Backend    │
              │  /hls/*    → MediaMTX   │
              │  /webrtc/* → MediaMTX   │
              │  /*        → Frontend   │
              └────────────┬────────────┘
                           │
         ┌─────────────────┼─────────────────┐
         │                 │                 │
    ┌────▼─────┐    ┌─────▼──────┐   ┌─────▼──────┐
    │ Frontend │    │  Backend   │   │  MediaMTX  │
    │  React   │    │  Go API    │   │   Stream   │
    │  :80     │    │  :3000     │   │  :8888 HLS │
    └──────────┘    └─────┬──────┘   │  :8889 RTC │
                          │          │  :1935 RTMP│
                     ┌────▼─────┐    └────────────┘
                     │  SQLite  │
                     │ Database │
                     └──────────┘
```

## Stream Flow

```
Camera → RTMP (:1935) → MediaMTX → HLS/WebRTC
                                    ↓
Frontend ← Nginx ← HLS (:8888) or WebRTC (:8889)
```

## Component Responsibilities

**Nginx**: Reverse proxy, single entry point (port 8081)

**Frontend**: React SPA, HLS.js + WebRTC player

**Backend**: Go Fiber, REST API, JWT auth, SQLite database

**MediaMTX**: RTMP ingestion, HLS + WebRTC streaming

## Ports

| Service | Internal | External |
|---------|----------|----------|
| Nginx | 80 | 8081 |
| Backend | 3000 | - |
| Frontend | 80 | - |
| MediaMTX HLS | 8888 | - |
| MediaMTX WebRTC | 8889 | - |
| MediaMTX RTMP | 1935 | 1935 |
