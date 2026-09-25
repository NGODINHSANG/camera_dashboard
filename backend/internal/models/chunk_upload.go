package models

import "time"

// ChunkUpload represents a chunked upload session
type ChunkUpload struct {
	ID               string    `db:"id" json:"id"`
	CameraID         int64     `db:"camera_id" json:"camera_id"`
	ProjectID        int64     `db:"project_id" json:"project_id"`
	OriginalFilename string    `db:"original_filename" json:"original_filename"`
	TotalSize        int64     `db:"total_size" json:"total_size"`
	TotalChunks      int       `db:"total_chunks" json:"total_chunks"`
	UploadedChunks   int       `db:"uploaded_chunks" json:"uploaded_chunks"`
	ChunkSize        int64     `db:"chunk_size" json:"chunk_size"`
	Status           string    `db:"status" json:"status"` // uploading, merging, processing, completed, failed
	Error            *string   `db:"error" json:"error,omitempty"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time `db:"updated_at" json:"updated_at"`
}

// ChunkUploadStatus constants
const (
	ChunkStatusUploading  = "uploading"
	ChunkStatusMerging    = "merging"
	ChunkStatusProcessing = "processing"
	ChunkStatusCompleted  = "completed"
	ChunkStatusFailed     = "failed"
)
