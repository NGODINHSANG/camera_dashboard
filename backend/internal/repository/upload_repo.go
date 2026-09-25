package repository

import (
	"camera-dashboard-backend/internal/models"
	"time"

	"github.com/jmoiron/sqlx"
)

type UploadRepository struct {
	db *sqlx.DB
}

func NewUploadRepository(db *sqlx.DB) *UploadRepository {
	return &UploadRepository{db: db}
}

// Create creates a new upload job
func (r *UploadRepository) Create(job *models.UploadJob) error {
	query := `
		INSERT INTO upload_jobs (id, camera_id, project_id, filename, original_filename, status, progress, file_size, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	now := time.Now()
	job.CreatedAt = now
	job.UpdatedAt = now

	_, err := r.db.Exec(query,
		job.ID,
		job.CameraID,
		job.ProjectID,
		job.Filename,
		job.OriginalFilename,
		job.Status,
		job.Progress,
		job.FileSize,
		job.CreatedAt,
		job.UpdatedAt,
	)
	return err
}

// GetByID gets an upload job by ID
func (r *UploadRepository) GetByID(id string) (*models.UploadJob, error) {
	var job models.UploadJob
	err := r.db.Get(&job, "SELECT * FROM upload_jobs WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &job, nil
}

// UpdateStatus updates the status and progress of an upload job
func (r *UploadRepository) UpdateStatus(id string, status string, progress int) error {
	query := `UPDATE upload_jobs SET status = ?, progress = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, status, progress, time.Now(), id)
	return err
}

// UpdateError updates the error message and sets status to failed
func (r *UploadRepository) UpdateError(id string, errMsg string) error {
	query := `UPDATE upload_jobs SET status = ?, error = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, models.UploadStatusFailed, errMsg, time.Now(), id)
	return err
}

// GetByCameraID gets all upload jobs for a camera
func (r *UploadRepository) GetByCameraID(cameraID int64) ([]models.UploadJob, error) {
	var jobs []models.UploadJob
	err := r.db.Select(&jobs, "SELECT * FROM upload_jobs WHERE camera_id = ? ORDER BY created_at DESC", cameraID)
	return jobs, err
}

// GetPending gets all pending (uploading or processing) jobs
func (r *UploadRepository) GetPending() ([]models.UploadJob, error) {
	var jobs []models.UploadJob
	err := r.db.Select(&jobs, "SELECT * FROM upload_jobs WHERE status IN (?, ?) ORDER BY created_at ASC",
		models.UploadStatusUploading, models.UploadStatusProcessing)
	return jobs, err
}

// Delete deletes an upload job
func (r *UploadRepository) Delete(id string) error {
	_, err := r.db.Exec("DELETE FROM upload_jobs WHERE id = ?", id)
	return err
}

// CleanupOld removes completed/failed jobs older than given duration
func (r *UploadRepository) CleanupOld(olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)
	_, err := r.db.Exec("DELETE FROM upload_jobs WHERE status IN (?, ?) AND created_at < ?",
		models.UploadStatusCompleted, models.UploadStatusFailed, cutoff)
	return err
}
