package handlers

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"camera-dashboard-backend/internal/models"
	"camera-dashboard-backend/internal/services"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
)

type RecordingHandler struct {
	db             *sqlx.DB
	recorder       *services.RecorderService
	recordingsPath string
}

func NewRecordingHandler(db *sqlx.DB, recorder *services.RecorderService, recordingsPath string) *RecordingHandler {
	return &RecordingHandler{
		db:             db,
		recorder:       recorder,
		recordingsPath: recordingsPath,
	}
}

// Helper to write JSON response
func (h *RecordingHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// Helper to write error response
func (h *RecordingHandler) writeError(w http.ResponseWriter, status int, message string) {
	h.writeJSON(w, status, map[string]string{"error": message})
}

// sanitizeName removes or replaces characters that are not safe for directory names
func sanitizeName(name string) string {
	// Replace unsafe characters with underscore
	unsafe := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	result := name
	for _, char := range unsafe {
		result = strings.ReplaceAll(result, char, "_")
	}
	// Trim spaces
	result = strings.TrimSpace(result)
	if result == "" {
		result = "unknown"
	}
	return result
}

// StartRecording starts recording a camera
func (h *RecordingHandler) StartRecording(w http.ResponseWriter, r *http.Request) {
	var req models.StartRecordingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	if req.CameraID == 0 {
		h.writeError(w, http.StatusBadRequest, "Camera ID is required")
		return
	}

	// Check if camera is already being recorded
	if isRecording, recordingID := h.recorder.IsRecording(req.CameraID); isRecording {
		h.writeJSON(w, http.StatusConflict, map[string]interface{}{
			"error":       "Camera is already being recorded",
			"recordingId": recordingID,
		})
		return
	}

	// Get camera and project info
	var camera struct {
		ID        int64  `db:"id"`
		Name      string `db:"name"`
		StreamURL string `db:"stream_url"`
		ProjectID int64  `db:"project_id"`
	}
	err := h.db.Get(&camera, "SELECT id, name, stream_url, project_id FROM cameras WHERE id = ?", req.CameraID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Camera not found")
		return
	}

	var projectName string
	err = h.db.Get(&projectName, "SELECT name FROM projects WHERE id = ?", camera.ProjectID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Project not found")
		return
	}

	// Auto-generate output directory: /recordings/<project_name>/<camera_name>
	outputDir := filepath.Join(h.recordingsPath, sanitizeName(projectName), sanitizeName(camera.Name))

	// Create directory if not exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		h.writeError(w, http.StatusInternalServerError, "Cannot create recording directory: "+err.Error())
		return
	}

	// Start recording
	recording, err := h.recorder.StartRecording(
		camera.ID,
		camera.ProjectID,
		projectName,
		camera.Name,
		camera.StreamURL,
		outputDir,
	)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, recording)
}

// StopRecording stops a running recording
func (h *RecordingHandler) StopRecording(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid recording ID")
		return
	}

	if err := h.recorder.StopRecording(id); err != nil {
		h.writeError(w, http.StatusNotFound, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Recording stopped",
		"id":      id,
	})
}

// GetRecordings gets all recordings
func (h *RecordingHandler) GetRecordings(w http.ResponseWriter, r *http.Request) {
	recordings, err := h.recorder.GetAllRecordings()
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, recordings)
}

// GetRecording gets a recording by ID
func (h *RecordingHandler) GetRecording(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid recording ID")
		return
	}

	recording, err := h.recorder.GetRecording(id)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Recording not found")
		return
	}

	h.writeJSON(w, http.StatusOK, recording)
}

// GetActiveRecordings gets all active recordings
func (h *RecordingHandler) GetActiveRecordings(w http.ResponseWriter, r *http.Request) {
	recordings, err := h.recorder.GetActiveRecordings()
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, recordings)
}

// GetCameraRecordingStatus checks if a camera is currently recording
func (h *RecordingHandler) GetCameraRecordingStatus(w http.ResponseWriter, r *http.Request) {
	cameraIDStr := chi.URLParam(r, "cameraId")
	cameraID, err := strconv.ParseInt(cameraIDStr, 10, 64)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid camera ID")
		return
	}

	isRecording, recordingID := h.recorder.IsRecording(cameraID)
	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"isRecording": isRecording,
		"recordingId": recordingID,
	})
}

