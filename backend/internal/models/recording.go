package models

import "time"

type RecordingStatus string

const (
	RecordingStatusRecording RecordingStatus = "recording"
	RecordingStatusCompleted RecordingStatus = "completed"
	RecordingStatusFailed    RecordingStatus = "failed"
)

type Recording struct {
	ID        int64           `db:"id" json:"id"`
	CameraID  int64           `db:"camera_id" json:"cameraId"`
	ProjectID int64           `db:"project_id" json:"projectId"`
	Filename  string          `db:"filename" json:"filename"`
	FilePath  string          `db:"file_path" json:"filePath"`
	OutputDir string          `db:"output_dir" json:"outputDir"`
	Status    RecordingStatus `db:"status" json:"status"`
	FileSize  int64           `db:"file_size" json:"fileSize"`
	StartedAt time.Time       `db:"started_at" json:"startedAt"`
	StoppedAt *time.Time      `db:"stopped_at" json:"stoppedAt"`
	CreatedAt time.Time       `db:"created_at" json:"createdAt"`
	// Joined fields
	CameraName  string `db:"camera_name" json:"cameraName,omitempty"`
	ProjectName string `db:"project_name" json:"projectName,omitempty"`
	StreamURL   string `db:"stream_url" json:"streamUrl,omitempty"`
}

type StartRecordingRequest struct {
	CameraID  int64  `json:"cameraId"`
	OutputDir string `json:"outputDir"` // Thư mục lưu video do user chọn
}

type StopRecordingRequest struct {
	RecordingID int64 `json:"recordingId"`
}

type RecordingResponse struct {
	Recording
}
