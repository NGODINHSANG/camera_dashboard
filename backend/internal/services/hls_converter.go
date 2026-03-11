package services

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
)

type HLSConverter struct {
	recordingsPath string
	converting     map[string]bool // track files being converted
	mu             sync.Mutex
	db             *sqlx.DB // DB to check active recordings
}

func NewHLSConverter(recordingsPath string, db *sqlx.DB) *HLSConverter {
	return &HLSConverter{
		recordingsPath: recordingsPath,
		converting:     make(map[string]bool),
		db:             db,
	}
}

// getActiveRecordingPaths returns file paths of recordings currently in progress
func (h *HLSConverter) getActiveRecordingPaths() map[string]bool {
	activeFiles := make(map[string]bool)
	if h.db == nil {
		return activeFiles
	}

	var paths []string
	err := h.db.Select(&paths, `SELECT file_path FROM recordings WHERE status = 'recording'`)
	if err != nil {
		log.Printf("[HLS] Failed to query active recordings: %v", err)
		return activeFiles
	}

	for _, p := range paths {
		activeFiles[p] = true
	}
	return activeFiles
}

// Global HLS cache settings
var (
	hlsCachePath   string // Local disk path for HLS cache
	hlsRecBasePath string // Recordings base path (for relative path calculation)
)

// InitHLSCache configures HLS to use local cache instead of storing next to video files
func InitHLSCache(cachePath, recordingsPath string) {
	if cachePath == "" {
		return
	}
	hlsCachePath = cachePath
	hlsRecBasePath = recordingsPath
	os.MkdirAll(cachePath, 0755)
	log.Printf("[HLS] Cache enabled: %s (recordings: %s)", cachePath, recordingsPath)
}

// GetHLSDir returns the HLS directory path for a video file
// Uses local cache if configured, otherwise stores next to video file
func GetHLSDir(videoPath string) string {
	ext := filepath.Ext(videoPath)
	base := strings.TrimSuffix(videoPath, ext)

	if hlsCachePath != "" && hlsRecBasePath != "" {
		// Map SMB path to local cache: /app/recordings/X/video.mp4 → /app/hls-cache/X/video_hls
		relPath, err := filepath.Rel(hlsRecBasePath, base)
		if err == nil {
			return filepath.Join(hlsCachePath, relPath+"_hls")
		}
	}

	return base + "_hls"
}

// GetHLSPlaylist returns the playlist path for a video file
func GetHLSPlaylist(videoPath string) string {
	return filepath.Join(GetHLSDir(videoPath), "playlist.m3u8")
}

// HasHLS checks if HLS version exists for a video
func HasHLS(videoPath string) bool {
	playlist := GetHLSPlaylist(videoPath)
	_, err := os.Stat(playlist)
	return err == nil
}

// ConvertToHLS converts a video file to HLS segments using -c copy
// Keeps ORIGINAL QUALITY, bandwidth saved by only downloading segments being watched
func (h *HLSConverter) ConvertToHLS(videoPath string) error {
	h.mu.Lock()
	if h.converting[videoPath] {
		h.mu.Unlock()
		return fmt.Errorf("already converting: %s", videoPath)
	}
	h.converting[videoPath] = true
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		delete(h.converting, videoPath)
		h.mu.Unlock()
	}()

	hlsDir := GetHLSDir(videoPath)
	playlist := filepath.Join(hlsDir, "playlist.m3u8")
	segmentPattern := filepath.Join(hlsDir, "seg_%05d.ts")

	// Create HLS directory
	if err := os.MkdirAll(hlsDir, 0755); err != nil {
		return fmt.Errorf("failed to create HLS directory: %w", err)
	}

	log.Printf("[HLS] Converting: %s (re-encode with forced keyframes for fast playback)", filepath.Base(videoPath))
	start := time.Now()

	// Re-encode with forced keyframes every 2 seconds
	// This ensures each HLS segment is small (~1-2MB) for instant playback
	// -c copy created 30-40MB segments (unusable over internet)
	cmd := exec.Command("ffmpeg",
		"-i", videoPath,
		"-c:v", "libx264",
		"-preset", "fast",
		"-b:v", "2000k", // 2 Mbps video bitrate
		"-maxrate", "2500k",
		"-bufsize", "4000k",
		"-g", "48", // Keyframe every 48 frames (2 seconds at 24fps)
		"-keyint_min", "48",
		"-c:a", "aac",
		"-b:a", "128k",
		"-hls_time", "4", // 4 second segments (~1MB each)
		"-hls_list_size", "0", // Keep all segments in playlist
		"-hls_segment_filename", segmentPattern,
		"-f", "hls",
		"-y",
		playlist,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Cleanup on failure
		os.RemoveAll(hlsDir)
		log.Printf("[HLS] Failed to convert %s: %v\nOutput: %s", filepath.Base(videoPath), err, string(output))
		return fmt.Errorf("ffmpeg failed: %w", err)
	}

	elapsed := time.Since(start)
	log.Printf("[HLS] Converted: %s in %v", filepath.Base(videoPath), elapsed)
	return nil
}

// StartWorker runs a background worker that scans for videos without HLS versions
func (h *HLSConverter) StartWorker(interval time.Duration) {
	go func() {
		// Initial scan after 10 seconds
		time.Sleep(10 * time.Second)
		h.scanAndConvert()

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			h.scanAndConvert()
		}
	}()
	log.Printf("[HLS] Worker started, scanning every %v", interval)
}

func (h *HLSConverter) scanAndConvert() {
	videoExtensions := map[string]bool{
		".mp4": true, ".mkv": true, ".avi": true,
		".mov": true, ".webm": true, ".ts": true,
		".flv": true, ".wmv": true,
	}

	// Get list of files currently being recorded - SKIP these!
	activeRecordings := h.getActiveRecordingPaths()
	if len(activeRecordings) > 0 {
		log.Printf("[HLS] Skipping %d files currently being recorded", len(activeRecordings))
	}

	// Collect files to convert first
	var toConvert []string

	filepath.Walk(h.recordingsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			if strings.HasSuffix(path, "_hls") {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if !videoExtensions[ext] {
			return nil
		}

		if info.Size() < 10*1024*1024 {
			return nil
		}

		if HasHLS(path) {
			return nil
		}

		// Skip files currently being recorded
		if activeRecordings[path] {
			log.Printf("[HLS] Skipping active recording: %s", filepath.Base(path))
			return nil
		}

		// Extra safety: skip files modified very recently (30 seconds buffer)
		if time.Since(info.ModTime()) < 30*time.Second {
			return nil
		}

		toConvert = append(toConvert, path)
		return nil
	})

	if len(toConvert) == 0 {
		return
	}

	log.Printf("[HLS] Found %d files to convert, processing with 4 workers", len(toConvert))

	// Convert with 4 parallel workers
	const maxWorkers = 4
	sem := make(chan struct{}, maxWorkers)
	var wg sync.WaitGroup

	for _, path := range toConvert {
		wg.Add(1)
		sem <- struct{}{} // acquire

		go func(p string) {
			defer wg.Done()
			defer func() { <-sem }() // release

			log.Printf("[HLS] Converting: %s", filepath.Base(p))
			if err := h.ConvertToHLS(p); err != nil {
				log.Printf("[HLS] Error converting %s: %v", filepath.Base(p), err)
			}
		}(path)
	}

	wg.Wait()
}