// DeleteRecording deletes a recording
func (h *RecordingHandler) DeleteRecording(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid recording ID")
		return
	}

	if err := h.recorder.DeleteRecording(id); err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "Recording deleted"})
}

// DownloadRecording downloads a recording file
func (h *RecordingHandler) DownloadRecording(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid recording ID")
		return
	}

	recording, err := h.recorder.GetRecording(id)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Recording not found")
		return
	}

	// Check if file exists
	if _, err := os.Stat(recording.FilePath); os.IsNotExist(err) {
		h.writeError(w, http.StatusNotFound, "Recording file not found")
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+recording.Filename)
	w.Header().Set("Content-Type", "video/mp4")
	http.ServeFile(w, r, recording.FilePath)
}

// StreamRecording streams a recording file with HTTP Range support for seeking
func (h *RecordingHandler) StreamRecording(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid recording ID")
		return
	}

	recording, err := h.recorder.GetRecording(id)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Recording not found")
		return
	}

	// Check if file exists
	if _, err := os.Stat(recording.FilePath); os.IsNotExist(err) {
		h.writeError(w, http.StatusNotFound, "Recording file not found")
		return
	}

	// Set content type for streaming (no Content-Disposition = inline playback)
	w.Header().Set("Content-Type", "video/mp4")
	// http.ServeFile automatically handles Range requests for seeking
	http.ServeFile(w, r, recording.FilePath)
}

// GetCameraRecordings gets all completed recordings for a specific camera (DB-based - legacy)
func (h *RecordingHandler) GetCameraRecordings(w http.ResponseWriter, r *http.Request) {
	cameraIDStr := chi.URLParam(r, "cameraId")
	cameraID, err := strconv.ParseInt(cameraIDStr, 10, 64)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid camera ID")
		return
	}

	recordings, err := h.recorder.GetRecordingsByCamera(cameraID, string(models.RecordingStatusCompleted))
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, recordings)
}

// VideoFileInfo represents info about a video file
type VideoFileInfo struct {
	Name     string    `json:"name"`
	Path     string    `json:"path"`
	Size     int64     `json:"size"`
	ModTime  time.Time `json:"modTime"`
	Duration string    `json:"duration,omitempty"`
	HLSUrl   string    `json:"hlsUrl,omitempty"` // HLS playlist URL if available
	HasHLS   bool      `json:"hasHls"`
}

// ListCameraRecordingFiles lists all video files in the camera's recording directory
func (h *RecordingHandler) ListCameraRecordingFiles(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "projectName")
	cameraName := chi.URLParam(r, "cameraName")

	if projectName == "" || cameraName == "" {
		h.writeError(w, http.StatusBadRequest, "Project name and camera name are required")
		return
	}

	// Build the directory path
	dirPath := filepath.Join(h.recordingsPath, sanitizeName(projectName), sanitizeName(cameraName))

	// Check if directory exists
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		// Return empty array if directory doesn't exist (no recordings yet)
		h.writeJSON(w, http.StatusOK, []VideoFileInfo{})
		return
	}

	// Read directory
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Cannot read directory: "+err.Error())
		return
	}

	// Filter video files and get info
	var files []VideoFileInfo
	videoExtensions := []string{".mp4", ".mkv", ".avi", ".mov", ".webm", ".ts"}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))

		isVideo := false
		for _, ve := range videoExtensions {
			if ext == ve {
				isVideo = true
				break
			}
		}

		if !isVideo {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		fullPath := filepath.Join(dirPath, name)
		fileInfo := VideoFileInfo{
			Name:    name,
			Path:    fullPath,
			Size:    info.Size(),
			ModTime: info.ModTime(),
		}

		// Check if HLS version exists
		if services.HasHLS(fullPath) {
			fileInfo.HasHLS = true
			// Build HLS URL: /api/recordings/hls?path=<hls_dir>/playlist.m3u8
			playlist := services.GetHLSPlaylist(fullPath)
			fileInfo.HLSUrl = "/api/recordings/hls?path=" + playlist
		}

		files = append(files, fileInfo)
	}

	// Sort by modification time (newest first)
	for i := 0; i < len(files)-1; i++ {
		for j := i + 1; j < len(files); j++ {
			if files[j].ModTime.After(files[i].ModTime) {
				files[i], files[j] = files[j], files[i]
			}
		}
	}

	h.writeJSON(w, http.StatusOK, files)
}

