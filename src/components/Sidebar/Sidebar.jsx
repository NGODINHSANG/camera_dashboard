import { useState } from 'react'
import './Sidebar.css'

function Sidebar({
    projects,
    selectedProject,
    onProjectSelect,
    onAddProject,
    onEditProject,
    onDeleteProject,
    isAdmin = false
}) {
    const [isCollapsed, setIsCollapsed] = useState(false)

    return (
        <aside className={`sidebar ${isCollapsed ? 'collapsed' : ''}`}>
            <div className="sidebar-content">
                {/* Header với nút toggle */}
                <div className="sidebar-header">
                    <h2 className="sidebar-title">DỰ ÁN</h2>
                    <div className="header-actions">
                        {isAdmin && (
                            <button
                                className="add-project-btn"
                                onClick={onAddProject}
                                title="Thêm dự án mới"
                            >
                                <svg viewBox="0 0 24 24" fill="currentColor">
                                    <path d="M19 13h-6v6h-2v-6H5v-2h6V5h2v6h6v2z" />
                                </svg>
                            </button>
                        )}
                        <button
                            className="toggle-btn"
                            onClick={() => setIsCollapsed(!isCollapsed)}
                            title={isCollapsed ? 'Mở rộng' : 'Thu gọn'}
                        >
                            <svg viewBox="0 0 24 24" fill="currentColor">
                                {isCollapsed ? (
                                    <path d="M10 6L8.59 7.41 13.17 12l-4.58 4.59L10 18l6-6z" />
                                ) : (
                                    <path d="M15.41 7.41L14 6l-6 6 6 6 1.41-1.41L10.83 12z" />
                                )}
                            </svg>
                        </button>
                    </div>
                </div>

                {/* Danh sách dự án */}
                <nav className="project-list">
                    {projects.map((project) => {
                        const isActive = selectedProject?.id === project.id
                        const camCount = project.cameras?.length || 0
                        return (
                            <div key={project.id} className="project-node">
                                <div className={`project-item ${isActive ? 'active' : ''}`}
                                    onClick={() => onProjectSelect(project)}
                                >
                                    {/* Icon */}
                                    <div className="project-icon">
                                        <svg viewBox="0 0 24 24" fill="currentColor">
                                            <path d="M17 2H7C5.9 2 5 2.9 5 4v16c0 1.1.9 2 2 2h10c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2zm-5 14c-1.1 0-2-.9-2-2s.9-2 2-2 2 .9 2 2-.9 2-2 2zm3-8H9V6h6v2z"/>
                                        </svg>
                                    </div>

                                    {/* Info */}
                                    <div className="project-info">
                                        <span className="project-name-text">{project.name}</span>
                                        {isActive && <span className="project-selected-label">SELECTED</span>}
                                    </div>

                                    {/* Right side */}
                                    <div className="project-right">
                                        <span className="camera-count">{camCount} Cameras</span>
                                        {isAdmin && (
                                            <div className="project-actions" onClick={e => e.stopPropagation()}>
                                                <button className="edit-project-btn" onClick={() => onEditProject(project)} title="Chỉnh sửa">
                                                    <svg viewBox="0 0 24 24" fill="currentColor">
                                                        <path d="M3 17.25V21h3.75L17.81 9.94l-3.75-3.75L3 17.25zM20.71 7.04c.39-.39.39-1.02 0-1.41l-2.34-2.34c-.39-.39-1.02-.39-1.41 0l-1.83 1.83 3.75 3.75 1.83-1.83z" />
                                                    </svg>
                                                </button>
                                                <button className="delete-project-btn" onClick={() => onDeleteProject(project)} title="Xóa">
                                                    <svg viewBox="0 0 24 24" fill="currentColor">
                                                        <path d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM19 4h-3.5l-1-1h-5l-1 1H5v2h14V4z" />
                                                    </svg>
                                                </button>
                                            </div>
                                        )}
                                    </div>
                                </div>
                            </div>
                        )
                    })}
                    {projects.length === 0 && (
                        <p className="no-projects">Chưa có dự án nào</p>
                    )}
                </nav>
            </div>

            {/* Collapsed state - chỉ hiện icon */}
            {isCollapsed && (
                <div className="sidebar-collapsed-content">
                    <button
                        className="expand-btn"
                        onClick={() => setIsCollapsed(false)}
                        title="Mở menu dự án"
                    >
                        <svg viewBox="0 0 24 24" fill="currentColor">
                            <path d="M3 13h2v-2H3v2zm0 4h2v-2H3v2zm0-8h2V7H3v2zm4 4h14v-2H7v2zm0 4h14v-2H7v2zM7 7v2h14V7H7z" />
                        </svg>
                    </button>
                </div>
            )}
        </aside>
    )
}

export default Sidebar
