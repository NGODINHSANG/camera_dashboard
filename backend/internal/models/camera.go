package models

import "time"

type Camera struct {
	ID          int64     `db:"id" json:"id"`
	ProjectID   int64     `db:"project_id" json:"projectId"`
	Name        string    `db:"name" json:"name"`
	Location    string    `db:"location" json:"location"`
	StreamURL   string    `db:"stream_url" json:"streamUrl"`
	IsRecording bool      `db:"is_recording" json:"isRecording"`
	CreatedAt   time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt   time.Time `db:"updated_at" json:"updatedAt"`
}

type CreateCameraRequest struct {
	Name        string `json:"name"`
	Location    string `json:"location"`
	StreamURL   string `json:"streamUrl"`
	IsRecording bool   `json:"isRecording"`
}

type UpdateCameraRequest struct {
	Name        string `json:"name"`
	Location    string `json:"location"`
	StreamURL   string `json:"streamUrl"`
	IsRecording bool   `json:"isRecording"`
}