// StreamVideoFile streams a video file by path (with security check)
func (h *RecordingHandler) StreamVideoFile(w http.ResponseWriter, r *http.Request) {
	// Get file path from query parameter
	filePath := r.URL.Query().Get("path")
	if filePath == "" {
		h.writeError(w, http.StatusBadRequest, "File path is required")
		return
	}

	// Security: ensure the file is within the recordings directory
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid file path")
		return
	}

	absRecordingsPath, _ := filepath.Abs(h.recordingsPath)
	if !strings.HasPrefix(absPath, absRecordingsPath) {
		h.writeError(w, http.StatusForbidden, "Access denied: file is outside recordings directory")
		return
	}

	// Check if file exists and get file info
	fileInfo, err := os.Stat(absPath)
	if os.IsNotExist(err) {
		h.writeError(w, http.StatusNotFound, "Video file not found")
		return
	}
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to get file info")
		return
	}

	// Set cache headers to prevent re-fetching
	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("Cache-Control", "public, max-age=86400") // Cache for 24 hours
	w.Header().Set("Accept-Ranges", "bytes")

	// Use ServeContent for proper range request support with caching
	file, err := os.Open(absPath)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to open file")
		return
	}
	defer file.Close()

	http.ServeContent(w, r, fileInfo.Name(), fileInfo.ModTime(), file)
}

// DownloadVideoFile downloads a video file by path (with security check)
func (h *RecordingHandler) DownloadVideoFile(w http.ResponseWriter, r *http.Request) {
	// Get file path from query parameter
	filePath := r.URL.Query().Get("path")
	if filePath == "" {
		h.writeError(w, http.StatusBadRequest, "File path is required")
		return
	}

	// Security: ensure the file is within the recordings directory
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid file path")
		return
	}

	absRecordingsPath, _ := filepath.Abs(h.recordingsPath)
	if !strings.HasPrefix(absPath, absRecordingsPath) {
		h.writeError(w, http.StatusForbidden, "Access denied: file is outside recordings directory")
		return
	}

	// Check if file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		h.writeError(w, http.StatusNotFound, "Video file not found")
		return
	}

	// Download the video
	filename := filepath.Base(absPath)
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Header().Set("Content-Type", "video/mp4")
	http.ServeFile(w, r, absPath)
}

// DeleteVideoFile deletes a video file by path (with security check)
func (h *RecordingHandler) DeleteVideoFile(w http.ResponseWriter, r *http.Request) {
	// Get file path from query parameter
	filePath := r.URL.Query().Get("path")
	if filePath == "" {
		h.writeError(w, http.StatusBadRequest, "File path is required")
		return
	}

	// Security: ensure the file is within the recordings directory
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid file path")
		return
	}

	absRecordingsPath, _ := filepath.Abs(h.recordingsPath)
	if !strings.HasPrefix(absPath, absRecordingsPath) {
		h.writeError(w, http.StatusForbidden, "Access denied: file is outside recordings directory")
		return
	}

	// Check if file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		h.writeError(w, http.StatusNotFound, "Video file not found")
		return
	}

	// Delete the file
	if err := os.Remove(absPath); err != nil {
		h.writeError(w, http.StatusInternalServerError, "Cannot delete file: "+err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "File deleted successfully"})
}

