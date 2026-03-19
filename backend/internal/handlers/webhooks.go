package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"camera-dashboard-backend/internal/models"
	"camera-dashboard-backend/internal/services"

	"github.com/jmoiron/sqlx"
)

const (
	// SegmentDuration is the maximum duration for each recording segment
	SegmentDuration = 15 * time.Minute

	// StopDebounceDelay is the delay before stopping recording after stream disconnects
	StopDebounceDelay = 5 * time.Second
)

type WebhookHandler struct {
	db             *sqlx.DB
	recorder       *services.RecorderService
	recordingsPath string

	// Debounce for NotReady events (prevent stopping on brief disconnects)
	pendingStops map[string]*time.Timer

	// Segment timers for auto-splitting recordings every 15 minutes
	segmentTimers map[string]*time.Timer

	// Track active streams (path -> camera info) for segment rotation
	activeStreams map[string]*CameraInfo

	// Track all live streams on MediaMTX (from webhooks)
	// Key: stream path (e.g., "drone_stream1"), Value: true if active
	liveStreams map[string]bool

	mu sync.Mutex
}

func NewWebhookHandler(db *sqlx.DB, recorder *services.RecorderService, recordingsPath string) *WebhookHandler {
	return &WebhookHandler{
		db:             db,
		recorder:       recorder,
		recordingsPath: recordingsPath,
		pendingStops:   make(map[string]*time.Timer),
		segmentTimers:  make(map[string]*time.Timer),
		activeStreams:  make(map[string]*CameraInfo),
		liveStreams:    make(map[string]bool),
	}
}

// StreamWebhookRequest represents the webhook payload from MediaMTX
type StreamWebhookRequest struct {
	Path string `json:"path"`
}

// CameraInfo holds camera and project info for auto-recording
type CameraInfo struct {
	CameraID    int64  `db:"id"`
	CameraName  string `db:"name"`
	ProjectID   int64  `db:"project_id"`
	ProjectName string `db:"project_name"`
	StreamURL   string `db:"stream_url"`
	AutoRecord  bool   `db:"auto_record"` // Auto-record enabled flag
	OutputDir   string // Computed output directory
}

// findCameraByPath finds a camera whose stream_url contains the given path
func (h *WebhookHandler) findCameraByPath(path string) (*CameraInfo, error) {
	// Normalize path (remove leading slash if present)
	path = strings.TrimPrefix(path, "/")

	var camera CameraInfo

	query := `
		SELECT c.id, c.name, c.project_id, p.name as project_name, c.stream_url, c.auto_record
		FROM cameras c
		JOIN projects p ON c.project_id = p.id
		WHERE c.stream_url LIKE ?
		LIMIT 1
	`

	// Try matching with path in stream_url
	err := h.db.Get(&camera, query, "%"+path+"%")
	if err == nil {
		return &camera, nil
	}

	// Strategy 2: Match just the last segment (e.g., "drone_stream1" from "live/drone_stream1")
	segments := strings.Split(path, "/")
	lastSegment := segments[len(segments)-1]

	err = h.db.Get(&camera, query, "%"+lastSegment+"%")
	if err == nil {
		return &camera, nil
	}

	return nil, err
}

// startRecordingWithSegment starts recording and sets up 15-minute segment timer
func (h *WebhookHandler) startRecordingWithSegment(path string, camera *CameraInfo) error {
	// Build output directory
	outputDir := filepath.Join(h.recordingsPath, sanitizeName(camera.ProjectName), sanitizeName(camera.CameraName))
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}
	camera.OutputDir = outputDir

	// Use camera's configured stream URL (could be RTSP, RTMP, or HLS)
	// Start recording with retry logic (HLS may need time to generate initial segments)
	var recording *models.Recording
	var err error
	maxRetries := 3
	retryDelay := 2 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		recording, err = h.recorder.StartRecording(
			camera.CameraID,
			camera.ProjectID,
			camera.ProjectName,
			camera.CameraName,
			camera.StreamURL,
			outputDir,
		)
		if err == nil {
			break
		}

		log.Printf("[Webhook] Recording attempt %d/%d failed for %s: %v", attempt, maxRetries, camera.CameraName, err)

		if attempt < maxRetries {
			log.Printf("[Webhook] Retrying in %v...", retryDelay)
			time.Sleep(retryDelay)
		}
	}
	if err != nil {
		return err
	}

	log.Printf("[Webhook] Auto-recording started: %s -> %s (segment timer: %v)", camera.CameraName, recording.FilePath, SegmentDuration)

	// Store active stream info
	h.mu.Lock()
	h.activeStreams[path] = camera

	// Cancel any existing segment timer
	if timer, exists := h.segmentTimers[path]; exists {
		timer.Stop()
	}

	// Set up 15-minute segment timer
	h.segmentTimers[path] = time.AfterFunc(SegmentDuration, func() {
		h.rotateSegment(path)
	})
	h.mu.Unlock()

	return nil
}

