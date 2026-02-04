import { useState, useEffect, useCallback } from 'react'
import CameraFrame from '../CameraFrame/CameraFrame'
import RecordingDialog from '../Modal/RecordingDialog'
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
    const [showRecordingDialog, setShowRecordingDialog] = useState(false)
    const [cameraToRecord, setCameraToRecord] = useState(null)

    // Load active recordings on mount
    useEffect(() => {
        loadActiveRecordings()
    }, [])

    const loadActiveRecordings = async () => {
        try {
            const recordings = await recordingsApi.getActive()
            const states = {}
            // recordings is array directly, not response.data
            if (Array.isArray(recordings)) {
                recordings.forEach(rec => {
                    states[rec.cameraId] = { isRecording: true, recordingId: rec.id }
                })
            }
            setRecordingStates(states)
        } catch (err) {
            console.error('Failed to load active recordings:', err)
        }
    }

    const handleRecordClick = useCallback((camera) => {
        const state = recordingStates[camera.id]
        if (state?.isRecording) {
            // Stop recording
            handleStopRecording(camera.id, state.recordingId)
        } else {
            // Show dialog to start recording
            setCameraToRecord({ ...camera, projectName: selectedProject?.name })
            setShowRecordingDialog(true)
        }
    }, [recordingStates, selectedProject])

    const handleStartRecording = async (outputDir) => {
        if (!cameraToRecord) return

        try {
            const recording = await recordingsApi.start(cameraToRecord.id, outputDir)
            setRecordingStates(prev => ({
                ...prev,
                [cameraToRecord.id]: { isRecording: true, recordingId: recording.id }
            }))
            setShowRecordingDialog(false)
            setCameraToRecord(null)
        } catch (err) {
            console.error('Failed to start recording:', err)
            alert('Lỗi bắt đầu ghi hình: ' + (err.message || 'Unknown error'))
        }
    }

    const handleStopRecording = async (cameraId, recordingId) => {
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
                />
            </div>
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

            {/* Recording Dialog */}
            {showRecordingDialog && cameraToRecord && (
                <RecordingDialog
                    camera={cameraToRecord}
                    onStart={handleStartRecording}
                    onCancel={() => {
                        setShowRecordingDialog(false)
                        setCameraToRecord(null)
                    }}
                />
            )}
        </div>
    )
}

export default CameraGrid

