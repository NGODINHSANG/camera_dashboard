import { useState } from 'react'
import FolderBrowser from './FolderBrowser'
import { recordingsApi } from '../../api/recordings'
import './RecordingDialog.css'

function RecordingDialog({ camera, onStart, onCancel }) {
    const [outputDir, setOutputDir] = useState('/home')
    const [showFolderBrowser, setShowFolderBrowser] = useState(false)
    const [isValidating, setIsValidating] = useState(false)
    const [error, setError] = useState('')

    const handleSubmit = async (e) => {
        e.preventDefault()
        if (!outputDir.trim()) return

        // Validate path before starting
        setIsValidating(true)
        setError('')

        try {
            await recordingsApi.validatePath(outputDir.trim())
            // Path is valid, start recording
            onStart(outputDir.trim())
        } catch (err) {
            setError(err.message || 'Đường dẫn không hợp lệ hoặc không có quyền ghi')
            setIsValidating(false)
        }
    }

    const handleFolderSelect = (path) => {
        setOutputDir(path)
        setError('')
    }

    return (
        <>
            <div className="modal-overlay" onClick={onCancel}>
                <div className="modal-content recording-dialog" onClick={e => e.stopPropagation()}>
                    <div className="modal-header">
                        <h3>Ghi hình Camera</h3>
                        <button className="close-btn" onClick={onCancel}>
                            <svg viewBox="0 0 24 24" fill="currentColor">
                                <path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12 19 6.41z" />
                            </svg>
                        </button>
                    </div>

                    <form onSubmit={handleSubmit}>
                        <div className="camera-info-section">
                            <div className="info-row">
                                <span className="label">Camera:</span>
                                <span className="value">{camera?.name}</span>
                            </div>
                            <div className="info-row">
                                <span className="label">Vị trí:</span>
                                <span className="value">{camera?.location}</span>
                            </div>
                        </div>

                        {error && (
                            <div className="recording-error">
                                <svg viewBox="0 0 24 24" fill="currentColor">
                                    <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z" />
                                </svg>
                                {error}
                            </div>
                        )}

                        <div className="form-group">
                            <label className="form-label">Thư mục lưu video (trên server)</label>
                            <div className="folder-input-group">
                                <input
                                    type="text"
                                    className="form-input"
                                    value={outputDir}
                                    onChange={(e) => {
                                        setOutputDir(e.target.value)
                                        setError('')
                                    }}
                                    placeholder="Nhập hoặc chọn thư mục..."
                                    required
                                />
                                <button
                                    type="button"
                                    className="btn-browse"
                                    onClick={() => setShowFolderBrowser(true)}
                                >
                                    <svg viewBox="0 0 24 24" fill="currentColor">
                                        <path d="M10 4H4c-1.1 0-2 .9-2 2v12c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2h-8l-2-2z" />
                                    </svg>
                                    Duyệt...
                                </button>
                            </div>
                            <span className="form-hint">
                                Ví dụ: /home/recordings hoặc /var/video
                            </span>
                        </div>

                        <div className="form-group">
                            <label className="form-label">Tên file sẽ được tạo</label>
                            <div className="filename-preview">
                                {camera?.projectName || 'Project'}_{camera?.name || 'Camera'}_{new Date().toISOString().slice(0, 10)}_HHMMSS.mp4
                            </div>
                        </div>

                        <div className="modal-actions">
                            <button type="button" className="btn btn-secondary" onClick={onCancel}>
                                Hủy
                            </button>
                            <button type="submit" className="btn btn-primary btn-record" disabled={isValidating}>
                                {isValidating ? (
                                    <>
                                        <div className="btn-spinner"></div>
                                        Đang kiểm tra...
                                    </>
                                ) : (
                                    <>
                                        <svg viewBox="0 0 24 24" fill="currentColor">
                                            <circle cx="12" cy="12" r="8" />
                                        </svg>
                                        Bắt đầu ghi hình
                                    </>
                                )}
                            </button>
                        </div>
                    </form>
                </div>
            </div>

            {/* Folder Browser Modal */}
            <FolderBrowser
                isOpen={showFolderBrowser}
                onClose={() => setShowFolderBrowser(false)}
                onSelect={handleFolderSelect}
                initialPath={outputDir}
            />
        </>
    )
}

export default RecordingDialog
