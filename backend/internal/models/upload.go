package models

import "time"

// UploadJob represents a video upload job
type UploadJob struct {
	ID               string     `db:"id" json:"id"`
	CameraID         int64      `db:"camera_id" json:"camera_id"`
	ProjectID        int64      `db:"project_id" json:"project_id"`
	Filename         string     `db:"filename" json:"filename"`
	OriginalFilename string     `db:"original_filename" json:"original_filename"`
	Status           string     `db:"status" json:"status"` // uploading, processing, completed, failed
	Progress         int        `db:"progress" json:"progress"`
	FileSize         int64      `db:"file_size" json:"file_size"`
	Error            *string    `db:"error" json:"error,omitempty"`
	CreatedAt        time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at" json:"updated_at"`
}

// UploadJobStatus constants
const (
	UploadStatusUploading  = "uploading"
	UploadStatusProcessing = "processing"
	UploadStatusCompleted  = "completed"
	UploadStatusFailed     = "failed"
)

// RecordingSource constants
const (
	RecordingSourceStream = "stream"
	RecordingSourceUpload = "upload"
)
