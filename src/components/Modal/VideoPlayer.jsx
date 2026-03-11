import { useState, useRef, useEffect } from 'react'
import Hls from 'hls.js'
import { recordingsApi } from '../../api/recordings'
import './VideoPlayer.css'

function VideoPlayer({ recording, onClose }) {
    const videoRef = useRef(null)
    const hlsRef = useRef(null)
    const [isPlaying, setIsPlaying] = useState(false)
    const [currentTime, setCurrentTime] = useState(0)
    const [duration, setDuration] = useState(0)
    const [volume, setVolume] = useState(1)
    const [isMuted, setIsMuted] = useState(false)
    const [playbackRate, setPlaybackRate] = useState(1)
    const [isFullscreen, setIsFullscreen] = useState(false)
    const [showControls, setShowControls] = useState(true)
    const containerRef = useRef(null)
    const controlsTimeoutRef = useRef(null)

    // Use HLS if available, otherwise fallback to direct MP4 stream
    const hlsUrl = recording.hlsUrl || null
    const streamUrl = recordingsApi.getFileStreamUrl(recording.path)
    const downloadUrl = recordingsApi.getFileDownloadUrl(recording.path)

    // Setup HLS.js or direct video source
    useEffect(() => {
        const video = videoRef.current
        if (!video) return

        if (hlsUrl && Hls.isSupported()) {
            // Use HLS.js for instant playback (loads small segments)
            const hls = new Hls({
                maxBufferLength: 15,
                maxMaxBufferLength: 30,
                maxBufferSize: 30 * 1000 * 1000,
            })
            hls.loadSource(hlsUrl)
            hls.attachMedia(video)
            hls.on(Hls.Events.MANIFEST_PARSED, () => {
                console.log('[VideoPlayer] HLS ready, starting playback')
                video.play().catch(() => {})
            })
            hls.on(Hls.Events.ERROR, (_, data) => {
                if (data.fatal) {
                    console.error('[VideoPlayer] HLS fatal error, falling back to MP4', data)
                    hls.destroy()
                    video.src = streamUrl
                }
            })
            hlsRef.current = hls
        } else {
            // Fallback: direct MP4 stream
            video.src = streamUrl
        }

        return () => {
            if (hlsRef.current) {
                hlsRef.current.destroy()
                hlsRef.current = null
            }
        }
    }, [hlsUrl, streamUrl])

    // Format time (seconds to mm:ss or hh:mm:ss)
    const formatTime = (seconds) => {
        if (isNaN(seconds)) return '00:00'
        const h = Math.floor(seconds / 3600)
        const m = Math.floor((seconds % 3600) / 60)
        const s = Math.floor(seconds % 60)
        if (h > 0) {
            return `${h}:${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`
        }
        return `${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`
    }

    // Format file size
    const formatFileSize = (bytes) => {
        if (!bytes) return '0 B'
        const k = 1024
        const sizes = ['B', 'KB', 'MB', 'GB']
        const i = Math.floor(Math.log(bytes) / Math.log(k))
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
    }

    // Format date
    const formatDate = (dateString) => {
        if (!dateString) return ''
        return new Date(dateString).toLocaleString('vi-VN', {
            day: '2-digit',
            month: '2-digit',
            year: 'numeric',
            hour: '2-digit',
            minute: '2-digit'
        })
    }

    // Video event handlers
    const handleLoadedMetadata = () => {
        if (videoRef.current) {
            setDuration(videoRef.current.duration)
        }
    }

    const handleTimeUpdate = () => {
        if (videoRef.current) {
            setCurrentTime(videoRef.current.currentTime)
        }
    }

    const handlePlayPause = () => {
        if (videoRef.current) {
            if (isPlaying) {
                videoRef.current.pause()
            } else {
                videoRef.current.play()
            }
            setIsPlaying(!isPlaying)
        }
    }

    const handleSeek = (e) => {
        const rect = e.currentTarget.getBoundingClientRect()
        const pos = (e.clientX - rect.left) / rect.width
        if (videoRef.current) {
            videoRef.current.currentTime = pos * duration
        }
    }

    const handleVolumeChange = (e) => {
        const newVolume = parseFloat(e.target.value)
        setVolume(newVolume)
        if (videoRef.current) {
            videoRef.current.volume = newVolume
        }
        setIsMuted(newVolume === 0)
    }

    const handleMuteToggle = () => {
        if (videoRef.current) {
            if (isMuted) {
                videoRef.current.volume = volume || 1
                setIsMuted(false)
            } else {
                videoRef.current.volume = 0
                setIsMuted(true)
            }
        }
    }

    const handlePlaybackRateChange = (rate) => {
        setPlaybackRate(rate)
        if (videoRef.current) {
            videoRef.current.playbackRate = rate
        }
    }

    const handleFullscreen = () => {
        if (!document.fullscreenElement) {
            containerRef.current?.requestFullscreen()
            setIsFullscreen(true)
        } else {
            document.exitFullscreen()
            setIsFullscreen(false)
        }
    }

    const handleDownload = () => {
        window.open(downloadUrl, '_blank')
    }

    // Auto-hide controls
    const handleMouseMove = () => {
        setShowControls(true)
        if (controlsTimeoutRef.current) {
            clearTimeout(controlsTimeoutRef.current)
        }
        if (isPlaying) {
            controlsTimeoutRef.current = setTimeout(() => {
                setShowControls(false)
            }, 3000)
        }
    }

    // Keyboard shortcuts
    useEffect(() => {
        const handleKeyDown = (e) => {
            switch (e.key) {
                case ' ':
                case 'k':
                    e.preventDefault()
                    handlePlayPause()
                    break
                case 'ArrowLeft':
                    e.preventDefault()
                    if (videoRef.current) {
                        videoRef.current.currentTime -= 5
                    }
                    break
                case 'ArrowRight':
                    e.preventDefault()
                    if (videoRef.current) {
                        videoRef.current.currentTime += 5
                    }
                    break
                case 'ArrowUp':
                    e.preventDefault()
                    setVolume(v => Math.min(1, v + 0.1))
                    break
                case 'ArrowDown':
                    e.preventDefault()
                    setVolume(v => Math.max(0, v - 0.1))
                    break
                case 'f':
                    e.preventDefault()
                    handleFullscreen()
                    break
                case 'm':
                    e.preventDefault()
                    handleMuteToggle()
                    break
                case 'Escape':
                    if (isFullscreen) {
                        document.exitFullscreen()
                    } else {
                        onClose()
                    }
                    break
            }
        }

        document.addEventListener('keydown', handleKeyDown)
        return () => document.removeEventListener('keydown', handleKeyDown)
    }, [isPlaying, isFullscreen, onClose])

    // Update volume when state changes
    useEffect(() => {
        if (videoRef.current) {
            videoRef.current.volume = volume
        }
    }, [volume])

    // Fullscreen change listener
    useEffect(() => {
        const handleFullscreenChange = () => {
            setIsFullscreen(!!document.fullscreenElement)
        }
        document.addEventListener('fullscreenchange', handleFullscreenChange)
        return () => document.removeEventListener('fullscreenchange', handleFullscreenChange)
    }, [])

    return (
        <div className="video-player-overlay" onClick={onClose}>
            <div
                ref={containerRef}
                className={`video-player-container ${isFullscreen ? 'fullscreen' : ''}`}
                onClick={(e) => e.stopPropagation()}
                onMouseMove={handleMouseMove}
            >
                {/* Header */}
                <div className={`video-player-header ${showControls ? 'visible' : ''}`}>
                    <div className="video-info">
                        <h3 className="video-title">{recording.name}</h3>
                        <div className="video-meta">
                            <span>{formatDate(recording.modTime)}</span>
                            <span className="separator">|</span>
                            <span>{formatFileSize(recording.size)}</span>
                        </div>
                    </div>
                    <button className="video-close-btn" onClick={onClose}>
                        <svg viewBox="0 0 24 24" fill="currentColor">
                            <path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z" />
                        </svg>
                    </button>
                </div>

                {/* Video */}
                <video
                    ref={videoRef}
                    className="video-element"
                    onLoadedMetadata={handleLoadedMetadata}
                    onTimeUpdate={handleTimeUpdate}
                    onPlay={() => setIsPlaying(true)}
                    onPause={() => setIsPlaying(false)}
                    onEnded={() => setIsPlaying(false)}
                    onClick={handlePlayPause}
                />

                {/* Play overlay */}
                {!isPlaying && (
                    <div className="play-overlay" onClick={handlePlayPause}>
                        <svg viewBox="0 0 24 24" fill="currentColor">
                            <path d="M8 5v14l11-7z" />
                        </svg>
                    </div>
                )}

                {/* Controls */}
                <div className={`video-controls ${showControls ? 'visible' : ''}`}>
                    {/* Progress bar */}
                    <div className="progress-container" onClick={handleSeek}>
                        <div className="progress-bar">
                            <div
                                className="progress-filled"
                                style={{ width: `${(currentTime / duration) * 100}%` }}
                            />
                        </div>
                    </div>

                    <div className="controls-row">
                        {/* Left controls */}
                        <div className="controls-left">
                            <button className="control-btn" onClick={handlePlayPause}>
                                {isPlaying ? (
                                    <svg viewBox="0 0 24 24" fill="currentColor">
                                        <path d="M6 19h4V5H6v14zm8-14v14h4V5h-4z" />
                                    </svg>
                                ) : (
                                    <svg viewBox="0 0 24 24" fill="currentColor">
                                        <path d="M8 5v14l11-7z" />
                                    </svg>
                                )}
                            </button>

                            <button className="control-btn" onClick={handleMuteToggle}>
                                {isMuted ? (
                                    <svg viewBox="0 0 24 24" fill="currentColor">
                                        <path d="M16.5 12c0-1.77-1.02-3.29-2.5-4.03v2.21l2.45 2.45c.03-.2.05-.41.05-.63zm2.5 0c0 .94-.2 1.82-.54 2.64l1.51 1.51C20.63 14.91 21 13.5 21 12c0-4.28-2.99-7.86-7-8.77v2.06c2.89.86 5 3.54 5 6.71zM4.27 3L3 4.27 7.73 9H3v6h4l5 5v-6.73l4.25 4.25c-.67.52-1.42.93-2.25 1.18v2.06c1.38-.31 2.63-.95 3.69-1.81L19.73 21 21 19.73l-9-9L4.27 3zM12 4L9.91 6.09 12 8.18V4z" />
                                    </svg>
                                ) : (
                                    <svg viewBox="0 0 24 24" fill="currentColor">
                                        <path d="M3 9v6h4l5 5V4L7 9H3zm13.5 3c0-1.77-1.02-3.29-2.5-4.03v8.05c1.48-.73 2.5-2.25 2.5-4.02zM14 3.23v2.06c2.89.86 5 3.54 5 6.71s-2.11 5.85-5 6.71v2.06c4.01-.91 7-4.49 7-8.77s-2.99-7.86-7-8.77z" />
                                    </svg>
                                )}
                            </button>

                            <input
                                type="range"
                                className="volume-slider"
                                min="0"
                                max="1"
                                step="0.1"
                                value={isMuted ? 0 : volume}
                                onChange={handleVolumeChange}
                            />

                            <span className="time-display">
                                {formatTime(currentTime)} / {formatTime(duration)}
                            </span>
                        </div>

                        {/* Right controls */}
                        <div className="controls-right">
                            {/* Playback speed */}
                            <div className="speed-control">
                                <select
                                    value={playbackRate}
                                    onChange={(e) => handlePlaybackRateChange(parseFloat(e.target.value))}
                                    className="speed-select"
                                >
                                    <option value="0.5">0.5x</option>
                                    <option value="1">1x</option>
                                    <option value="1.5">1.5x</option>
                                    <option value="2">2x</option>
                                </select>
                            </div>

                            <button className="control-btn" onClick={handleDownload} title="Tai xuong">
                                <svg viewBox="0 0 24 24" fill="currentColor">
                                    <path d="M19 9h-4V3H9v6H5l7 7 7-7zM5 18v2h14v-2H5z" />
                                </svg>
                            </button>

                            <button className="control-btn" onClick={handleFullscreen} title="Toan man hinh">
                                {isFullscreen ? (
                                    <svg viewBox="0 0 24 24" fill="currentColor">
                                        <path d="M5 16h3v3h2v-5H5v2zm3-8H5v2h5V5H8v3zm6 11h2v-3h3v-2h-5v5zm2-11V5h-2v5h5V8h-3z" />
                                    </svg>
                                ) : (
                                    <svg viewBox="0 0 24 24" fill="currentColor">
                                        <path d="M7 14H5v5h5v-2H7v-3zm-2-4h2V7h3V5H5v5zm12 7h-3v2h5v-5h-2v3zM14 5v2h3v3h2V5h-5z" />
                                    </svg>
                                )}
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    )
}

export default VideoPlayer