// BrowseDirectory returns list of subdirectories in a given path
func (h *RecordingHandler) BrowseDirectory(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")

	// Clean and normalize the path
	path = strings.TrimSpace(path)

	// Log for debugging
	log.Printf("[BrowseDirectory] Received path: '%s', OS: %s", path, runtime.GOOS)

	isWindows := runtime.GOOS == "windows"

	// Handle empty path or root request
	if path == "" || path == "drives" {
		if isWindows {
			// On Windows, list available drives
			drives := []map[string]interface{}{}
			for _, drive := range "CDEFGHIJ" {
				drivePath := string(drive) + ":\\"
				if _, err := os.Stat(drivePath); err == nil {
					drives = append(drives, map[string]interface{}{
						"name": string(drive) + ":",
						"path": drivePath,
					})
				}
			}
			log.Printf("[BrowseDirectory] Returning Windows drives list: %d drives", len(drives))
			h.writeJSON(w, http.StatusOK, map[string]interface{}{
				"currentPath": "Ổ đĩa",
				"parentPath":  "",
				"directories": drives,
				"isDriveList": true,
			})
		} else {
			// On Linux/Unix, start from root or home
			rootPath := "/"
			if path == "" {
				// Default to mounted host directories and container paths
				dirs := []map[string]interface{}{}
				// /mnt/* are host directories mounted via docker-compose
				// /app/recordings is the default recording directory
				commonPaths := []string{"/mnt/home", "/mnt/opt", "/mnt/var", "/app/recordings"}
				for _, p := range commonPaths {
					if info, err := os.Stat(p); err == nil && info.IsDir() {
						dirs = append(dirs, map[string]interface{}{
							"name": filepath.Base(p),
							"path": p,
						})
					}
				}
				log.Printf("[BrowseDirectory] Returning Linux common paths: %d dirs", len(dirs))
				h.writeJSON(w, http.StatusOK, map[string]interface{}{
					"currentPath": "/",
					"parentPath":  "",
					"directories": dirs,
					"isRootList":  true,
				})
			} else {
				// Browse from root
				path = rootPath
			}
		}
		if path == "" || path == "drives" {
			return
		}
	}

	// On Windows, ensure proper path format
	if isWindows && len(path) >= 2 && path[1] == ':' {
		if len(path) == 2 {
			path = path + "\\"
		}
		path = filepath.Clean(path)
	}

	log.Printf("[BrowseDirectory] Normalized path: '%s'", path)

	// Security: check path is absolute
	if !filepath.IsAbs(path) {
		h.writeError(w, http.StatusBadRequest, "Path must be absolute")
		return
	}

	// Check if path exists and is a directory
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			h.writeError(w, http.StatusNotFound, "Directory not found")
		} else {
			h.writeError(w, http.StatusForbidden, "Cannot access directory")
		}
		return
	}
	if !info.IsDir() {
		h.writeError(w, http.StatusBadRequest, "Path is not a directory")
		return
	}

	// Read directory entries
	entries, err := os.ReadDir(path)
	if err != nil {
		h.writeError(w, http.StatusForbidden, "Cannot read directory")
		return
	}

	// Filter only directories
	dirs := []map[string]interface{}{}
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			dirs = append(dirs, map[string]interface{}{
				"name": entry.Name(),
				"path": filepath.Join(path, entry.Name()),
			})
		}
	}

	// Get parent path
	parent := filepath.Dir(path)
	if isWindows {
		// Windows: if at drive root, go to drives list
		if parent == path || len(path) <= 3 {
			parent = "drives"
		}
	} else {
		// Linux: if at root /, no parent
		if path == "/" || parent == path {
			parent = ""
		}
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"currentPath": path,
		"parentPath":  parent,
		"directories": dirs,
	})
}

// ValidatePath checks if a path is valid and writable
func (h *RecordingHandler) ValidatePath(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	if req.Path == "" {
		h.writeError(w, http.StatusBadRequest, "Path is required")
		return
	}

	// Security: check path is absolute
	if !filepath.IsAbs(req.Path) {
		h.writeError(w, http.StatusBadRequest, "Path must be absolute")
		return
	}

	// Try to create directory if not exists
	if err := os.MkdirAll(req.Path, 0755); err != nil {
		h.writeError(w, http.StatusBadRequest, "Cannot create directory: "+err.Error())
		return
	}

	// Try to write a test file
	testFile := filepath.Join(req.Path, ".write_test_"+strconv.FormatInt(time.Now().UnixNano(), 10))
	f, err := os.Create(testFile)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Directory is not writable: "+err.Error())
		return
	}
	f.Close()
	os.Remove(testFile)

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"valid":   true,
		"path":    req.Path,
		"message": "Directory is valid and writable",
	})
}

