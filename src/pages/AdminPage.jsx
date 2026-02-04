import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { adminApi } from '../api/admin';
import Modal from '../components/Modal/Modal';
import ConfirmDialog from '../components/Modal/ConfirmDialog';
import './AuthPages.css';

function AdminPage() {
    const { user } = useAuth();
    const [users, setUsers] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');
    const [showDeleteModal, setShowDeleteModal] = useState(false);
    const [userToDelete, setUserToDelete] = useState(null);

    useEffect(() => {
        loadUsers();
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

    const handleRoleChange = async (userId, currentRole) => {
        const newRole = currentRole === 'admin' ? 'user' : 'admin';
        try {
            await adminApi.updateUserRole(userId, newRole);
            setUsers(users.map(u =>
                u.id === userId ? { ...u, role: newRole } : u
            ));
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

    if (loading) {
        return (
            <div className="loading-screen">
                <div className="loading-spinner"></div>
                <p>Đang tải...</p>
            </div>
        );
    }

    return (
        <div className="admin-page">
            <div className="admin-header">
                <h1>Quản trị hệ thống</h1>
                <Link to="/" className="back-link">← Quay lại Dashboard</Link>
            </div>

            {error && <div className="auth-error">{error}</div>}

            <div className="admin-section">
                <h2>Danh sách người dùng ({users.length})</h2>
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
        </div>
    );
}

export default AdminPage;
