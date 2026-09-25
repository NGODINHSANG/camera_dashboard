package models

import "time"

type ProjectPermission struct {
	ID        int64     `db:"id" json:"id"`
	UserID    int64     `db:"user_id" json:"user_id"`
	ProjectID int64     `db:"project_id" json:"project_id"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type ProjectMember struct {
	UserID    int64     `db:"user_id" json:"user_id"`
	Name      string    `db:"name" json:"name"`
	Email     string    `db:"email" json:"email"`
	GrantedAt time.Time `db:"granted_at" json:"granted_at"`
}
