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

	return nil
}
