import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { adminApi } from '../api/admin';
import Modal from '../components/Modal/Modal';
import ConfirmDialog from '../components/Modal/ConfirmDialog';
import './AuthPages.css';
import './AdminPage.css';

function AdminPage() {
    const { user } = useAuth();
    const [activeTab, setActiveTab] = useState('users');

    // Users tab state
    const [users, setUsers] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');
    const [showDeleteModal, setShowDeleteModal] = useState(false);
    const [userToDelete, setUserToDelete] = useState(null);

    // Project permission modal (per user)
    const [showPermModal, setShowPermModal] = useState(false);
    const [permUser, setPermUser] = useState(null);
    const [permUserProjectIDs, setPermUserProjectIDs] = useState([]);
    const [allProjects, setAllProjects] = useState([]);

    // Project members tab state
    const [selectedProjectId, setSelectedProjectId] = useState('');
    const [projectMembers, setProjectMembers] = useState([]);
    const [membersLoading, setMembersLoading] = useState(false);

    useEffect(() => {
        loadUsers();
        loadAllProjects();
    }, []);

    const loadUsers = async () => {
        try {
            const response = await adminApi.getUsers();
            setUsers(response.data);
        } catch (err) {
            setError(err.message);
        } finally {
            setLoading(false);
        }
    };

    const loadAllProjects = async () => {
        try {
            const response = await adminApi.getAllProjects();
            setAllProjects(response.data || []);
        } catch {}
    };

    // ── Users tab ────────────────────────────────────────────────
    const handleRoleChange = async (userId, currentRole) => {
        const newRole = currentRole === 'admin' ? 'user' : 'admin';
        try {
            await adminApi.updateUserRole(userId, newRole);
            setUsers(users.map(u => u.id === userId ? { ...u, role: newRole } : u));
        } catch (err) {
            alert('Lỗi: ' + err.message);
        }
    };

    const handleDeleteClick = (userItem) => {
        setUserToDelete(userItem);
        setShowDeleteModal(true);
    };

    const handleConfirmDelete = async () => {
        if (!userToDelete) return;
        try {
            await adminApi.deleteUser(userToDelete.id);
            setUsers(users.filter(u => u.id !== userToDelete.id));
            setShowDeleteModal(false);
            setUserToDelete(null);
        } catch (err) {
            alert('Lỗi: ' + err.message);
        }
    };

    // ── Per-user project permission modal ────────────────────────
    const handleOpenPermModal = async (userItem) => {
        setPermUser(userItem);
        try {
            const res = await adminApi.getUserProjectPermissions(userItem.id);
            setPermUserProjectIDs(res.data.project_ids || []);
        } catch {
            setPermUserProjectIDs([]);
        }
        setShowPermModal(true);
    };

    const handleToggleUserProject = async (projectId) => {
        const has = permUserProjectIDs.includes(projectId);
        try {
            if (has) {
                await adminApi.revokeProjectPermission(projectId, permUser.id);
                setPermUserProjectIDs(prev => prev.filter(id => id !== projectId));
            } else {
                await adminApi.grantProjectPermission(projectId, permUser.id);
                setPermUserProjectIDs(prev => [...prev, projectId]);
            }
        } catch (err) {
            alert('Lỗi: ' + err.message);
        }
    };

    // ── Project members tab ──────────────────────────────────────
    const handleProjectSelect = async (projectId) => {
        setSelectedProjectId(projectId);
        if (!projectId) { setProjectMembers([]); return; }
        setMembersLoading(true);
        try {
            const res = await adminApi.getProjectMembers(projectId);
            setProjectMembers(res.data || []);
        } catch {
            setProjectMembers([]);
        } finally {
            setMembersLoading(false);
        }
    };

    const handleGrantToProject = async (userId) => {
        try {
            await adminApi.grantProjectPermission(selectedProjectId, userId);
            const res = await adminApi.getProjectMembers(selectedProjectId);
            setProjectMembers(res.data || []);
        } catch (err) {
            alert('Lỗi: ' + err.message);
        }
    };

    const handleRevokeFromProject = async (userId) => {
        try {
            await adminApi.revokeProjectPermission(selectedProjectId, userId);
            setProjectMembers(prev => prev.filter(m => m.user_id !== userId));
        } catch (err) {
            alert('Lỗi: ' + err.message);
        }
    };

    if (loading) {
        return (
            <div className="loading-screen">
                <div className="loading-spinner"></div>
                <p>Đang tải...</p>
            </div>
        );
    }

    const selectedProject = allProjects.find(p => p.id === Number(selectedProjectId));
    const memberUserIds = new Set(projectMembers.map(m => m.user_id));
    const nonMemberUsers = users.filter(u => !memberUserIds.has(u.id) && u.role !== 'admin');

    return (
        <div className="admin-page">
            <div className="admin-header">
                <h1>Quản trị hệ thống</h1>
                <Link to="/" className="back-link">← Quay lại Dashboard</Link>
            </div>

            {error && <div className="auth-error">{error}</div>}

            {/* Tabs */}
            <div className="admin-tabs">
                <button
                    className={`admin-tab ${activeTab === 'users' ? 'active' : ''}`}
                    onClick={() => setActiveTab('users')}
                >
                    Người dùng ({users.length})
                </button>
                <button
                    className={`admin-tab ${activeTab === 'permissions' ? 'active' : ''}`}
                    onClick={() => setActiveTab('permissions')}
                >
                    Phân quyền dự án
                </button>
            </div>

            {/* Tab: Người dùng */}
            {activeTab === 'users' && (
                <div className="admin-section">
                    <h2>Danh sách người dùng</h2>
                    <table className="users-table">
                        <thead>
                            <tr>
                                <th>ID</th>
                                <th>Tên</th>
                                <th>Email</th>
                                <th>Vai trò</th>
                                <th>Số dự án</th>
                                <th>Ngày tạo</th>
                                <th>Hành động</th>
                            </tr>
                        </thead>
                        <tbody>
                            {users.map(userItem => (
                                <tr key={userItem.id}>
                                    <td>{userItem.id}</td>
                                    <td>{userItem.name}</td>
                                    <td>{userItem.email}</td>
                                    <td>
                                        <span className={`role-badge ${userItem.role}`}>
                                            {userItem.role === 'admin' ? 'Admin' : 'User'}
                                        </span>
                                    </td>
                                    <td>{userItem.projectCount || 0}</td>
                                    <td>{new Date(userItem.createdAt).toLocaleDateString('vi-VN')}</td>
                                    <td>
                                        <button
                                            className="action-btn role"
                                            onClick={() => handleRoleChange(userItem.id, userItem.role)}
                                            disabled={userItem.id === user?.id}
                                        >
                                            {userItem.role === 'admin' ? 'Hạ cấp' : 'Nâng cấp'}
                                        </button>
                                        {userItem.role !== 'admin' && (
                                            <button
                                                className="action-btn perm"
                                                onClick={() => handleOpenPermModal(userItem)}
                                            >
                                                Phân quyền
                                            </button>
                                        )}
                                        <button
                                            className="action-btn delete"
                                            onClick={() => handleDeleteClick(userItem)}
                                            disabled={userItem.id === user?.id}
                                        >
                                            Xóa
                                        </button>
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            )}

            {/* Tab: Phân quyền dự án */}
            {activeTab === 'permissions' && (
                <div className="admin-section">
                    <h2>Phân quyền theo dự án</h2>

                    {allProjects.length === 0 ? (
                        <p className="members-empty">Chưa có dự án nào. Hãy tạo dự án ở Dashboard trước.</p>
                    ) : (
                        <div className="project-select-row">
                            <label>Chọn dự án:</label>
                            <select
                                className="project-select"
                                value={selectedProjectId}
                                onChange={e => handleProjectSelect(e.target.value)}
                            >
                                <option value="">-- Chọn dự án --</option>
                                {allProjects.map(p => (
                                    <option key={p.id} value={p.id}>{p.name}</option>
                                ))}
                            </select>
                        </div>
                    )}

                    {selectedProjectId && (
                        <>
                            <div className="members-section">
                                <h3>
                                    Admin của dự án "{selectedProject?.name}"
                                    <span className="members-count">({projectMembers.length})</span>
                                </h3>
                                {membersLoading ? (
                                    <p className="members-loading">Đang tải...</p>
                                ) : projectMembers.length === 0 ? (
                                    <p className="members-empty">Chưa có admin nào được phân quyền</p>
                                ) : (
                                    <table className="users-table">
                                        <thead>
                                            <tr>
                                                <th>Tên</th>
                                                <th>Email</th>
                                                <th>Ngày cấp quyền</th>
                                                <th>Hành động</th>
                                            </tr>
                                        </thead>
                                        <tbody>
                                            {projectMembers.map(m => (
                                                <tr key={m.user_id}>
                                                    <td>{m.name}</td>
                                                    <td>{m.email}</td>
                                                    <td>{new Date(m.granted_at).toLocaleDateString('vi-VN')}</td>
                                                    <td>
                                                        <button
                                                            className="action-btn delete"
                                                            onClick={() => handleRevokeFromProject(m.user_id)}
                                                        >
                                                            Thu quyền
                                                        </button>
                                                    </td>
                                                </tr>
                                            ))}
                                        </tbody>
                                    </table>
                                )}
                            </div>

                            {nonMemberUsers.length > 0 && (
                                <div className="members-section">
                                    <h3>Cấp quyền admin cho user</h3>
                                    <table className="users-table">
                                        <thead>
                                            <tr>
                                                <th>Tên</th>
                                                <th>Email</th>
                                                <th>Hành động</th>
                                            </tr>
                                        </thead>
                                        <tbody>
                                            {nonMemberUsers.map(u => (
                                                <tr key={u.id}>
                                                    <td>{u.name}</td>
                                                    <td>{u.email}</td>
                                                    <td>
                                                        <button
                                                            className="action-btn role"
                                                            onClick={() => handleGrantToProject(u.id)}
                                                        >
                                                            Cấp quyền
                                                        </button>
                                                    </td>
                                                </tr>
                                            ))}
                                        </tbody>
                                    </table>
                                </div>
                            )}
                        </>
                    )}
                </div>
            )}

            {/* Modal xóa user */}
            <Modal
                isOpen={showDeleteModal}
                onClose={() => setShowDeleteModal(false)}
                title="Xác nhận xóa"
                size="small"
            >
                <ConfirmDialog
                    message="Bạn có chắc chắn muốn xóa người dùng này? Tất cả dự án và camera của họ cũng sẽ bị xóa."
                    itemName={userToDelete?.name}
                    onConfirm={handleConfirmDelete}
                    onCancel={() => setShowDeleteModal(false)}
                />
            </Modal>

            {/* Modal phân quyền dự án cho user */}
            <Modal
                isOpen={showPermModal}
                onClose={() => setShowPermModal(false)}
                title={`Phân quyền dự án — ${permUser?.name}`}
                size="medium"
            >
                <div className="perm-modal-body">
                    <p className="perm-modal-desc">
                        Chọn các dự án mà <strong>{permUser?.name}</strong> được phép quản lý (thêm/sửa/xóa camera, upload video):
                    </p>
                    {allProjects.length === 0 ? (
                        <p className="members-empty">Không có dự án nào</p>
                    ) : (
                        <ul className="perm-project-list">
                            {allProjects.map(p => {
                                const has = permUserProjectIDs.includes(p.id);
                                return (
                                    <li key={p.id} className={`perm-project-item ${has ? 'granted' : ''}`}>
                                        <span className="perm-project-name">{p.name}</span>
                                        <button
                                            className={`action-btn ${has ? 'delete' : 'role'}`}
                                            onClick={() => handleToggleUserProject(p.id)}
                                        >
                                            {has ? 'Thu quyền' : 'Cấp quyền'}
                                        </button>
                                    </li>
                                );
                            })}
                        </ul>
                    )}
                </div>
            </Modal>
        </div>
    );
}

export default AdminPage;
