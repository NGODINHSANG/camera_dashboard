import { useState, useEffect } from 'react'
import { recordingsApi } from '../../api/recordings'
import './RecordingsPanel.css'

function RecordingsPanel({ camera, onClose, onPlayRecording, onPlayPlaylist }) {
    const [recordings, setRecordings] = useState([])
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState(null)

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

    // Fetch recordings from file system
    useEffect(() => {
        const fetchRecordings = async () => {
            setLoading(true)
            setError(null)
            try {
                // Use file-based listing with project name and camera name
                const projectName = camera.projectName || 'Unknown'
                const cameraName = camera.name
                const files = await recordingsApi.listFiles(projectName, cameraName)
                setRecordings(Array.isArray(files) ? files : [])
            } catch (err) {
                console.error('Error fetching recordings:', err)
                setError('Khong the tai danh sach ban ghi')
            } finally {
                setLoading(false)
            }
        }

        if (camera?.name) {
            fetchRecordings()
        }
    }, [camera?.name, camera?.projectName])

    const handlePlay = (recording, index) => {
        console.log('handlePlay called:', recording, 'index:', index)
        // Create playback object with URLs
        // Priority: hlsUrl (pre-converted) > restreamUrl (on-demand HLS) > streamUrl (direct)
        const playbackRecording = {
            ...recording,
            streamUrl: recordingsApi.getFileStreamUrl(recording.path),
            restreamUrl: recordingsApi.getRestreamUrl(recording.path),  // On-demand HLS
            hlsUrl: recording.hlsUrl || null,  // Pre-converted HLS if available
        }

        // Create full playlist with all recordings
        const playlist = recordings.map(rec => ({
            ...rec,
            streamUrl: recordingsApi.getFileStreamUrl(rec.path),
            restreamUrl: recordingsApi.getRestreamUrl(rec.path),
            hlsUrl: rec.hlsUrl || null,
        }))

        console.log('Calling onPlayPlaylist with playlist:', playlist, 'startIndex:', index)
        if (onPlayPlaylist) {
            // Pass full playlist and starting index
            onPlayPlaylist(playlist, index)
        } else if (onPlayRecording) {
            // Fallback to single recording
            onPlayRecording(playbackRecording)
        } else {
            console.error('onPlayRecording/onPlayPlaylist is not defined!')
            alert('Lỗi: Callback chưa được định nghĩa')
        }
    }

    const handleDownload = (recording) => {
        window.open(recordingsApi.getFileDownloadUrl(recording.path), '_blank')
    }

    const handleDelete = async (recording) => {
        if (!window.confirm(`Ban co chac muon xoa ban ghi "${recording.name}"?`)) {
            return
        }

        try {
            await recordingsApi.deleteFile(recording.path)
            setRecordings(prev => prev.filter(r => r.path !== recording.path))
        } catch (err) {
            console.error('Error deleting recording:', err)
            alert('Khong the xoa ban ghi')
        }
    }

    return (
        <>
            <div className="recordings-panel-overlay" onClick={onClose}>
                <div className="recordings-panel" onClick={(e) => e.stopPropagation()}>
                    {/* Header */}
                    <div className="recordings-header">
                        <div className="recordings-title-wrapper">
                            <svg className="recordings-icon" viewBox="0 0 24 24" fill="currentColor">
                                <path d="M4 6H2v14c0 1.1.9 2 2 2h14v-2H4V6zm16-4H8c-1.1 0-2 .9-2 2v12c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2zm-8 12.5v-9l6 4.5-6 4.5z" />
                            </svg>
                            <div>
                                <h2 className="recordings-title">Ban ghi - {camera.name}</h2>
                                <p className="recordings-subtitle">{camera.projectName || 'Du an'}</p>
                            </div>
                        </div>
                        <button className="recordings-close" onClick={onClose}>
                            <svg viewBox="0 0 24 24" fill="currentColor">
                                <path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z" />
                            </svg>
                        </button>
                    </div>

                    {/* Content */}
                    <div className="recordings-content">
                        {loading && (
                            <div className="recordings-loading">
                                <div className="loading-spinner"></div>
                                <p>Dang tai...</p>
                            </div>
                        )}

                        {error && (
                            <div className="recordings-error">
                                <svg viewBox="0 0 24 24" fill="currentColor">
                                    <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z" />
                                </svg>
                                <p>{error}</p>
                            </div>
                        )}

                        {!loading && !error && recordings.length === 0 && (
                            <div className="recordings-empty">
                                <svg viewBox="0 0 24 24" fill="currentColor">
                                    <path d="M18 4l2 4h-3l-2-4h-2l2 4h-3l-2-4H8l2 4H7L5 4H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V4h-4z" />
                                </svg>
                                <p>Chua co ban ghi nao</p>
                                <span>Bat dau ghi hinh de xem lai sau</span>
                            </div>
                        )}

                        {!loading && !error && recordings.length > 0 && (
                            <div className="recordings-list">
                                {recordings.map((recording, index) => (
                                    <div
                                        key={recording.path || index}
                                        className="recording-item"
                                        onClick={() => handlePlay(recording, index)}
                                        style={{ cursor: 'pointer' }}
                                    >
                                        <div className="recording-thumbnail">
                                            <svg viewBox="0 0 24 24" fill="currentColor">
                                                <path d="M8 5v14l11-7z" />
                                            </svg>
                                        </div>
                                        <div className="recording-info">
                                            <h4 className="recording-name">{recording.name}</h4>
                                            <div className="recording-meta">
                                                <span className="recording-date">
                                                    <svg viewBox="0 0 24 24" fill="currentColor">
                                                        <path d="M11.99 2C6.47 2 2 6.48 2 12s4.47 10 9.99 10C17.52 22 22 17.52 22 12S17.52 2 11.99 2zM12 20c-4.42 0-8-3.58-8-8s3.58-8 8-8 8 3.58 8 8-3.58 8-8 8zm.5-13H11v6l5.25 3.15.75-1.23-4.5-2.67z" />
                                                    </svg>
                                                    {formatDate(recording.modTime)}
                                                </span>
                                                <span className="recording-size">
                                                    <svg viewBox="0 0 24 24" fill="currentColor">
                                                        <path d="M2 20h20v-4H2v4zm2-3h2v2H4v-2zM2 4v4h20V4H2zm4 3H4V5h2v2zm-4 7h20v-4H2v4zm2-3h2v2H4v-2z" />
                                                    </svg>
                                                    {formatFileSize(recording.size)}
                                                </span>
                                            </div>
                                        </div>
                                        <div className="recording-actions" onClick={(e) => e.stopPropagation()}>
                                            <button
                                                className="action-btn play-btn"
                                                onClick={() => handlePlay(recording, index)}
                                                title="Phat"
                                            >
                                                <svg viewBox="0 0 24 24" fill="currentColor">
                                                    <path d="M8 5v14l11-7z" />
                                                </svg>
                                            </button>
                                            <button
                                                className="action-btn download-btn"
                                                onClick={() => handleDownload(recording)}
                                                title="Tai xuong"
                                            >
                                                <svg viewBox="0 0 24 24" fill="currentColor">
                                                    <path d="M19 9h-4V3H9v6H5l7 7 7-7zM5 18v2h14v-2H5z" />
                                                </svg>
                                            </button>
                                            <button
                                                className="action-btn delete-btn"
                                                onClick={() => handleDelete(recording)}
                                                title="Xoa"
                                            >
                                                <svg viewBox="0 0 24 24" fill="currentColor">
                                                    <path d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM19 4h-3.5l-1-1h-5l-1 1H5v2h14V4z" />
                                                </svg>
                                            </button>
                                        </div>
                                    </div>
                                ))}
                            </div>
                        )}
                    </div>
                </div>
            </div>

        </>
    )
}

export default RecordingsPanel
