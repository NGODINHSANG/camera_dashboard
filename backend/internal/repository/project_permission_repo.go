package repository

import (
	"camera-dashboard-backend/internal/models"

	"github.com/jmoiron/sqlx"
)

type ProjectPermissionRepository struct {
	db *sqlx.DB
}

func NewProjectPermissionRepository(db *sqlx.DB) *ProjectPermissionRepository {
	return &ProjectPermissionRepository{db: db}
}

func (r *ProjectPermissionRepository) Grant(userID, projectID int64) error {
	_, err := r.db.Exec(
		`INSERT OR IGNORE INTO project_permissions (user_id, project_id) VALUES (?, ?)`,
		userID, projectID,
	)
	return err
}

func (r *ProjectPermissionRepository) Revoke(userID, projectID int64) error {
	_, err := r.db.Exec(
		`DELETE FROM project_permissions WHERE user_id = ? AND project_id = ?`,
		userID, projectID,
	)
	return err
}

func (r *ProjectPermissionRepository) HasPermission(userID, projectID int64) bool {
	var count int
	err := r.db.Get(&count,
		`SELECT COUNT(*) FROM project_permissions WHERE user_id = ? AND project_id = ?`,
		userID, projectID,
	)
	return err == nil && count > 0
}

func (r *ProjectPermissionRepository) GetProjectMembers(projectID int64) ([]models.ProjectMember, error) {
	var members []models.ProjectMember
	err := r.db.Select(&members, `
		SELECT u.id as user_id, u.name, u.email, pp.created_at as granted_at
		FROM project_permissions pp
		JOIN users u ON u.id = pp.user_id
		WHERE pp.project_id = ?
		ORDER BY pp.created_at DESC
	`, projectID)
	return members, err
}

func (r *ProjectPermissionRepository) GetUserProjectIDs(userID int64) ([]int64, error) {
	var ids []int64
	err := r.db.Select(&ids,
		`SELECT project_id FROM project_permissions WHERE user_id = ?`,
		userID,
	)
	if ids == nil {
		ids = []int64{}
	}
	return ids, err
}