// rotateSegment stops current recording and starts a new one (segment rotation)
func (h *WebhookHandler) rotateSegment(path string) {
	h.mu.Lock()
	camera, exists := h.activeStreams[path]
	if !exists {
		h.mu.Unlock()
		log.Printf("[Webhook] Segment rotation: no active stream for path %s", path)
		return
	}

	// Check if still recording
	isRecording, recordingID := h.recorder.IsRecording(camera.CameraID)
	if !isRecording {
		delete(h.activeStreams, path)
		delete(h.segmentTimers, path)
		h.mu.Unlock()
		log.Printf("[Webhook] Segment rotation: camera %s not recording anymore", camera.CameraName)
		return
	}
	h.mu.Unlock()

	log.Printf("[Webhook] Segment rotation: stopping recording %d for camera %s", recordingID, camera.CameraName)

	// Stop current recording
	if err := h.recorder.StopRecording(recordingID); err != nil {
		log.Printf("[Webhook] Segment rotation: failed to stop recording: %v", err)
		return
	}

	// Wait a moment for FFmpeg to finalize
	time.Sleep(500 * time.Millisecond)

	// Start new recording
	if err := h.startRecordingWithSegment(path, camera); err != nil {
		log.Printf("[Webhook] Segment rotation: failed to start new recording: %v", err)
		// Clear active stream on failure
		h.mu.Lock()
		delete(h.activeStreams, path)
		delete(h.segmentTimers, path)
		h.mu.Unlock()
	} else {
		log.Printf("[Webhook] Segment rotation: new recording started for camera %s", camera.CameraName)
	}
}

// stopRecordingAndCleanup stops recording and cleans up timers
func (h *WebhookHandler) stopRecordingAndCleanup(path string, cameraID int64, recordingID int64) {
	h.mu.Lock()
	// Cancel segment timer
	if timer, exists := h.segmentTimers[path]; exists {
		timer.Stop()
		delete(h.segmentTimers, path)
	}
	// Remove from active streams
	delete(h.activeStreams, path)
	h.mu.Unlock()

	// Stop recording
	if err := h.recorder.StopRecording(recordingID); err != nil {
		log.Printf("[Webhook] Failed to stop recording: %v", err)
	} else {
		log.Printf("[Webhook] Auto-recording stopped for path: %s", path)
	}
}

// HandleStreamReady handles MediaMTX runOnReady webhook
// Called when a stream starts publishing
func (h *WebhookHandler) HandleStreamReady(w http.ResponseWriter, r *http.Request) {
	var req StreamWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[Webhook] Invalid request body: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	log.Printf("[Webhook] Stream READY: %s", req.Path)

	// Track this stream as live
	h.mu.Lock()
	h.liveStreams[req.Path] = true

	// Cancel any pending stop for this path (stream reconnected quickly)
	if timer, exists := h.pendingStops[req.Path]; exists {
		timer.Stop()
		delete(h.pendingStops, req.Path)
		log.Printf("[Webhook] Cancelled pending stop for: %s (stream reconnected)", req.Path)
	}
	h.mu.Unlock()

	// Find camera by path
	camera, err := h.findCameraByPath(req.Path)
	if err != nil {
		log.Printf("[Webhook] No camera found for path: %s", req.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ignored","reason":"no matching camera"}`))
		return
	}

	log.Printf("[Webhook] Found camera: %s (ID: %d, Project: %s, AutoRecord: %v)",
		camera.CameraName, camera.CameraID, camera.ProjectName, camera.AutoRecord)

	// Check if auto-record is enabled for this camera
	if !camera.AutoRecord {
		log.Printf("[Webhook] Auto-record DISABLED for camera: %s (skipping)", camera.CameraName)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ignored","reason":"auto-record disabled"}`))
		return
	}

	// Check if already recording this camera
	if isRecording, recordingID := h.recorder.IsRecording(camera.CameraID); isRecording {
		log.Printf("[Webhook] Camera %s already recording (ID: %d)", camera.CameraName, recordingID)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":      "already_recording",
			"recordingId": recordingID,
		})
		return
	}

	// Start recording with segment timer
	if err := h.startRecordingWithSegment(req.Path, camera); err != nil {
		log.Printf("[Webhook] Failed to start recording: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":          "recording_started",
		"cameraId":        camera.CameraID,
		"cameraName":      camera.CameraName,
		"segmentDuration": SegmentDuration.String(),
	})
}

