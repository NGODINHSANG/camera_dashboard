package database

import (
	"github.com/jmoiron/sqlx"
)

func RunMigrations(db *sqlx.DB) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			name TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'user' CHECK(role IN ('admin', 'user')),
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS projects (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS cameras (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			project_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			location TEXT DEFAULT '',
			stream_url TEXT NOT NULL,
			is_recording BOOLEAN DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_projects_user_id ON projects(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_cameras_project_id ON cameras(project_id)`,
		`CREATE TABLE IF NOT EXISTS recordings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			camera_id INTEGER NOT NULL,
			project_id INTEGER NOT NULL,
			filename TEXT NOT NULL,
			file_path TEXT NOT NULL,
			output_dir TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'recording' CHECK(status IN ('recording', 'completed', 'failed')),
			file_size INTEGER DEFAULT 0,
			started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			stopped_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (camera_id) REFERENCES cameras(id) ON DELETE CASCADE,
			FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_recordings_camera_id ON recordings(camera_id)`,
		`CREATE INDEX IF NOT EXISTS idx_recordings_project_id ON recordings(project_id)`,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return err
		}
	}

	// Add auto_record column to cameras (SQLite ALTER TABLE)
	// This will fail silently if column already exists
	db.Exec(`ALTER TABLE cameras ADD COLUMN auto_record BOOLEAN DEFAULT 0`)

	// Add source column to recordings (stream or upload)
	db.Exec(`ALTER TABLE recordings ADD COLUMN source TEXT DEFAULT 'stream'`)

	// Create upload_jobs table for tracking upload progress
	db.Exec(`CREATE TABLE IF NOT EXISTS upload_jobs (
		id TEXT PRIMARY KEY,
		camera_id INTEGER NOT NULL,
		project_id INTEGER NOT NULL,
		filename TEXT NOT NULL,
		original_filename TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'uploading' CHECK(status IN ('uploading', 'processing', 'completed', 'failed')),
		progress INTEGER DEFAULT 0,
		file_size INTEGER DEFAULT 0,
		error TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (camera_id) REFERENCES cameras(id) ON DELETE CASCADE,
		FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
	)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_upload_jobs_camera_id ON upload_jobs(camera_id)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_upload_jobs_status ON upload_jobs(status)`)

	// Create project_permissions table for per-project admin roles
	db.Exec(`CREATE TABLE IF NOT EXISTS project_permissions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		project_id INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(user_id, project_id),
		FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE
	)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_project_permissions_user_id ON project_permissions(user_id)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_project_permissions_project_id ON project_permissions(project_id)`)

	// Create chunk_uploads table for chunked upload sessions
	db.Exec(`CREATE TABLE IF NOT EXISTS chunk_uploads (
		id TEXT PRIMARY KEY,
		camera_id INTEGER NOT NULL,
		project_id INTEGER NOT NULL,
		original_filename TEXT NOT NULL,
		total_size INTEGER NOT NULL,
		total_chunks INTEGER NOT NULL,
		uploaded_chunks INTEGER DEFAULT 0,
		chunk_size INTEGER NOT NULL,
		status TEXT NOT NULL DEFAULT 'uploading' CHECK(status IN ('uploading', 'merging', 'processing', 'completed', 'failed')),
		error TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (camera_id) REFERENCES cameras(id) ON DELETE CASCADE,
		FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
	)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_chunk_uploads_camera_id ON chunk_uploads(camera_id)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_chunk_uploads_status ON chunk_uploads(status)`)

	// ── Chat tables ──────────────────────────────────────────────────────────

	// conversations: DM hoặc Group
	db.Exec(`CREATE TABLE IF NOT EXISTS conversations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		type TEXT NOT NULL DEFAULT 'dm' CHECK(type IN ('dm','group')),
		name TEXT,
		created_by INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(created_by) REFERENCES users(id) ON DELETE CASCADE
	)`)

	// conversation_members: ai tham gia cuộc trò chuyện nào
	db.Exec(`CREATE TABLE IF NOT EXISTS conversation_members (
		conversation_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		joined_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_read_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY(conversation_id, user_id),
		FOREIGN KEY(conversation_id) REFERENCES conversations(id) ON DELETE CASCADE,
		FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
	)`)

	// messages: tin nhắn
	db.Exec(`CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		conversation_id INTEGER NOT NULL,
		sender_id INTEGER NOT NULL,
		content TEXT,
		message_type TEXT NOT NULL DEFAULT 'text' CHECK(message_type IN ('text','file','image')),
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(conversation_id) REFERENCES conversations(id) ON DELETE CASCADE,
		FOREIGN KEY(sender_id) REFERENCES users(id) ON DELETE CASCADE
	)`)

	// message_attachments: file/ảnh đính kèm
	db.Exec(`CREATE TABLE IF NOT EXISTS message_attachments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		message_id INTEGER NOT NULL,
		filename TEXT NOT NULL,
		original_name TEXT NOT NULL,
		file_size INTEGER DEFAULT 0,
		mime_type TEXT,
		url TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(message_id) REFERENCES messages(id) ON DELETE CASCADE
	)`)

	db.Exec(`CREATE INDEX IF NOT EXISTS idx_conversation_members_user ON conversation_members(user_id)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_messages_conversation ON messages(conversation_id)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_messages_sender ON messages(sender_id)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_conversations_updated ON conversations(updated_at DESC)`)

	return nil
}
