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

const maxBitrateBps = 2_000_000 // 2 Mbps threshold

// probeBitrate returns the overall bitrate of a video file in bps using ffprobe
func probeBitrate(videoPath string) (int64, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=bit_rate",
		"-of", "default=noprint_wrappers=1:nokey=1",
		videoPath,
	)
	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	var bitrate int64
	_, err = fmt.Sscanf(strings.TrimSpace(string(out)), "%d", &bitrate)
	if err != nil || bitrate == 0 {
		// fallback: use container-level bitrate
		cmd2 := exec.Command("ffprobe",
			"-v", "error",
			"-show_entries", "format=bit_rate",
			"-of", "default=noprint_wrappers=1:nokey=1",
			videoPath,
		)
		out2, err2 := cmd2.Output()
		if err2 != nil {
			return 0, err2
		}
		fmt.Sscanf(strings.TrimSpace(string(out2)), "%d", &bitrate)
	}
	return bitrate, nil
}

// ConvertToHLS converts a video file to HLS segments.
// If source bitrate <= 2 Mbps: uses -c copy (no re-encode, keeps original size).
// If source bitrate > 2 Mbps: re-encodes down to 2 Mbps.
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

	if err := os.MkdirAll(hlsDir, 0755); err != nil {
		return fmt.Errorf("failed to create HLS directory: %w", err)
	}

	// Probe source bitrate to decide encode strategy
	bitrate, err := probeBitrate(videoPath)
	if err != nil {
		log.Printf("[HLS] Could not probe bitrate for %s: %v — defaulting to copy", filepath.Base(videoPath), err)
	}

	start := time.Now()

	hlsArgs := []string{
		"-hls_time", "4",
		"-hls_list_size", "0",
		"-hls_segment_filename", segmentPattern,
		"-f", "hls",
		"-y",
		playlist,
	}

	var cmd *exec.Cmd
	if bitrate == 0 || bitrate <= maxBitrateBps {
		// Bitrate thấp hoặc không đo được → thử copy trước
		log.Printf("[HLS] %s bitrate=%d bps (<=2Mbps) → trying -c copy", filepath.Base(videoPath), bitrate)
		args := append([]string{
			"-fflags", "+genpts+igndts",
			"-err_detect", "ignore_err",
			"-i", videoPath,
			"-c", "copy",
		}, hlsArgs...)
		cmd = exec.Command("ffmpeg", args...)
	} else {
		// Bitrate cao → re-encode xuống 2 Mbps
		log.Printf("[HLS] %s bitrate=%d bps (>2Mbps) → re-encoding to 2Mbps", filepath.Base(videoPath), bitrate)
		args := append([]string{
			"-fflags", "+genpts+igndts",
			"-err_detect", "ignore_err",
			"-i", videoPath,
			"-c:v", "libx264", "-preset", "fast",
			"-b:v", "2000k", "-maxrate", "2500k", "-bufsize", "4000k",
			"-g", "48", "-keyint_min", "48",
			"-c:a", "aac", "-b:a", "128k",
		}, hlsArgs...)
		cmd = exec.Command("ffmpeg", args...)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Nếu -c copy lỗi (thường do file corrupt) → fallback sang re-encode
		if bitrate == 0 || bitrate <= maxBitrateBps {
			log.Printf("[HLS] -c copy failed for %s (possibly corrupt), retrying with re-encode", filepath.Base(videoPath))
			os.RemoveAll(hlsDir)
			if err2 := os.MkdirAll(hlsDir, 0755); err2 != nil {
				return fmt.Errorf("failed to recreate HLS directory: %w", err2)
			}
			args := append([]string{
				"-fflags", "+genpts+igndts",
				"-err_detect", "ignore_err",
				"-i", videoPath,
				"-c:v", "libx264", "-preset", "fast",
				"-b:v", "2000k", "-maxrate", "2500k", "-bufsize", "4000k",
				"-g", "48", "-keyint_min", "48",
				"-c:a", "aac", "-b:a", "128k",
			}, hlsArgs...)
			cmd2 := exec.Command("ffmpeg", args...)
			output, err = cmd2.CombinedOutput()
		}
		if err != nil {
			os.RemoveAll(hlsDir)
			log.Printf("[HLS] Failed to convert %s: %v\nOutput: %s", filepath.Base(videoPath), err, string(output))
			return fmt.Errorf("ffmpeg failed: %w", err)
		}
		log.Printf("[HLS] Re-encode fallback succeeded for %s", filepath.Base(videoPath))
	}

	elapsed := time.Since(start)
	log.Printf("[HLS] Done: %s in %v", filepath.Base(videoPath), elapsed)
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
