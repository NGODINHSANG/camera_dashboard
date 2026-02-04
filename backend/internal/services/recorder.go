package services

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"camera-dashboard-backend/internal/models"

	"github.com/jmoiron/sqlx"
)

type ffmpegProcess struct {
	cmd   *exec.Cmd
	stdin io.WriteCloser
}

type RecorderService struct {
	db        *sqlx.DB
	processes map[int64]*ffmpegProcess // map[recordingID]*ffmpegProcess
	mu        sync.RWMutex
}

func NewRecorderService(db *sqlx.DB) *RecorderService {
	svc := &RecorderService{
		db:        db,
		processes: make(map[int64]*ffmpegProcess),
	}
	// Cleanup any orphaned recordings from previous runs
	svc.CleanupOrphanedRecordings()
	return svc
}

// CleanupOrphanedRecordings marks any recordings stuck in "recording" status as failed
func (r *RecorderService) CleanupOrphanedRecordings() {
	result, err := r.db.Exec(`
		UPDATE recordings SET status = ?, stopped_at = CURRENT_TIMESTAMP
		WHERE status = ?
	`, models.RecordingStatusFailed, models.RecordingStatusRecording)
	if err != nil {
		log.Printf("[Recorder] Failed to cleanup orphaned recordings: %v", err)
		return
	}
	if rows, _ := result.RowsAffected(); rows > 0 {
		log.Printf("[Recorder] Cleaned up %d orphaned recordings from previous session", rows)
	}
}

// sanitizeFilename removes invalid characters from filename
func sanitizeFilename(name string) string {
	name = strings.ReplaceAll(name, " ", "_")
	reg := regexp.MustCompile(`[<>:"/\\|?*]`)
	name = reg.ReplaceAllString(name, "")
	return name
}

