package database

import (
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func Connect(dbPath string) (*sqlx.DB, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	db, err := sqlx.Connect("sqlite", dbPath+"?_foreign_keys=on")
	if err != nil {
		return nil, err
	}

	// Run migrations
	if err := RunMigrations(db); err != nil {
		return nil, err
	}

	return db, nil
}
