package repository

import (
	"camera-dashboard-backend/internal/models"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
)

var ErrCameraNotFound = errors.New("camera not found")

type CameraRepository struct {
	db *sqlx.DB
}

func NewCameraRepository(db *sqlx.DB) *CameraRepository {
	return &CameraRepository{db: db}
}

func (r *CameraRepository) Create(camera *models.Camera) error {
	query := `INSERT INTO cameras (project_id, name, location, stream_url, is_recording, created_at, updated_at)
			  VALUES (?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	result, err := r.db.Exec(query, camera.ProjectID, camera.Name, camera.Location, camera.StreamURL, camera.IsRecording, now, now)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	camera.ID = id
	camera.CreatedAt = now
	camera.UpdatedAt = now
	return nil
}

func (r *CameraRepository) GetByID(id int64) (*models.Camera, error) {
	var camera models.Camera
	query := `SELECT * FROM cameras WHERE id = ?`

	err := r.db.Get(&camera, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrCameraNotFound
		}
		return nil, err
	}

	return &camera, nil
}

func (r *CameraRepository) GetByProjectID(projectID int64) ([]models.Camera, error) {
	var cameras []models.Camera
	query := `SELECT * FROM cameras WHERE project_id = ? ORDER BY created_at ASC`

	err := r.db.Select(&cameras, query, projectID)
	if err != nil {
		return nil, err
	}

	if cameras == nil {
		cameras = []models.Camera{}
	}

	return cameras, nil
}

func (r *CameraRepository) Update(camera *models.Camera) error {
	query := `UPDATE cameras SET name = ?, location = ?, stream_url = ?, is_recording = ?, updated_at = ? WHERE id = ?`
	result, err := r.db.Exec(query, camera.Name, camera.Location, camera.StreamURL, camera.IsRecording, time.Now(), camera.ID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrCameraNotFound
	}

	return nil
}

func (r *CameraRepository) Delete(id int64) error {
	query := `DELETE FROM cameras WHERE id = ?`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrCameraNotFound
	}

	return nil
}

func (r *CameraRepository) BelongsToProject(cameraID, projectID int64) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM cameras WHERE id = ? AND project_id = ?`
	err := r.db.Get(&count, query, cameraID, projectID)
	return count > 0, err
}
