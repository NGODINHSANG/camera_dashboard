package repository

import (
	"camera-dashboard-backend/internal/models"
	"time"

	"github.com/jmoiron/sqlx"
)

type ChunkRepository struct {
	db *sqlx.DB
}

func NewChunkRepository(db *sqlx.DB) *ChunkRepository {
	return &ChunkRepository{db: db}
}

// Create creates a new chunk upload session
func (r *ChunkRepository) Create(chunk *models.ChunkUpload) error {
	query := `
		INSERT INTO chunk_uploads (id, camera_id, project_id, original_filename, total_size, total_chunks, uploaded_chunks, chunk_size, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	now := time.Now()
	chunk.CreatedAt = now
	chunk.UpdatedAt = now

	_, err := r.db.Exec(query,
		chunk.ID,
		chunk.CameraID,
		chunk.ProjectID,
		chunk.OriginalFilename,
		chunk.TotalSize,
		chunk.TotalChunks,
		chunk.UploadedChunks,
		chunk.ChunkSize,
		chunk.Status,
		chunk.CreatedAt,
		chunk.UpdatedAt,
	)
	return err
}

// GetByID gets a chunk upload session by ID
func (r *ChunkRepository) GetByID(id string) (*models.ChunkUpload, error) {
	var chunk models.ChunkUpload
	err := r.db.Get(&chunk, "SELECT * FROM chunk_uploads WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &chunk, nil
}

// UpdateStatus updates the status of a chunk upload
func (r *ChunkRepository) UpdateStatus(id string, status string) error {
	query := `UPDATE chunk_uploads SET status = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, status, time.Now(), id)
	return err
}

// UpdateError updates the error message and sets status to failed
func (r *ChunkRepository) UpdateError(id string, errMsg string) error {
	query := `UPDATE chunk_uploads SET status = ?, error = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, models.ChunkStatusFailed, errMsg, time.Now(), id)
	return err
}

// Delete deletes a chunk upload session
func (r *ChunkRepository) Delete(id string) error {
	_, err := r.db.Exec("DELETE FROM chunk_uploads WHERE id = ?", id)
	return err
}

// CleanupOld removes old chunk uploads
func (r *ChunkRepository) CleanupOld(olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)
	_, err := r.db.Exec("DELETE FROM chunk_uploads WHERE created_at < ?", cutoff)
	return err
}