// HandleStreamNotReady handles MediaMTX runOnNotReady webhook
// Called when a stream stops publishing
func (h *WebhookHandler) HandleStreamNotReady(w http.ResponseWriter, r *http.Request) {
	var req StreamWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[Webhook] Invalid request body: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	log.Printf("[Webhook] Stream NOT READY: %s", req.Path)

	// Remove from live streams tracking
	h.mu.Lock()
	delete(h.liveStreams, req.Path)
	h.mu.Unlock()

	// Find camera by path
	camera, err := h.findCameraByPath(req.Path)
	if err != nil {
		log.Printf("[Webhook] No camera found for path: %s", req.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ignored","reason":"no matching camera"}`))
		return
	}

	// Check if this camera is currently recording
	isRecording, recordingID := h.recorder.IsRecording(camera.CameraID)
	if !isRecording {
		log.Printf("[Webhook] Camera %s not recording, ignoring", camera.CameraName)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ignored","reason":"not recording"}`))
		return
	}

	// Debounce: wait before actually stopping (in case of brief disconnect)
	h.mu.Lock()

	// Cancel existing pending stop if any
	if timer, exists := h.pendingStops[req.Path]; exists {
		timer.Stop()
	}

	log.Printf("[Webhook] Scheduling stop for camera %s in %v (recording ID: %d)", camera.CameraName, StopDebounceDelay, recordingID)

	// Capture values for closure
	path := req.Path
	camID := camera.CameraID
	recID := recordingID

	h.pendingStops[req.Path] = time.AfterFunc(StopDebounceDelay, func() {
		h.mu.Lock()
		delete(h.pendingStops, path)
		h.mu.Unlock()

		// Re-check if still recording the same recording
		if stillRecording, currentID := h.recorder.IsRecording(camID); stillRecording && currentID == recID {
			log.Printf("[Webhook] Executing delayed stop for camera (recording ID: %d)", recID)
			h.stopRecordingAndCleanup(path, camID, recID)
		} else {
			log.Printf("[Webhook] Recording already stopped or changed, skipping stop")
		}
	})

	h.mu.Unlock()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":      "stop_scheduled",
		"recordingId": recordingID,
		"cameraId":    camera.CameraID,
		"cameraName":  camera.CameraName,
		"delay":       StopDebounceDelay.String(),
	})
}

// StartPeriodicCheck starts a goroutine that periodically checks all cameras
// with auto_record=1 and active streams, ensuring they are recording
func (h *WebhookHandler) StartPeriodicCheck(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		// Run once immediately on startup
		h.checkAllCamerasAutoRecord()

		for range ticker.C {
			h.checkAllCamerasAutoRecord()
		}
	}()
	log.Printf("[Webhook] Started periodic auto-record check (interval: %v)", interval)
}

// checkAllCamerasAutoRecord checks all cameras with auto_record=1 and starts recording if stream is active
func (h *WebhookHandler) checkAllCamerasAutoRecord() {
	var cameras []CameraInfo
	err := h.db.Select(&cameras, `
		SELECT c.id, c.name, c.project_id, p.name as project_name, c.stream_url, c.auto_record
		FROM cameras c
		JOIN projects p ON c.project_id = p.id
		WHERE c.auto_record = 1
	`)
	if err != nil {
		log.Printf("[Webhook] Failed to query cameras: %v", err)
		return
	}

	for _, camera := range cameras {
		// Skip if already recording
		if isRecording, _ := h.recorder.IsRecording(camera.CameraID); isRecording {
			continue
		}

		// Check if stream is active
		streamPath := extractStreamPath(camera.StreamURL)
		if h.isStreamActive(streamPath) {
			log.Printf("[Webhook] Auto-record check: stream active, starting recording for %s", camera.CameraName)
			cameraCopy := camera
			h.startRecordingWithSegment(streamPath, &cameraCopy)
		}
	}
}

