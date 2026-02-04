import { useState, useEffect, useCallback } from 'react'
import { recordingsApi } from '../../api/recordings'
import './FolderBrowser.css'

// Component hiển thị 1 folder với khả năng expand
function FolderNode({ name, path, level = 0, onSelect, selectedPath }) {
    const [isExpanded, setIsExpanded] = useState(false)
    const [children, setChildren] = useState([])
    const [loading, setLoading] = useState(false)
    const isSelected = selectedPath === path

    const handleToggle = async (e) => {
        e.stopPropagation()

        if (isExpanded) {
            setIsExpanded(false)
            return
        }

        setLoading(true)
        try {
            const response = await recordingsApi.browse(path)
            setChildren(response.directories || [])
            setIsExpanded(true)
        } catch (err) {
            console.error('Failed to load:', err)
        } finally {
            setLoading(false)
        }
    }

    const handleSelect = () => {
        onSelect(path)
    }

    return (
        <div className="folder-node">
            <div
                className={`folder-row ${isSelected ? 'selected' : ''}`}
                style={{ paddingLeft: `${level * 20 + 12}px` }}
                onClick={handleSelect}
            >
                <button
                    className={`expand-btn ${isExpanded ? 'expanded' : ''}`}
                    onClick={handleToggle}
                >
                    {loading ? (
                        <div className="mini-spinner"></div>
                    ) : (
                        <svg viewBox="0 0 24 24" fill="currentColor">
                            <path d="M10 6L8.59 7.41 13.17 12l-4.58 4.59L10 18l6-6z" />
                        </svg>
                    )}
                </button>
                <svg className="folder-icon" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M10 4H4c-1.1 0-2 .9-2 2v12c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2h-8l-2-2z" />
                </svg>
                <span className="folder-name">{name}</span>
            </div>

            {isExpanded && children.length > 0 && (
                <div className="folder-children">
                    {children.map(child => (
                        <FolderNode
                            key={child.path}
                            name={child.name}
                            path={child.path}
                            level={level + 1}
                            onSelect={onSelect}
                            selectedPath={selectedPath}
                        />
                    ))}
                </div>
            )}
        </div>
    )
}

function FolderBrowser({ isOpen, onClose, onSelect }) {
    const [drives, setDrives] = useState([])
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState('')
    const [selectedPath, setSelectedPath] = useState('')

    // Load drives when opened
    useEffect(() => {
        if (isOpen) {
            loadDrives()
        }
    }, [isOpen])

    const loadDrives = async () => {
        setLoading(true)
        setError('')
        try {
            const response = await recordingsApi.browse('')
            setDrives(response.directories || [])
        } catch (err) {
            setError(err.message || 'Không thể đọc ổ đĩa')
        } finally {
            setLoading(false)
        }
    }

    const handleFolderSelect = useCallback((path) => {
        setSelectedPath(path)
    }, [])

    const handleConfirm = () => {
        if (selectedPath) {
            onSelect(selectedPath)
            onClose()
        }
    }

    if (!isOpen) return null

    return (
        <div className="folder-browser-overlay" onClick={onClose}>
            <div className="folder-browser" onClick={e => e.stopPropagation()}>
                <div className="folder-browser-header">
                    <h3>Chọn thư mục lưu video</h3>
                    <button className="close-btn" onClick={onClose}>
                        <svg viewBox="0 0 24 24" fill="currentColor">
                            <path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12 19 6.41z" />
                        </svg>
                    </button>
                </div>

                {selectedPath && (
                    <div className="folder-browser-path">
                        <span className="path-label">Đã chọn:</span>
                        <span className="path-value">{selectedPath}</span>
                    </div>
                )}

                {error && (
                    <div className="folder-browser-error">{error}</div>
                )}

                <div className="folder-browser-content">
                    {loading ? (
                        <div className="folder-browser-loading">
                            <div className="loading-spinner"></div>
                            <span>Đang tải...</span>
                        </div>
                    ) : (
                        <div className="folder-tree">
                            {drives.map(drive => (
                                <FolderNode
                                    key={drive.path}
                                    name={drive.name}
                                    path={drive.path}
                                    level={0}
                                    onSelect={handleFolderSelect}
                                    selectedPath={selectedPath}
                                />
                            ))}
                        </div>
                    )}
                </div>

                <div className="folder-browser-actions">
                    <button className="btn btn-secondary" onClick={onClose}>
                        Hủy
                    </button>
                    <button
                        className="btn btn-primary"
                        onClick={handleConfirm}
                        disabled={!selectedPath}
                    >
                        <svg viewBox="0 0 24 24" fill="currentColor">
                            <path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z" />
                        </svg>
                        Chọn thư mục này
                    </button>
                </div>
            </div>
        </div>
    )
}

export default FolderBrowser
