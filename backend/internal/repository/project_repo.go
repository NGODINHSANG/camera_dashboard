package repository

import (
	"camera-dashboard-backend/internal/models"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
)

var ErrProjectNotFound = errors.New("project not found")

type ProjectRepository struct {
	db *sqlx.DB
}

func NewProjectRepository(db *sqlx.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) Create(project *models.Project) error {
	query := `INSERT INTO projects (user_id, name, created_at, updated_at) VALUES (?, ?, ?, ?)`

	now := time.Now()
	result, err := r.db.Exec(query, project.UserID, project.Name, now, now)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	project.ID = id
	project.CreatedAt = now
	project.UpdatedAt = now
	project.Cameras = []models.Camera{}
	return nil
}

func (r *ProjectRepository) GetByID(id int64) (*models.Project, error) {
	var project models.Project
	query := `SELECT * FROM projects WHERE id = ?`

	err := r.db.Get(&project, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}

	return &project, nil
}

func (r *ProjectRepository) GetByUserID(userID int64) ([]models.Project, error) {
	var projects []models.Project
	query := `SELECT * FROM projects WHERE user_id = ? ORDER BY created_at DESC`

	err := r.db.Select(&projects, query, userID)
	if err != nil {
		return nil, err
	}

	if projects == nil {
		projects = []models.Project{}
	}

	return projects, nil
}

func (r *ProjectRepository) GetAll() ([]models.Project, error) {
	var projects []models.Project
	query := `SELECT * FROM projects ORDER BY created_at DESC`

	err := r.db.Select(&projects, query)
	if err != nil {
		return nil, err
	}

	if projects == nil {
		projects = []models.Project{}
	}

	return projects, nil
}

func (r *ProjectRepository) Update(project *models.Project) error {
	query := `UPDATE projects SET name = ?, updated_at = ? WHERE id = ?`
	result, err := r.db.Exec(query, project.Name, time.Now(), project.ID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrProjectNotFound
	}

	return nil
}

func (r *ProjectRepository) Delete(id int64) error {
	query := `DELETE FROM projects WHERE id = ?`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrProjectNotFound
	}

	return nil
}

func (r *ProjectRepository) BelongsToUser(projectID, userID int64) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM projects WHERE id = ? AND user_id = ?`
	err := r.db.Get(&count, query, projectID, userID)
	return count > 0, err
}