// TriggerAutoRecordForCamera checks if camera's stream is active and starts recording
// Called when auto_record is enabled on a camera
func (h *WebhookHandler) TriggerAutoRecordForCamera(cameraID int64) {
	log.Printf("[Webhook] Triggering auto-record check for camera ID: %d", cameraID)

	// Get camera info
	var camera CameraInfo
	err := h.db.Get(&camera, `
		SELECT c.id, c.name, c.project_id, p.name as project_name, c.stream_url, c.auto_record
		FROM cameras c
		JOIN projects p ON c.project_id = p.id
		WHERE c.id = ?
	`, cameraID)
	if err != nil {
		log.Printf("[Webhook] Camera not found: %d", cameraID)
		return
	}

	// Condition 1: auto_record must be enabled
	if !camera.AutoRecord {
		log.Printf("[Webhook] Auto-record not enabled for camera: %s", camera.CameraName)
		return
	}

	// Check if already recording
	if isRecording, recordingID := h.recorder.IsRecording(camera.CameraID); isRecording {
		log.Printf("[Webhook] Camera %s already recording (ID: %d)", camera.CameraName, recordingID)
		return
	}

	// Extract stream path from URL
	streamPath := extractStreamPath(camera.StreamURL)
	log.Printf("[Webhook] Camera %s: checking stream %s", camera.CameraName, streamPath)

	// Condition 2: stream must be active (check HLS like frontend)
	if !h.isStreamActive(streamPath) {
		log.Printf("[Webhook] Stream %s not active for camera %s", streamPath, camera.CameraName)
		return
	}

	// Both conditions met: auto_record=1 AND stream active -> start recording
	log.Printf("[Webhook] Starting auto-record for camera %s (stream: %s)", camera.CameraName, streamPath)
	if err := h.startRecordingWithSegment(streamPath, &camera); err != nil {
		log.Printf("[Webhook] Failed to start recording for %s: %v", camera.CameraName, err)
	} else {
		log.Printf("[Webhook] Auto-recording started for camera %s", camera.CameraName)
	}
}

// StopRecordingForCamera stops recording for a camera when auto_record is disabled
func (h *WebhookHandler) StopRecordingForCamera(cameraID int64) {
	log.Printf("[Webhook] Stopping recording for camera ID: %d (auto_record disabled)", cameraID)

	isRecording, recordingID := h.recorder.IsRecording(cameraID)
	if !isRecording {
		log.Printf("[Webhook] Camera %d not recording, nothing to stop", cameraID)
		return
	}

	if err := h.recorder.StopRecording(recordingID); err != nil {
		log.Printf("[Webhook] Failed to stop recording %d: %v", recordingID, err)
	} else {
		log.Printf("[Webhook] Recording %d stopped for camera %d", recordingID, cameraID)
	}
}

// extractStreamPath extracts stream name from URL
// e.g., "/hls/drone_stream5/index.m3u8" -> "drone_stream5"
// e.g., "/hls/grid/stream_1/index.m3u8" -> "stream_1"
func extractStreamPath(url string) string {
	// Remove more specific prefixes FIRST, then general ones
	url = strings.TrimPrefix(url, "/hls/grid/")
	url = strings.TrimPrefix(url, "/hls/full/")
	url = strings.TrimPrefix(url, "/hls/record/")
	url = strings.TrimPrefix(url, "/hls/") // Most general - last

	// Remove suffix like /index.m3u8
	if idx := strings.Index(url, "/"); idx > 0 {
		url = url[:idx]
	}

	return url
}

// isStreamActive checks if stream is active by requesting HLS endpoint (same as frontend)
func (h *WebhookHandler) isStreamActive(streamPath string) bool {
	client := &http.Client{Timeout: 5 * time.Second}

	// Check HLS endpoint directly - same as frontend does
	hlsURL := "http://camera-mediamtx:8888/" + streamPath + "/index.m3u8"
	log.Printf("[Webhook] Checking HLS: %s", hlsURL)

	resp, err := client.Get(hlsURL)
	if err != nil {
		log.Printf("[Webhook] HLS check failed: %v", err)
		return false
	}
	defer resp.Body.Close()

	log.Printf("[Webhook] HLS %s: status=%d", streamPath, resp.StatusCode)
	return resp.StatusCode == http.StatusOK
}

// Note: sanitizeName is defined in recordings.go (same package)
