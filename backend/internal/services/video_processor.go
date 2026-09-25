package services

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"camera-dashboard-backend/internal/models"
	"camera-dashboard-backend/internal/repository"

	"github.com/jmoiron/sqlx"
)

const (
	// Video encoding settings
	VideoBitrate = "2M"      // 2 Mbps
	MaxBitrate   = "2.5M"    // Max bitrate for VBR
	BufferSize   = "4M"      // Buffer size
	AudioBitrate = "128k"    // Audio bitrate
)

type VideoProcessor struct {
	db            *sqlx.DB
	uploadRepo    *repository.UploadRepository
	recordingsPath string
	tempPath      string
}

func NewVideoProcessor(db *sqlx.DB, recordingsPath string) *VideoProcessor {
	tempPath := filepath.Join(recordingsPath, ".tmp_uploads")
	os.MkdirAll(tempPath, 0755)

	return &VideoProcessor{
		db:            db,
		uploadRepo:    repository.NewUploadRepository(db),
		recordingsPath: recordingsPath,
		tempPath:      tempPath,
	}
}

// GetTempPath returns the temporary upload path
func (vp *VideoProcessor) GetTempPath() string {
	return vp.tempPath
}

// ProcessVideo compresses video and moves to final destination
func (vp *VideoProcessor) ProcessVideo(job *models.UploadJob, projectName, cameraName string) error {
	log.Printf("[VideoProcessor] Starting processing job %s: %s", job.ID, job.OriginalFilename)

	// Update status to processing
	vp.uploadRepo.UpdateStatus(job.ID, models.UploadStatusProcessing, 0)

	// Input file path (in temp directory)
	inputPath := filepath.Join(vp.tempPath, job.Filename)

	// Check if input file exists
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		errMsg := fmt.Sprintf("Input file not found: %s", inputPath)
		vp.uploadRepo.UpdateError(job.ID, errMsg)
		return fmt.Errorf(errMsg)
	}

	// Get video duration for progress calculation
	duration, err := vp.getVideoDuration(inputPath)
	if err != nil {
		log.Printf("[VideoProcessor] Warning: Could not get duration: %v", err)
		duration = 0
	}

	// Create output directory (use original names to match existing folder structure)
	outputDir := filepath.Join(vp.recordingsPath, projectName, cameraName)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		errMsg := fmt.Sprintf("Failed to create output directory: %v", err)
		vp.uploadRepo.UpdateError(job.ID, errMsg)
		return fmt.Errorf(errMsg)
	}

	// Generate output filename (sanitize for safe filename)
	timestamp := time.Now().Format("2006-01-02_150405")
	outputFilename := fmt.Sprintf("%s_%s_upload_%s.mp4",
		sanitizeName(projectName),
		sanitizeName(cameraName),
		timestamp,
	)
	outputPath := filepath.Join(outputDir, outputFilename)

	// Build FFmpeg command
	args := []string{
		"-i", inputPath,
		"-c:v", "libx264",
		"-b:v", VideoBitrate,
		"-maxrate", MaxBitrate,
		"-bufsize", BufferSize,
		"-preset", "fast",
		"-c:a", "aac",
		"-b:a", AudioBitrate,
		"-movflags", "+faststart",
		"-progress", "pipe:1",
		"-y",
		outputPath,
	}

	log.Printf("[VideoProcessor] Running FFmpeg: ffmpeg %s", strings.Join(args, " "))

	cmd := exec.Command("ffmpeg", args...)

	// Capture stdout for progress
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		errMsg := fmt.Sprintf("Failed to create stdout pipe: %v", err)
		vp.uploadRepo.UpdateError(job.ID, errMsg)
		return fmt.Errorf(errMsg)
	}

	// Start command
	if err := cmd.Start(); err != nil {
		errMsg := fmt.Sprintf("Failed to start FFmpeg: %v", err)
		vp.uploadRepo.UpdateError(job.ID, errMsg)
		return fmt.Errorf(errMsg)
	}

	// Parse progress from stdout
	go vp.parseProgress(job.ID, stdout, duration)

	// Wait for completion
	if err := cmd.Wait(); err != nil {
		errMsg := fmt.Sprintf("FFmpeg failed: %v", err)
		vp.uploadRepo.UpdateError(job.ID, errMsg)
		// Cleanup failed output
		os.Remove(outputPath)
		return fmt.Errorf(errMsg)
	}

	// Get output file size
	outputInfo, err := os.Stat(outputPath)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to stat output file: %v", err)
		vp.uploadRepo.UpdateError(job.ID, errMsg)
		return fmt.Errorf(errMsg)
	}

	// Insert into recordings table
	if err := vp.insertRecording(job, projectName, cameraName, outputFilename, outputPath, outputDir, outputInfo.Size()); err != nil {
		errMsg := fmt.Sprintf("Failed to insert recording: %v", err)
		vp.uploadRepo.UpdateError(job.ID, errMsg)
		return fmt.Errorf(errMsg)
	}

	// Update job status to completed
	vp.uploadRepo.UpdateStatus(job.ID, models.UploadStatusCompleted, 100)

	// Remove temp file
	os.Remove(inputPath)

	log.Printf("[VideoProcessor] Completed job %s: %s -> %s", job.ID, job.OriginalFilename, outputPath)
	return nil
}

// getVideoDuration gets video duration in seconds using ffprobe
func (vp *VideoProcessor) getVideoDuration(inputPath string) (float64, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		inputPath,
	)

	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	duration, err := strconv.ParseFloat(strings.TrimSpace(string(output)), 64)
	if err != nil {
		return 0, err
	}

	return duration, nil
}

// parseProgress parses FFmpeg progress output and updates job status
func (vp *VideoProcessor) parseProgress(jobID string, stdout io.Reader, totalDuration float64) {
	scanner := bufio.NewScanner(stdout)
	timeRegex := regexp.MustCompile(`out_time_ms=(\d+)`)

	for scanner.Scan() {
		line := scanner.Text()

		if totalDuration > 0 {
			matches := timeRegex.FindStringSubmatch(line)
			if len(matches) > 1 {
				timeMs, _ := strconv.ParseInt(matches[1], 10, 64)
				currentSec := float64(timeMs) / 1000000.0
				progress := int((currentSec / totalDuration) * 100)
				if progress > 100 {
					progress = 100
				}
				if progress > 0 {
					vp.uploadRepo.UpdateStatus(jobID, models.UploadStatusProcessing, progress)
				}
			}
		}
	}
}

// insertRecording inserts a new recording into the database
func (vp *VideoProcessor) insertRecording(job *models.UploadJob, projectName, cameraName, filename, filePath, outputDir string, fileSize int64) error {
	query := `
		INSERT INTO recordings (camera_id, project_id, filename, file_path, output_dir, status, file_size, source, started_at, stopped_at, created_at)
		VALUES (?, ?, ?, ?, ?, 'completed', ?, 'upload', ?, ?, ?)
	`
	now := time.Now()
	_, err := vp.db.Exec(query,
		job.CameraID,
		job.ProjectID,
		filename,
		filePath,
		outputDir,
		fileSize,
		now,
		now,
		now,
	)
	return err
}

// sanitizeName makes a string safe for use in file paths
func sanitizeName(name string) string {
	// Replace spaces with underscores
	name = strings.ReplaceAll(name, " ", "_")
	// Remove special characters
	reg := regexp.MustCompile(`[^a-zA-Z0-9_\-]`)
	name = reg.ReplaceAllString(name, "")
	return name
}
