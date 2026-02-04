# API Documentation

**Base URL**: `/api`

---

## Authentication

### POST `/auth/register`
```json
Request:
{
  "email": "user@example.com",
  "password": "password123",
  "name": "John Doe"
}

Response:
{
  "user": { "id": 1, "email": "...", "name": "...", "role": "user" },
  "token": "eyJhbGciOiJIUzI1NiIs..."
}
```

### POST `/auth/login`
```json
Request:
{
  "email": "user@example.com",
  "password": "password123"
}

Response:
{
  "user": { "id": 1, "email": "...", "name": "...", "role": "admin" },
  "token": "eyJhbGciOiJIUzI1NiIs..."
}
```

### GET `/auth/me`
```
Headers: Authorization: Bearer <token>

Response:
{
  "id": 1,
  "email": "user@example.com",
  "name": "John Doe",
  "role": "admin"
}
```

---

## Projects

### GET `/projects`
```
Headers: Authorization: Bearer <token>

Response:
[
  {
    "id": 1,
    "name": "Project Name",
    "cameras": [...]
  }
]
```

### POST `/projects`
```json
Request:
{
  "name": "New Project"
}

Response:
{
  "id": 2,
  "name": "New Project",
  "cameras": []
}
```

### PUT `/projects/:id`
```json
Request:
{
  "name": "Updated Name"
}

Response:
{
  "id": 1,
  "name": "Updated Name"
}
```

### DELETE `/projects/:id`
```json
Response:
{
  "message": "Project deleted successfully"
}
```

---

## Cameras

### GET `/projects/:projectId/cameras`
```
Response:
[
  {
    "id": 1,
    "name": "Camera 1",
    "location": "Location A",
    "streamUrl": "/hls/drone_stream1/index.m3u8",
    "isRecording": false,
    "aspectRatio": "16:9"
  }
]
```

### POST `/projects/:projectId/cameras`
```json
Request:
{
  "name": "Camera 5",
  "location": "Location B",
  "streamUrl": "/webrtc/drone_stream5/whep",
  "aspectRatio": "16:9"
}

Response:
{
  "id": 5,
  "name": "Camera 5",
  ...
}
```

### PUT `/projects/:projectId/cameras/:cameraId`
```json
Request:
{
  "name": "Updated Name",
  "location": "New Location",
  "streamUrl": "/hls/new_stream/index.m3u8"
}

Response:
{
  "id": 1,
  "name": "Updated Name",
  ...
}
```

### DELETE `/projects/:projectId/cameras/:cameraId`
```json
Response:
{
  "message": "Camera deleted successfully"
}
```

---

## Recordings

### POST `/recordings/start`
```json
Request:
{
  "cameraId": 1,
  "outputDir": "/recordings/camera1"
}

Response:
{
  "id": 1,
  "cameraId": 1,
  "filePath": "/recordings/camera1/20240101_120000.mp4",
  "startTime": "2024-01-01T12:00:00Z"
}
```

### POST `/recordings/stop/:id`
```json
Response:
{
  "id": 1,
  "filePath": "/recordings/camera1/20240101_120000.mp4",
  "fileSize": 1024000,
  "duration": 600
}
```

### GET `/recordings/active`
```json
Response:
[
  {
    "id": 1,
    "cameraId": 1,
    "filePath": "/recordings/camera1/20240101_120000.mp4",
    "startTime": "2024-01-01T12:00:00Z"
  }
]
```

---

## Admin

### GET `/admin/users` (Admin only)
```json
Response:
[
  {
    "id": 1,
    "email": "admin@example.com",
    "name": "Admin User",
    "role": "admin"
  }
]
```

### PUT `/admin/users/:id/role` (Admin only)
```json
Request:
{
  "role": "admin"
}

Response:
{
  "id": 2,
  "email": "user@example.com",
  "role": "admin"
}
```

### DELETE `/admin/users/:id` (Admin only)
```json
Response:
{
  "message": "User deleted successfully"
}
```
