# TÀI LIỆU GIỚI THIỆU HỆ THỐNG

**Hệ thống giám sát camera đa dự án**

---

## TỔNG QUAN

Hệ thống giám sát camera trực tuyến với streaming HLS và WebRTC, hỗ trợ quản lý đa dự án và phân quyền Admin/User.

**Tính năng:**
- Quản lý nhiều dự án độc lập, mỗi dự án có danh sách camera riêng
- Streaming HLS (độ trễ < 3s) và WebRTC (độ trễ < 500ms)
- Giao diện grid linh hoạt: 2x2, 3x3, 4x4, 5x5, 6x6
- Fullscreen mode và ghi hình
- Authentication JWT với phân quyền Admin/User

---

## KIẾN TRÚC

```
Internet → Router (:8081, :1935)
    ↓
Nginx Reverse Proxy (:8081)
    ├─→ Frontend (React)
    ├─→ Backend (Go API)
    └─→ MediaMTX (Streaming)
```

**Tech Stack:**
- Frontend: React 18 + Vite + HLS.js + WebRTC
- Backend: Go + Fiber + SQLite + GORM + JWT
- Streaming: MediaMTX (RTMP/HLS/WebRTC) + FFmpeg
- Infrastructure: Docker + Docker Compose + Nginx

---

## CƠ SỞ DỮ LIỆU

```
Users → Projects → Cameras → Recordings
```

**4 bảng chính:**
- users: id, email, password_hash, name, role
- projects: id, user_id, name, description
- cameras: id, project_id, name, location, stream_url, is_recording
- recordings: id, camera_id, file_path, file_size, start_time, end_time

---

## API ENDPOINTS

**Authentication:**
```
POST /api/auth/register    - Đăng ký
POST /api/auth/login       - Đăng nhập
GET  /api/auth/me          - Thông tin user
```

**Projects:**
```
GET    /api/projects       - Danh sách projects
POST   /api/projects       - Tạo project
PUT    /api/projects/:id   - Cập nhật
DELETE /api/projects/:id   - Xóa
```

**Cameras:**
```
POST   /api/projects/:id/cameras           - Thêm camera
PUT    /api/projects/:id/cameras/:cameraId - Cập nhật camera
DELETE /api/projects/:id/cameras/:cameraId - Xóa camera
```

**Admin:**
```
GET    /api/admin/users          - Danh sách users
DELETE /api/admin/users/:id      - Xóa user
PUT    /api/admin/users/:id/role - Đổi role
```

---

## TRIỂN KHAI

**Yêu cầu:**
- CPU: 4 cores, RAM: 8 GB, Storage: 100 GB SSD
- Docker 24.0+, Docker Compose 2.20+

**Các bước:**

1. Build images:
```bash
build-images.bat
```

2. Chuyển lên server:
```bash
zip -r dashboard.zip . -x "node_modules/*" -x ".git/*"
scp dashboard.zip user@server:/home/user/
```

3. Setup trên server:
```bash
unzip dashboard.zip -d camera-dashboard
cd camera-dashboard
docker load -i deploy/frontend.tar
docker load -i deploy/backend.tar
```

4. Tạo file `.env`:
```
JWT_SECRET=your-super-secret-key-minimum-32-chars
JWT_EXPIRY=24h
```

5. Khởi động:
```bash
docker-compose -f docker-compose.prod.yml up -d
```

6. Cấu hình Router (Port Forwarding):
```
External 8081 → Server:8081
External 1935 → Server:1935
```

7. Firewall:
```bash
sudo ufw allow 8081/tcp
sudo ufw allow 1935/tcp
```

8. Truy cập: `http://server-ip:8081`

**Login mặc định:**
- Email: admin@example.com
- Password: admin123

---

## PORTS

| Service | Port | Mô tả |
|---------|------|-------|
| Nginx | 8081 | Web access |
| MediaMTX RTMP | 1935 | Stream ingestion |
| Backend | 3000 | API (internal) |
| MediaMTX HLS | 8888 | HLS (internal) |
| MediaMTX WebRTC | 8889 | WebRTC (internal) |

---

## STREAM URLs

**RTMP Push:**
```
rtmp://server-ip:1935/stream_name
```

**HLS Play:**
```
http://server-ip:8081/hls/stream_name/index.m3u8
```

**WebRTC Play:**
```
http://server-ip:8081/webrtc/stream_name/whep
```

---

## BACKUP

```bash
# Database
docker cp camera-backend:/app/data/camera_dashboard.db \
  ./backups/db_$(date +%Y%m%d).db

# Recordings
tar -czf backups/recordings_$(date +%Y%m%d).tar.gz recordings/
```

---

**Phiên bản:** 1.0 | **Cập nhật:** 04/02/2026
