package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
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
	db       *sqlx.DB
	recorder *services.RecorderService
}

func NewRecordingHandler(db *sqlx.DB, recorder *services.RecorderService) *RecordingHandler {
	return &RecordingHandler{
		db:       db,
		recorder: recorder,
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

	if req.OutputDir == "" {
		h.writeError(w, http.StatusBadRequest, "Output directory is required")
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

	// Start recording
	recording, err := h.recorder.StartRecording(
		camera.ID,
		camera.ProjectID,
		projectName,
		camera.Name,
		camera.StreamURL,
		req.OutputDir,
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