// StartRecording starts FFmpeg process to record HLS stream
func (r *RecorderService) StartRecording(cameraID int64, projectID int64, projectName, cameraName, streamURL, outputDir string) (*models.Recording, error) {
	// Convert relative URL to absolute URL for FFmpeg
	if strings.HasPrefix(streamURL, "/") {
		// Relative URL - prepend MediaMTX base URL
		// Use environment variable or default to local/docker
		mediamtxHost := os.Getenv("MEDIAMTX_HOST")
		if mediamtxHost == "" {
			mediamtxHost = "http://camera-mediamtx:8888" // Docker default
		}

		// Strip /hls/ prefix if present - nginx proxies /hls/ to MediaMTX
		// but MediaMTX serves streams directly at /<stream_name>/
		if strings.HasPrefix(streamURL, "/hls/") {
			streamURL = strings.TrimPrefix(streamURL, "/hls")
		}

		streamURL = mediamtxHost + streamURL
		log.Printf("[Recorder] Converted relative URL to: %s", streamURL)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate filename: {project}_{camera}_{timestamp}.mp4
	timestamp := time.Now().Format("2006-01-02_150405")
	filename := fmt.Sprintf("%s_%s_%s.mp4",
		sanitizeFilename(projectName),
		sanitizeFilename(cameraName),
		timestamp,
	)
	filePath := filepath.Join(outputDir, filename)

	// Create recording record in DB
	now := time.Now()
	result, err := r.db.Exec(`
		INSERT INTO recordings (camera_id, project_id, filename, file_path, output_dir, status, started_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, cameraID, projectID, filename, filePath, outputDir, models.RecordingStatusRecording, now)
	if err != nil {
		return nil, fmt.Errorf("failed to create recording record: %w", err)
	}

	recordingID, _ := result.LastInsertId()

	// Start FFmpeg process with stdin pipe for graceful stop
	cmd := exec.Command("ffmpeg",
		"-i", streamURL,
		"-c", "copy",
		"-bsf:a", "aac_adtstoasc",
		"-movflags", "+faststart",
		"-y",
		filePath,
	)

	// Create stdin pipe to send 'q' for graceful stop
	stdin, err := cmd.StdinPipe()
	if err != nil {
		r.db.Exec("UPDATE recordings SET status = ? WHERE id = ?", models.RecordingStatusFailed, recordingID)
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	// Start process
	if err := cmd.Start(); err != nil {
		r.db.Exec("UPDATE recordings SET status = ? WHERE id = ?", models.RecordingStatusFailed, recordingID)
		return nil, fmt.Errorf("failed to start FFmpeg: %w", err)
	}

	// Store process reference with stdin
	r.mu.Lock()
	r.processes[recordingID] = &ffmpegProcess{cmd: cmd, stdin: stdin}
	r.mu.Unlock()

	log.Printf("[Recorder] Started recording ID=%d, Camera=%s, File=%s", recordingID, cameraName, filePath)

	// Monitor process in background
	go r.monitorProcess(recordingID, cmd)

	return &models.Recording{
		ID:        recordingID,
		CameraID:  cameraID,
		ProjectID: projectID,
		Filename:  filename,
		FilePath:  filePath,
		OutputDir: outputDir,
		Status:    models.RecordingStatusRecording,
		StartedAt: now,
	}, nil
}

// monitorProcess monitors FFmpeg process and updates status when done
func (r *RecorderService) monitorProcess(recordingID int64, cmd *exec.Cmd) {
	err := cmd.Wait()

	r.mu.Lock()
	delete(r.processes, recordingID)
	r.mu.Unlock()

	now := time.Now()
	status := models.RecordingStatusCompleted
	if err != nil {
		// Exit code 255 means graceful quit via 'q'
		if cmd.ProcessState != nil && (cmd.ProcessState.ExitCode() == 255 || cmd.ProcessState.ExitCode() == 0) {
			status = models.RecordingStatusCompleted
		} else {
			status = models.RecordingStatusFailed
			log.Printf("[Recorder] Recording ID=%d failed: %v (exit code: %d)", recordingID, err, cmd.ProcessState.ExitCode())
		}
	}

	// Get file size
	var fileSize int64 = 0
	var filePath string
	r.db.Get(&filePath, "SELECT file_path FROM recordings WHERE id = ?", recordingID)
	if info, err := os.Stat(filePath); err == nil {
		fileSize = info.Size()
	}

	// Update recording status
	r.db.Exec(`
		UPDATE recordings SET status = ?, stopped_at = ?, file_size = ?
		WHERE id = ?
	`, status, now, fileSize, recordingID)

	log.Printf("[Recorder] Recording ID=%d completed, Status=%s, Size=%d bytes", recordingID, status, fileSize)
}

// StopRecording stops a running recording gracefully
func (r *RecorderService) StopRecording(recordingID int64) error {
	r.mu.RLock()
	proc, exists := r.processes[recordingID]
	r.mu.RUnlock()

	if !exists {
		return fmt.Errorf("recording not found or already stopped")
	}

	// Send 'q' to FFmpeg stdin for graceful shutdown (works on Windows)
	log.Printf("[Recorder] Sending 'q' to stop recording ID=%d", recordingID)
	_, err := proc.stdin.Write([]byte("q"))
	if err != nil {
		log.Printf("[Recorder] Failed to send 'q', force killing: %v", err)
		proc.cmd.Process.Kill()
	}
	proc.stdin.Close()

	return nil
}

// GetRecording gets a recording by ID
func (r *RecorderService) GetRecording(recordingID int64) (*models.Recording, error) {
	var recording models.Recording
	err := r.db.Get(&recording, `
		SELECT r.*, c.name as camera_name, p.name as project_name
		FROM recordings r
		JOIN cameras c ON r.camera_id = c.id
		JOIN projects p ON r.project_id = p.id
		WHERE r.id = ?
	`, recordingID)
	if err != nil {
		return nil, err
	}
	return &recording, nil
}

// GetRecordingsByCamera gets all recordings for a camera
func (r *RecorderService) GetRecordingsByCamera(cameraID int64) ([]models.Recording, error) {
	var recordings []models.Recording
	err := r.db.Select(&recordings, `
		SELECT r.*, c.name as camera_name, p.name as project_name
		FROM recordings r
		JOIN cameras c ON r.camera_id = c.id
		JOIN projects p ON r.project_id = p.id
		WHERE r.camera_id = ?
		ORDER BY r.created_at DESC
	`, cameraID)
	return recordings, err
}

// GetAllRecordings gets all recordings
func (r *RecorderService) GetAllRecordings() ([]models.Recording, error) {
	var recordings []models.Recording
	err := r.db.Select(&recordings, `
		SELECT r.*, c.name as camera_name, p.name as project_name
		FROM recordings r
		JOIN cameras c ON r.camera_id = c.id
		JOIN projects p ON r.project_id = p.id
		ORDER BY r.created_at DESC
	`)
	return recordings, err
}

// GetActiveRecordings gets all currently running recordings
func (r *RecorderService) GetActiveRecordings() ([]models.Recording, error) {
	var recordings []models.Recording
	err := r.db.Select(&recordings, `
		SELECT r.*, c.name as camera_name, p.name as project_name
		FROM recordings r
		JOIN cameras c ON r.camera_id = c.id
		JOIN projects p ON r.project_id = p.id
		WHERE r.status = ?
		ORDER BY r.started_at DESC
	`, models.RecordingStatusRecording)
	return recordings, err
}

// DeleteRecording deletes a recording and its file
func (r *RecorderService) DeleteRecording(recordingID int64) error {
	var filePath string
	err := r.db.Get(&filePath, "SELECT file_path FROM recordings WHERE id = ?", recordingID)
	if err != nil {
		return err
	}

	os.Remove(filePath)
	_, err = r.db.Exec("DELETE FROM recordings WHERE id = ?", recordingID)
	return err
}

// IsRecording checks if a camera is currently being recorded
func (r *RecorderService) IsRecording(cameraID int64) (bool, int64) {
	var recordingID int64
	err := r.db.Get(&recordingID, `
		SELECT id FROM recordings 
		WHERE camera_id = ? AND status = ?
	`, cameraID, models.RecordingStatusRecording)
	if err != nil {
		return false, 0
	}
	return true, recordingID
}
