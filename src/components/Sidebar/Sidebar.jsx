import { useState, useEffect } from 'react'
import { camerasApi } from '../../api/cameras'
import './Sidebar.css'

function Sidebar({
    projects,
    selectedProject,
    onProjectSelect,
    onAddProject,
    onEditProject,
    onDeleteProject,
    onCameraFullscreen,
    isAdmin = false
}) {
    const [isCollapsed, setIsCollapsed] = useState(false)
    const [expandedProjects, setExpandedProjects] = useState({})
    const [projectCameras, setProjectCameras] = useState({})
    const [loadingCameras, setLoadingCameras] = useState({})

    // Toggle expand/collapse project
    const handleProjectClick = async (project) => {
        const projectId = project.id
        const isExpanded = expandedProjects[projectId]

        if (isExpanded) {
            // Collapse
            setExpandedProjects(prev => ({ ...prev, [projectId]: false }))
        } else {
            // Expand and load cameras if not already loaded
            setExpandedProjects(prev => ({ ...prev, [projectId]: true }))

            if (!projectCameras[projectId]) {
                setLoadingCameras(prev => ({ ...prev, [projectId]: true }))
                try {
                    const response = await camerasApi.getByProject(projectId)
                    // API returns axios response, extract data
                    const cameras = response.data || response || []
                    setProjectCameras(prev => ({ ...prev, [projectId]: cameras }))
                } catch (err) {
                    console.error('Failed to load cameras:', err)
                    setProjectCameras(prev => ({ ...prev, [projectId]: [] }))
                } finally {
                    setLoadingCameras(prev => ({ ...prev, [projectId]: false }))
                }
            }
        }

        // Also select the project
        onProjectSelect(project)
    }

    // Handle camera click - show fullscreen
    const handleCameraClick = (camera, project) => {
        if (onCameraFullscreen) {
            onCameraFullscreen(camera, project)
        }
    }

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
                    {projects.map((project) => (
                        <div key={project.id} className="project-node">
                            <div className={`project-item ${selectedProject?.id === project.id ? 'active' : ''}`}>
                                <button
                                    className="project-btn"
                                    onClick={() => handleProjectClick(project)}
                                    title={project.name}
                                >
                                    <span className={`expand-arrow ${expandedProjects[project.id] ? 'expanded' : ''}`}>
                                        <svg viewBox="0 0 24 24" fill="currentColor">
                                            <path d="M10 6L8.59 7.41 13.17 12l-4.58 4.59L10 18l6-6z" />
                                        </svg>
                                    </span>
                                    <span className="project-name">{project.name}</span>
                                    <span className="camera-count">{project.cameras?.length || 0}</span>
                                </button>
                                {isAdmin && (
                                    <div className="project-actions">
                                        <button
                                            className="edit-project-btn"
                                            onClick={(e) => {
                                                e.stopPropagation()
                                                onEditProject(project)
                                            }}
                                            title="Chỉnh sửa dự án"
                                        >
                                            <svg viewBox="0 0 24 24" fill="currentColor">
                                                <path d="M3 17.25V21h3.75L17.81 9.94l-3.75-3.75L3 17.25zM20.71 7.04c.39-.39.39-1.02 0-1.41l-2.34-2.34c-.39-.39-1.02-.39-1.41 0l-1.83 1.83 3.75 3.75 1.83-1.83z" />
                                            </svg>
                                        </button>
                                        <button
                                            className="delete-project-btn"
                                            onClick={(e) => {
                                                e.stopPropagation()
                                                onDeleteProject(project)
                                            }}
                                            title="Xóa dự án"
                                        >
                                            <svg viewBox="0 0 24 24" fill="currentColor">
                                                <path d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM19 4h-3.5l-1-1h-5l-1 1H5v2h14V4z" />
                                            </svg>
                                        </button>
                                    </div>
                                )}
                            </div>

                            {/* Camera list khi expand */}
                            {expandedProjects[project.id] && (
                                <div className="camera-list">
                                    {loadingCameras[project.id] ? (
                                        <div className="camera-loading">
                                            <div className="mini-spinner"></div>
                                            <span>Đang tải...</span>
                                        </div>
                                    ) : projectCameras[project.id]?.length > 0 ? (
                                        projectCameras[project.id].map(camera => (
                                            <div
                                                key={camera.id}
                                                className="camera-item"
                                                onClick={() => handleCameraClick(camera, project)}
                                            >
                                                <svg className="camera-icon" viewBox="0 0 24 24" fill="currentColor">
                                                    <path d="M17 10.5V7c0-.55-.45-1-1-1H4c-.55 0-1 .45-1 1v10c0 .55.45 1 1 1h12c.55 0 1-.45 1-1v-3.5l4 4v-11l-4 4z" />
                                                </svg>
                                                <span className="camera-name">{camera.name}</span>
                                            </div>
                                        ))
                                    ) : (
                                        <div className="no-cameras">
                                            Chưa có camera
                                        </div>
                                    )}
                                </div>
                            )}
                        </div>
                    ))}
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