// LiveRestream creates HLS stream on-the-fly from a video file
// This allows instant playback without pre-conversion
// FFmpeg reads file and outputs HLS segments in real-time
func (h *RecordingHandler) LiveRestream(w http.ResponseWriter, r *http.Request) {
	filePath := r.URL.Query().Get("path")
	if filePath == "" {
		h.writeError(w, http.StatusBadRequest, "File path is required")
		return
	}

	// Security: ensure the file is within the recordings directory
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid file path")
		return
	}

	absRecordingsPath, _ := filepath.Abs(h.recordingsPath)
	if !strings.HasPrefix(absPath, absRecordingsPath) {
		h.writeError(w, http.StatusForbidden, "Access denied")
		return
	}

	// Check file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		h.writeError(w, http.StatusNotFound, "Video file not found")
		return
	}

	// Check if requesting playlist or segment
	segmentName := r.URL.Query().Get("segment")

	// Get or create restream session
	sessionID := fmt.Sprintf("%x", md5.Sum([]byte(absPath)))
	sessionDir := filepath.Join(os.TempDir(), "restream_"+sessionID)

	// First segment path - used to verify FFmpeg is working
	firstSegment := filepath.Join(sessionDir, "seg_00000.ts")
	playlistPath := filepath.Join(sessionDir, "stream.m3u8")

	// Check if session already exists AND has segments
	needsStart := false
	if _, err := os.Stat(firstSegment); os.IsNotExist(err) {
		needsStart = true
	}

	if needsStart {
		// Clean up old session if exists
		os.RemoveAll(sessionDir)

		// Start new FFmpeg restream process
		if err := os.MkdirAll(sessionDir, 0755); err != nil {
			h.writeError(w, http.StatusInternalServerError, "Failed to create session directory")
			return
		}

		segmentPattern := filepath.Join(sessionDir, "seg_%05d.ts")

		log.Printf("[Restream] Starting FFmpeg for: %s", filepath.Base(absPath))

		// FFmpeg: read file, output HLS on-the-fly with -c copy (fast, original quality)
		cmd := exec.Command("ffmpeg",
			"-i", absPath,
			"-c", "copy",
			"-hls_time", "4", // 4 second segments
			"-hls_list_size", "0", // Keep all segments in playlist
			"-hls_flags", "independent_segments", // Each segment can be decoded independently
			"-hls_segment_filename", segmentPattern,
			"-f", "hls",
			"-y",
			playlistPath,
		)

		// Capture stderr for debugging
		cmd.Stderr = os.Stderr

		// Start FFmpeg in background
		if err := cmd.Start(); err != nil {
			log.Printf("[Restream] FFmpeg start failed: %v", err)
			h.writeError(w, http.StatusInternalServerError, "Failed to start restream: "+err.Error())
			return
		}

		// Wait for FIRST SEGMENT FILE to be ready (max 10 seconds)
		// This ensures we don't serve playlist until actual segments exist
		segmentReady := false
		for i := 0; i < 100; i++ {
			time.Sleep(100 * time.Millisecond)
			if info, err := os.Stat(firstSegment); err == nil && info.Size() > 0 {
				segmentReady = true
				log.Printf("[Restream] First segment ready: %s (%d bytes)", filepath.Base(absPath), info.Size())
				break
			}
		}

		if !segmentReady {
			log.Printf("[Restream] Timeout waiting for first segment: %s", filepath.Base(absPath))
			h.writeError(w, http.StatusInternalServerError, "FFmpeg failed to create segments")
			return
		}

		// Cleanup after 2 hours (background goroutine)
		go func() {
			time.Sleep(2 * time.Hour)
			os.RemoveAll(sessionDir)
			log.Printf("[Restream] Cleaned up session: %s", sessionID[:8])
		}()
	}

	// Serve requested file
	var servePath string
	var contentType string

	if segmentName != "" {
		// Serving a segment
		servePath = filepath.Join(sessionDir, segmentName)
		contentType = "video/mp2t"
	} else {
		// Serving playlist
		servePath = playlistPath
		contentType = "application/vnd.apple.mpegurl"
	}

	// Wait for file to exist (max 15 seconds for segments)
	fileReady := false
	maxWait := 50 // 5 seconds for playlist
	if segmentName != "" {
		maxWait = 150 // 15 seconds for segments (FFmpeg might still be processing)
	}

	for i := 0; i < maxWait; i++ {
		if info, err := os.Stat(servePath); err == nil && info.Size() > 0 {
			fileReady = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if !fileReady {
		log.Printf("[Restream] File not ready: %s", servePath)
		h.writeError(w, http.StatusNotFound, "File not ready yet")
		return
	}

	// Read and serve file
	content, err := os.ReadFile(servePath)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to read file")
		return
	}

	// For playlist, rewrite segment URLs to include path parameter
	if segmentName == "" {
		baseURL := fmt.Sprintf("/api/recordings/restream?path=%s&segment=", url.QueryEscape(filePath))
		content = []byte(strings.ReplaceAll(string(content), "seg_", baseURL+"seg_"))
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if segmentName != "" {
		w.Header().Set("Cache-Control", "public, max-age=3600")
	} else {
		w.Header().Set("Cache-Control", "no-cache")
	}
	w.Write(content)
}

// ServeHLSFile serves HLS playlist (.m3u8) and segment (.ts) files
func (h *RecordingHandler) ServeHLSFile(w http.ResponseWriter, r *http.Request) {
	filePath := r.URL.Query().Get("path")
	if filePath == "" {
		h.writeError(w, http.StatusBadRequest, "File path is required")
		return
	}

	// Security: ensure the file is within allowed directories
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid file path")
		return
	}

	// Allow files from recordings path OR HLS cache path
	absRecordingsPath, _ := filepath.Abs(h.recordingsPath)
	hlsCachePath := os.Getenv("HLS_CACHE_PATH")
	if hlsCachePath == "" {
		hlsCachePath = "/app/hls-cache"
	}
	absHLSCachePath, _ := filepath.Abs(hlsCachePath)

	if !strings.HasPrefix(absPath, absRecordingsPath) && !strings.HasPrefix(absPath, absHLSCachePath) {
		h.writeError(w, http.StatusForbidden, "Access denied")
		return
	}

	// Check file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		h.writeError(w, http.StatusNotFound, "HLS file not found")
		return
	}

	// Set correct content type
	ext := strings.ToLower(filepath.Ext(absPath))

	// CORS headers for HLS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")

	if ext == ".m3u8" {
		// For playlist: read, rewrite segment URLs, serve
		content, err := os.ReadFile(absPath)
		if err != nil {
			h.writeError(w, http.StatusInternalServerError, "Failed to read playlist")
			return
		}

		// Rewrite segment URLs: seg_00000.ts -> /api/recordings/hls?path=/full/path/seg_00000.ts
		hlsDir := filepath.Dir(absPath)
		lines := strings.Split(string(content), "\n")
		for i, line := range lines {
			if strings.HasSuffix(line, ".ts") {
				segPath := filepath.Join(hlsDir, strings.TrimSpace(line))
				lines[i] = "/api/recordings/hls?path=" + url.QueryEscape(segPath)
			}
		}
		content = []byte(strings.Join(lines, "\n"))

		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
		w.Header().Set("Cache-Control", "no-cache")
		w.Write(content)
		return
	}

	// For segments (.ts): serve directly with caching
	w.Header().Set("Content-Type", "video/mp2t")
	w.Header().Set("Cache-Control", "public, max-age=86400")

	file, err := os.Open(absPath)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to open file")
		return
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to stat file")
		return
	}

	http.ServeContent(w, r, fileInfo.Name(), fileInfo.ModTime(), file)
}
