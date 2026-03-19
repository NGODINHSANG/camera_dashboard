import { useState, useCallback, useEffect } from 'react'
import { Routes, Route } from 'react-router-dom'
import { useAuth } from './contexts/AuthContext'
import { projectsApi } from './api/projects'
import { camerasApi } from './api/cameras'
import Header from './components/Header/Header'
import Sidebar from './components/Sidebar/Sidebar'
import CameraGrid from './components/CameraGrid/CameraGrid'
import Footer from './components/Footer/Footer'
import Modal from './components/Modal/Modal'
import AddCameraForm from './components/Modal/AddCameraForm'
import EditCameraForm from './components/Modal/EditCameraForm'
import AddProjectForm from './components/Modal/AddProjectForm'
import EditProjectForm from './components/Modal/EditProjectForm'
import ConfirmDialog from './components/Modal/ConfirmDialog'
import ProtectedRoute from './components/ProtectedRoute'
import LoginPage from './pages/LoginPage'
import RegisterPage from './pages/RegisterPage'
import AdminPage from './pages/AdminPage'
import SettingsPage from './pages/SettingsPage'
import SupportPage from './pages/SupportPage'
import ProfilePage from './pages/ProfilePage'
import './App.css'

function Dashboard() {
    const { user, logout, isAdmin } = useAuth()

    // State cho dữ liệu từ API
    const [projects, setProjects] = useState([])
    const [selectedProjectId, setSelectedProjectId] = useState(null)
    const [gridLayout, setGridLayout] = useState(() => {
        return localStorage.getItem('dashboard_grid_layout') || '2x2'
    })
    const [aspectRatio, setAspectRatio] = useState(() => {
        return localStorage.getItem('dashboard_aspect_ratio') || '16:9'
    })
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState(null)

    // State tạm (không cần lưu)
    const [cameraStatus, setCameraStatus] = useState({})
    const [isLive, setIsLive] = useState(true)
    const [selectedCamera, setSelectedCamera] = useState(null)

    // State cho modals - Camera
    const [showAddCameraModal, setShowAddCameraModal] = useState(false)
    const [showEditCameraModal, setShowEditCameraModal] = useState(false)
    const [showDeleteCameraModal, setShowDeleteCameraModal] = useState(false)
    const [cameraToEdit, setCameraToEdit] = useState(null)
    const [cameraToDelete, setCameraToDelete] = useState(null)

    // State cho modals - Project
    const [showAddProjectModal, setShowAddProjectModal] = useState(false)
    const [showEditProjectModal, setShowEditProjectModal] = useState(false)
    const [showDeleteProjectModal, setShowDeleteProjectModal] = useState(false)
    const [projectToEdit, setProjectToEdit] = useState(null)
    const [projectToDelete, setProjectToDelete] = useState(null)

    // State cho modal đăng xuất
    const [showLogoutModal, setShowLogoutModal] = useState(false)

    // Lưu settings vào localStorage
    useEffect(() => {
        localStorage.setItem('dashboard_grid_layout', gridLayout)
    }, [gridLayout])

    useEffect(() => {
        localStorage.setItem('dashboard_aspect_ratio', aspectRatio)
    }, [aspectRatio])

    // Load projects từ API
    useEffect(() => {
        loadProjects()
    }, [])

    const loadProjects = async () => {
        try {
            setLoading(true)
            const response = await projectsApi.getAll()
            setProjects(response.data || [])
            if (response.data?.length > 0 && !selectedProjectId) {
                setSelectedProjectId(response.data[0].id)
            }
        } catch (err) {
            setError(err.message)
        } finally {
            setLoading(false)
        }
    }

    // Tìm project được chọn từ ID
    const selectedProject = projects.find(p => p.id === selectedProjectId) || projects[0] || null

    // Camera hiện tại dựa trên dự án được chọn
    const cameras = selectedProject?.cameras || []

    // Cập nhật selectedProjectId nếu project không tồn tại
    useEffect(() => {
        if (projects.length > 0 && !projects.find(p => p.id === selectedProjectId)) {
            setSelectedProjectId(projects[0].id)
        }
    }, [projects, selectedProjectId])

    const handleStatusChange = useCallback((cameraId, isConnected) => {
        setCameraStatus(prev => ({
            ...prev,
            [cameraId]: isConnected
        }))

        const hasAnyConnection = Object.values({ ...cameraStatus, [cameraId]: isConnected }).some(status => status)
        setIsLive(hasAnyConnection)
    }, [cameraStatus])

    // Click vào camera để xem fullscreen
    const handleCameraClick = useCallback((camera) => {
        setSelectedCamera(camera)
    }, [])

    // Quay lại grid view
    const handleBackToGrid = useCallback(() => {
        setSelectedCamera(null)
    }, [])

    // Chọn dự án
    const handleProjectSelect = useCallback((project) => {
        setSelectedProjectId(project.id)
        setSelectedCamera(null)
    }, [])

    // Thay đổi grid layout
    const handleGridLayoutChange = useCallback((layout) => {
        setGridLayout(layout)
    }, [])

    // === QUẢN LÝ DỰ ÁN ===

    // Thêm dự án mới
    const handleAddProject = useCallback(async (newProjectData) => {
        try {
            const response = await projectsApi.create(newProjectData.name)
            const newProject = response.data
            setProjects(prev => [...prev, newProject])
            setSelectedProjectId(newProject.id)
            setShowAddProjectModal(false)
        } catch (err) {
            alert('Lỗi: ' + err.message)
        }
    }, [])

    // Mở dialog chỉnh sửa dự án
    const handleEditProjectClick = useCallback((project) => {
        setProjectToEdit(project)
        setShowEditProjectModal(true)
    }, [])

    // Lưu thay đổi dự án
    const handleEditProject = useCallback(async (updatedProjectData) => {
        if (!projectToEdit) return

        try {
            const response = await projectsApi.update(projectToEdit.id, updatedProjectData.name)
            const updatedProject = response.data

            setProjects(prevProjects => {
                return prevProjects.map(p =>
                    p.id === projectToEdit.id
                        ? { ...p, name: updatedProject.name }
                        : p
                )
            })

            setShowEditProjectModal(false)
            setProjectToEdit(null)
        } catch (err) {
            alert('Lỗi: ' + err.message)
        }
    }, [projectToEdit])

    const handleDeleteProjectClick = useCallback((project) => {
        setProjectToDelete(project)
        setShowDeleteProjectModal(true)
    }, [])

    const handleConfirmDeleteProject = useCallback(async () => {
        if (!projectToDelete) return

        try {
            await projectsApi.delete(projectToDelete.id)
            const newProjects = projects.filter(p => p.id !== projectToDelete.id)
            setProjects(newProjects)
            if (selectedProjectId === projectToDelete.id) {
                setSelectedProjectId(newProjects[0]?.id || null)
                setSelectedCamera(null)
            }
            setShowDeleteProjectModal(false)
            setProjectToDelete(null)
        } catch (err) {
            alert('Lỗi: ' + err.message)
        }
    }, [projectToDelete, selectedProjectId, projects])

    // === QUẢN LÝ CAMERA ===

    const handleAddCamera = useCallback(async (newCameraData) => {
        if (!selectedProject) return

        try {
            const response = await camerasApi.create(selectedProject.id, {
                name: newCameraData.name,
                location: newCameraData.location,
                streamUrl: newCameraData.streamUrl,
                isRecording: newCameraData.isRecording,
                autoRecord: newCameraData.autoRecord,
                aspectRatio: newCameraData.aspectRatio || 'auto'
            })
            const newCamera = response.data

            setProjects(prevProjects => {
                return prevProjects.map(project => {
                    if (project.id === selectedProject.id) {
                        return {
                            ...project,
                            cameras: [...project.cameras, newCamera]
                        }
                    }
                    return project
                })
            })

            setShowAddCameraModal(false)
        } catch (err) {
            alert('Lỗi: ' + err.message)
        }
    }, [selectedProject])

    // Mở dialog chỉnh sửa camera
    const handleEditCameraClick = useCallback((camera) => {
        setCameraToEdit(camera)
        setShowEditCameraModal(true)
    }, [])

    // Lưu thay đổi camera
    const handleEditCamera = useCallback(async (updatedCameraData) => {
        if (!selectedProject || !cameraToEdit) return

        try {
            const response = await camerasApi.update(selectedProject.id, cameraToEdit.id, {
                name: updatedCameraData.name,
                location: updatedCameraData.location,
                streamUrl: updatedCameraData.streamUrl,
                isRecording: updatedCameraData.isRecording,
                autoRecord: updatedCameraData.autoRecord,
                aspectRatio: updatedCameraData.aspectRatio || 'auto'
            })
            const updatedCamera = response.data

            setProjects(prevProjects => {
                return prevProjects.map(project => {
                    if (project.id === selectedProject.id) {
                        return {
                            ...project,
                            cameras: project.cameras.map(cam =>
                                cam.id === cameraToEdit.id ? updatedCamera : cam
                            )
                        }
                    }
                    return project
                })
            })

            // Update selectedCamera nếu đang xem fullscreen
            if (selectedCamera?.id === cameraToEdit.id) {
                setSelectedCamera(updatedCamera)
            }

            setShowEditCameraModal(false)
            setCameraToEdit(null)
        } catch (err) {
            alert('Lỗi: ' + err.message)
        }
    }, [selectedProject, cameraToEdit, selectedCamera])

    // Mở dialog xác nhận xóa camera
    const handleDeleteCameraClick = useCallback((camera) => {
        setCameraToDelete(camera)
        setShowDeleteCameraModal(true)
    }, [])

    // Xác nhận xóa camera
    const handleConfirmDeleteCamera = useCallback(async () => {
        if (!selectedProject || !cameraToDelete) return

        try {
            await camerasApi.delete(selectedProject.id, cameraToDelete.id)

            setProjects(prevProjects => {
                return prevProjects.map(project => {
                    if (project.id === selectedProject.id) {
                        return {
                            ...project,
                            cameras: project.cameras.filter(cam => cam.id !== cameraToDelete.id)
                        }
                    }
                    return project
                })
            })

            setShowDeleteCameraModal(false)
            setCameraToDelete(null)

            if (selectedCamera?.id === cameraToDelete.id) {
                setSelectedCamera(null)
            }
        } catch (err) {
            alert('Lỗi: ' + err.message)
        }
    }, [selectedProject, cameraToDelete, selectedCamera])

    // Xử lý đăng xuất
    const handleLogoutClick = useCallback(() => {
        setShowLogoutModal(true)
    }, [])

    const handleConfirmLogout = useCallback(() => {
        setShowLogoutModal(false)
        logout()
    }, [logout])

    // Tính số camera online/offline của dự án hiện tại
    const onlineCount = cameras.filter(cam => cameraStatus[cam.id]).length
    const offlineCount = cameras.length - onlineCount

    if (loading) {
        return (
            <div className="loading-screen">
                <div className="loading-spinner"></div>
                <p>Đang tải dữ liệu...</p>
            </div>
        )
    }

    return (
        <div className="app">
            <Header
                title={`DỰ ÁN ${selectedProject?.name || ''}`}
                onlineCount={onlineCount}
                offlineCount={offlineCount}
                showBackButton={selectedCamera !== null}
                onBackClick={handleBackToGrid}
                user={user}
                onLogout={handleLogoutClick}
                isAdmin={isAdmin}
                onAddCamera={() => setShowAddCameraModal(true)}
                hasProject={!!selectedProject}
            />
            <div className="app-body">
                <Sidebar
                    projects={projects}
                    selectedProject={selectedProject}
                    onProjectSelect={handleProjectSelect}
                    onAddProject={() => setShowAddProjectModal(true)}
                    onEditProject={handleEditProjectClick}
                    onDeleteProject={handleDeleteProjectClick}
                    onCameraFullscreen={(camera, project) => {
                        // Select the project first, then show camera fullscreen
                        setSelectedProjectId(project.id)
                        setSelectedCamera(camera)
                    }}
                    isAdmin={isAdmin}
                />
                <main className="main-content">
                    {error && <div className="error-banner">{error}</div>}
                    <CameraGrid
                        cameras={cameras}
                        onStatusChange={handleStatusChange}
                        onCameraClick={handleCameraClick}
                        selectedCamera={selectedCamera}
                        gridLayout={gridLayout}
                        onEditCamera={handleEditCameraClick}
                        onDeleteCamera={handleDeleteCameraClick}
                        onAddCamera={() => setShowAddCameraModal(true)}
                        hasProject={!!selectedProject}
                        isAdmin={isAdmin}
                        selectedProject={selectedProject}
                    />
                </main>
            </div>
            <Footer
                isLive={isLive}
                gridLayout={gridLayout}
                onGridLayoutChange={handleGridLayoutChange}
                aspectRatio={aspectRatio}
                onAspectRatioChange={setAspectRatio}
            />

            {/* Modal thêm dự án */}
            <Modal
                isOpen={showAddProjectModal}
                onClose={() => setShowAddProjectModal(false)}
                title="Thêm Dự Án Mới"
                size="small"
            >
                <AddProjectForm
                    onSubmit={handleAddProject}
                    onCancel={() => setShowAddProjectModal(false)}
                />
            </Modal>

            {/* Modal chỉnh sửa dự án */}
            <Modal
                isOpen={showEditProjectModal}
                onClose={() => {
                    setShowEditProjectModal(false)
                    setProjectToEdit(null)
                }}
                title="Chỉnh Sửa Dự Án"
                size="small"
            >
                <EditProjectForm
                    project={projectToEdit}
                    onSubmit={handleEditProject}
                    onCancel={() => {
                        setShowEditProjectModal(false)
                        setProjectToEdit(null)
                    }}
                />
            </Modal>

            {/* Modal xác nhận xóa dự án */}
            <Modal
                isOpen={showDeleteProjectModal}
                onClose={() => {
                    setShowDeleteProjectModal(false)
                    setProjectToDelete(null)
                }}
                title="Xác Nhận Xóa Dự Án"
                size="small"
            >
                <ConfirmDialog
                    message="Bạn có chắc chắn muốn xóa dự án này? Tất cả camera trong dự án sẽ bị xóa."
                    itemName={projectToDelete?.name}
                    onConfirm={handleConfirmDeleteProject}
                    onCancel={() => {
                        setShowDeleteProjectModal(false)
                        setProjectToDelete(null)
                    }}
                    confirmText="Xóa Dự Án"
                    type="danger"
                />
            </Modal>

            {/* Modal thêm camera */}
            <Modal
                isOpen={showAddCameraModal}
                onClose={() => setShowAddCameraModal(false)}
                title="Thêm Camera Mới"
                size="medium"
            >
                <AddCameraForm
                    onSubmit={handleAddCamera}
                    onCancel={() => setShowAddCameraModal(false)}
                />
            </Modal>

            {/* Modal chỉnh sửa camera */}
            <Modal
                isOpen={showEditCameraModal}
                onClose={() => {
                    setShowEditCameraModal(false)
                    setCameraToEdit(null)
                }}
                title="Chỉnh Sửa Camera"
                size="medium"
            >
                <EditCameraForm
                    camera={cameraToEdit}
                    onSubmit={handleEditCamera}
                    onCancel={() => {
                        setShowEditCameraModal(false)
                        setCameraToEdit(null)
                    }}
                />
            </Modal>

            {/* Modal xác nhận xóa camera */}
            <Modal
                isOpen={showDeleteCameraModal}
                onClose={() => {
                    setShowDeleteCameraModal(false)
                    setCameraToDelete(null)
                }}
                title="Xác Nhận Xóa Camera"
                size="small"
            >
                <ConfirmDialog
                    message="Bạn có chắc chắn muốn xóa camera này?"
                    itemName={cameraToDelete?.name}
                    onConfirm={handleConfirmDeleteCamera}
                    onCancel={() => {
                        setShowDeleteCameraModal(false)
                        setCameraToDelete(null)
                    }}
                    confirmText="Xóa"
                    type="danger"
                />
            </Modal>

            {/* Modal xác nhận đăng xuất */}
            <Modal
                isOpen={showLogoutModal}
                onClose={() => setShowLogoutModal(false)}
                title="Xác Nhận Đăng Xuất"
                size="small"
            >
                <ConfirmDialog
                    message="Bạn có chắc chắn muốn đăng xuất khỏi hệ thống?"
                    onConfirm={handleConfirmLogout}
                    onCancel={() => setShowLogoutModal(false)}
                    confirmText="Đăng xuất"
                    type="warning"
                />
            </Modal>
        </div>
    )
}

function App() {
    return (
        <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route path="/register" element={<RegisterPage />} />
            <Route
                path="/admin"
                element={
                    <ProtectedRoute adminOnly>
                        <AdminPage />
                    </ProtectedRoute>
                }
            />
            <Route
                path="/settings"
                element={
                    <ProtectedRoute>
                        <SettingsPage />
                    </ProtectedRoute>
                }
            />
            <Route
                path="/support"
                element={
                    <ProtectedRoute>
                        <SupportPage />
                    </ProtectedRoute>
                }
            />
            <Route
                path="/profile"
                element={
                    <ProtectedRoute>
                        <ProfilePage />
                    </ProtectedRoute>
                }
            />
            <Route
                path="/"
                element={
                    <ProtectedRoute>
                        <Dashboard />
                    </ProtectedRoute>
                }
            />
        </Routes>
    )
}

export default App
