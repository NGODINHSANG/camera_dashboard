# Database Schema

**Engine**: SQLite
**ORM**: GORM

## ER Diagram

```
┌──────────────┐         ┌──────────────┐         ┌──────────────┐
│    Users     │         │   Projects   │         │   Cameras    │
├──────────────┤         ├──────────────┤         ├──────────────┤
│ id           │◄───────►│ id           │◄───────►│ id           │
│ email        │  1:N    │ user_id (FK) │  1:N    │ project_id   │
│ password     │         │ name         │         │ name         │
│ name         │         │ created_at   │         │ location     │
│ role         │         │ updated_at   │         │ stream_url   │
│ created_at   │         └──────────────┘         │ is_recording │
│ updated_at   │                                  │ aspect_ratio │
└──────────────┘                                  │ created_at   │
                                                  │ updated_at   │
                                                  └───────┬──────┘
                                                          │
                                                          │ 1:N
                                                  ┌───────▼──────┐
                                                  │  Recordings  │
                                                  ├──────────────┤
                                                  │ id           │
                                                  │ camera_id    │
                                                  │ file_path    │
                                                  │ file_size    │
                                                  │ start_time   │
                                                  │ end_time     │
                                                  │ created_at   │
                                                  └──────────────┘
```

## Tables

### users
```
id              INTEGER PRIMARY KEY
email           VARCHAR(255) UNIQUE NOT NULL
password        VARCHAR(255) NOT NULL (bcrypt hashed)
name            VARCHAR(100) NOT NULL
role            VARCHAR(20) DEFAULT 'user' (admin/user)
created_at      DATETIME
updated_at      DATETIME
```

### projects
```
id              INTEGER PRIMARY KEY
user_id         INTEGER NOT NULL (FK → users.id)
name            VARCHAR(255) NOT NULL
description     TEXT
created_at      DATETIME
updated_at      DATETIME
```

### cameras
```
id              INTEGER PRIMARY KEY
project_id      INTEGER NOT NULL (FK → projects.id)
name            VARCHAR(255) NOT NULL
location        VARCHAR(255)
stream_url      VARCHAR(500) NOT NULL
is_recording    BOOLEAN DEFAULT FALSE
aspect_ratio    VARCHAR(20) DEFAULT 'auto' (16:9, 4:3, 1:1, auto)
created_at      DATETIME
updated_at      DATETIME
```

### recordings
```
id              INTEGER PRIMARY KEY
camera_id       INTEGER NOT NULL (FK → cameras.id)
file_path       VARCHAR(500) NOT NULL
file_size       BIGINT DEFAULT 0
start_time      DATETIME NOT NULL
end_time        DATETIME (NULL if recording)
duration        INTEGER (seconds)
created_at      DATETIME
```

## Relationships

- **User → Projects**: 1:N (Một user có nhiều projects)
- **Project → Cameras**: 1:N (Một project có nhiều cameras)
- **Camera → Recordings**: 1:N (Một camera có nhiều recordings)

## Indexes

```sql
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_projects_user_id ON projects(user_id);
CREATE INDEX idx_cameras_project_id ON cameras(project_id);
CREATE INDEX idx_recordings_camera_id ON recordings(camera_id);
```
