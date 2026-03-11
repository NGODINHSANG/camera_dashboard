import { useState, useEffect, useCallback } from 'react'
import CameraFrame from '../CameraFrame/CameraFrame'
import RecordingsPanel from '../Modal/RecordingsPanel'
import { recordingsApi } from '../../api/recordings'
import './CameraGrid.css'

function CameraGrid({
    cameras,
    onStatusChange,
    onCameraClick,
    selectedCamera,
    gridLayout = '2x2',
    onEditCamera,
    onDeleteCamera,
    onAddCamera,
    hasProject,
    isAdmin = false,
    selectedProject
}) {
    // Recording state
    const [recordingStates, setRecordingStates] = useState({}) // { cameraId: { isRecording, recordingId } }
    // Recordings panel state
    const [showRecordingsPanel, setShowRecordingsPanel] = useState(false)
    const [cameraForRecordings, setCameraForRecordings] = useState(null)
    // Playback state: { cameraId: { name, path, streamUrl, size, modTime } }
    const [playbackRecordings, setPlaybackRecordings] = useState({})
    // Playlist state: { cameraId: { playlist: [...], currentIndex: 0 } }
    const [playlistStates, setPlaylistStates] = useState({})

    // Load active recordings on mount
    useEffect(() => {
        loadActiveRecordings()
    }, [])

    const loadActiveRecordings = async () => {
        try {
            console.log('[CameraGrid] Loading active recordings...')
            const recordings = await recordingsApi.getActive()
            console.log('[CameraGrid] Active recordings response:', recordings)
            const states = {}
            // recordings is array directly, not response.data
            if (Array.isArray(recordings)) {
                recordings.forEach(rec => {
                    states[rec.cameraId] = { isRecording: true, recordingId: rec.id }
                })
            }
            console.log('[CameraGrid] Recording states:', states)
            setRecordingStates(states)
        } catch (err) {
            console.error('Failed to load active recordings:', err)
        }
    }

    const handleRecordClick = useCallback(async (camera) => {
        const state = recordingStates[camera.id]
        if (state?.isRecording) {
            // Stop recording
            handleStopRecording(camera.id, state.recordingId)
        } else {
            // Start recording directly (auto-saves to /recordings/<project>/<camera>)
            try {
                const response = await recordingsApi.start(camera.id)
                const recording = response.data || response
                setRecordingStates(prev => ({
                    ...prev,
                    [camera.id]: { isRecording: true, recordingId: recording.id }
                }))
            } catch (err) {
                console.error('Failed to start recording:', err)
                alert('Loi bat dau ghi hinh: ' + (err.response?.data?.error || err.message || 'Unknown error'))
            }
        }
    }, [recordingStates])

    const handleViewRecordings = useCallback((camera) => {
        setCameraForRecordings({ ...camera, projectName: selectedProject?.name })
        setShowRecordingsPanel(true)
    }, [selectedProject])

    // Handle playing a recording in the camera frame
    const handlePlayRecording = useCallback((cameraId, recording) => {
        console.log('handlePlayRecording called:', { cameraId, recording })
        setPlaybackRecordings(prev => ({
            ...prev,
            [cameraId]: recording
        }))
        setShowRecordingsPanel(false)
        setCameraForRecordings(null)
    }, [])

    // Handle playing a playlist (loop through videos)
    const handlePlayPlaylist = useCallback((cameraId, playlist, startIndex = 0) => {
        console.log('handlePlayPlaylist called:', { cameraId, playlistLength: playlist.length, startIndex })
        if (!playlist || playlist.length === 0) return

        // Set playlist state with starting index
        setPlaylistStates(prev => ({
            ...prev,
            [cameraId]: { playlist, currentIndex: startIndex }
        }))

        // Start playing from selected video
        setPlaybackRecordings(prev => ({
            ...prev,
            [cameraId]: playlist[startIndex]
        }))

        setShowRecordingsPanel(false)
        setCameraForRecordings(null)
    }, [])

    // Handle video ended - play next in playlist
    const handleVideoEnded = useCallback((cameraId) => {
        const state = playlistStates[cameraId]
        if (!state || !state.playlist || state.playlist.length === 0) return

        const nextIndex = (state.currentIndex + 1) % state.playlist.length
        console.log('Video ended, playing next:', { cameraId, nextIndex, total: state.playlist.length })

        // Update playlist index
        setPlaylistStates(prev => ({
            ...prev,
            [cameraId]: { ...prev[cameraId], currentIndex: nextIndex }
        }))

        // Play next video
        setPlaybackRecordings(prev => ({
            ...prev,
            [cameraId]: state.playlist[nextIndex]
        }))
    }, [playlistStates])

    // Handle exiting playback mode
    const handleExitPlayback = useCallback((cameraId) => {
        setPlaybackRecordings(prev => {
            const newState = { ...prev }
            delete newState[cameraId]
            return newState
        })
        // Clear playlist state
        setPlaylistStates(prev => {
            const newState = { ...prev }
            delete newState[cameraId]
            return newState
        })
    }, [])

    const handleStopRecording = async (cameraId, recordingId) => {
        // Log stack trace to find who called this
        console.log('[CameraGrid] handleStopRecording called:', { cameraId, recordingId })
        console.trace('[CameraGrid] Stop recording stack trace')

        try {
            await recordingsApi.stop(recordingId)
            setRecordingStates(prev => {
                const newStates = { ...prev }
                delete newStates[cameraId]
                return newStates
            })
        } catch (err) {
            console.error('Failed to stop recording:', err)
            // Reset recording state anyway so user can try again
            setRecordingStates(prev => {
                const newStates = { ...prev }
                delete newStates[cameraId]
                return newStates
            })
            alert('Lỗi dừng ghi hình: ' + (err.message || 'Request failed') + '\nĐã reset trạng thái.')
        }
    }

    // Nếu có camera được chọn, chỉ hiển thị camera đó ở chế độ fullscreen
    if (selectedCamera) {
        return (
            <>
                <div className="camera-grid fullscreen">
                    <CameraFrame
                        camera={selectedCamera}
                        onStatusChange={onStatusChange}
                        isFullscreen={true}
                        onEdit={() => onEditCamera(selectedCamera)}
                        onDelete={() => onDeleteCamera(selectedCamera)}
                        isAdmin={isAdmin}
                        isRecordingActive={recordingStates[selectedCamera.id]?.isRecording || false}
                        onRecordClick={() => handleRecordClick(selectedCamera)}
                        onViewRecordings={() => handleViewRecordings(selectedCamera)}
                        playbackRecording={playbackRecordings[selectedCamera.id] || null}
                        onExitPlayback={() => handleExitPlayback(selectedCamera.id)}
                        onVideoEnded={() => handleVideoEnded(selectedCamera.id)}
                        playlistInfo={playlistStates[selectedCamera.id] || null}
                    />
                </div>
                {/* Recordings Panel */}
                {showRecordingsPanel && cameraForRecordings && (
                    <RecordingsPanel
                        camera={cameraForRecordings}
                        onClose={() => {
                            setShowRecordingsPanel(false)
                            setCameraForRecordings(null)
                        }}
                        onPlayRecording={(recording) => handlePlayRecording(cameraForRecordings.id, recording)}
                        onPlayPlaylist={(playlist, startIndex) => handlePlayPlaylist(cameraForRecordings.id, playlist, startIndex)}
                    />
                )}
            </>
        )
    }

    // Chế độ grid bình thường với layout động
    const gridClass = `camera-grid grid-${gridLayout}`

    return (
        <div className="camera-grid-container">
            <div className={gridClass}>
                {cameras.map((camera) => (
                    <CameraFrame
                        key={camera.id}
                        camera={camera}
                        onStatusChange={onStatusChange}
                        onClick={() => onCameraClick(camera)}
                        isFullscreen={false}
                        onEdit={() => onEditCamera(camera)}
                        onDelete={() => onDeleteCamera(camera)}
                        isAdmin={isAdmin}
                        isRecordingActive={recordingStates[camera.id]?.isRecording || false}
                        onRecordClick={() => handleRecordClick(camera)}
                        onViewRecordings={() => handleViewRecordings(camera)}
                        playbackRecording={playbackRecordings[camera.id] || null}
                        onExitPlayback={() => handleExitPlayback(camera.id)}
                        onVideoEnded={() => handleVideoEnded(camera.id)}
                        playlistInfo={playlistStates[camera.id] || null}
                    />
                ))}
                {hasProject && cameras.length === 0 && (
                    <div className="no-cameras-message">
                        <svg viewBox="0 0 24 24" fill="currentColor">
                            <path d="M18 10.48V6c0-1.1-.9-2-2-2H4c-1.1 0-2 .9-2 2v12c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2v-4.48l4 3.98v-11l-4 3.98zm-2-.79V18H4V6h12v3.69z" />
                        </svg>
                        <p>Chưa có camera nào trong dự án này</p>
                        {isAdmin && (
                            <button className="add-first-camera-btn" onClick={onAddCamera}>
                                Thêm camera đầu tiên
                            </button>
                        )}
                    </div>
                )}
            </div>

            {/* Recordings Panel */}
            {showRecordingsPanel && cameraForRecordings && (
                <RecordingsPanel
                    camera={cameraForRecordings}
                    onClose={() => {
                        setShowRecordingsPanel(false)
                        setCameraForRecordings(null)
                    }}
                    onPlayRecording={(recording) => handlePlayRecording(cameraForRecordings.id, recording)}
                    onPlayPlaylist={(playlist, startIndex) => handlePlayPlaylist(cameraForRecordings.id, playlist, startIndex)}
                />
            )}
        </div>
    )
}

export default CameraGrid

