import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'
import authApi from '../api/auth'
import './ProfilePage.css'

function ProfilePage() {
    const navigate = useNavigate()
    const { user, login } = useAuth()
    const [isEditing, setIsEditing] = useState(false)
    const [isChangingPassword, setIsChangingPassword] = useState(false)
    const [formData, setFormData] = useState({
        name: user?.name || '',
        email: user?.email || '',
    })
    const [passwordData, setPasswordData] = useState({
        currentPassword: '',
        newPassword: '',
        confirmPassword: '',
    })
    const [message, setMessage] = useState({ type: '', text: '' })

    useEffect(() => {
        if (user) {
            setFormData({
                name: user.name,
                email: user.email,
            })
        }
    }, [user])

    const handleUpdateProfile = async (e) => {
        e.preventDefault()
        setMessage({ type: '', text: '' })

        try {
            // TODO: Implement update profile API
            // await authApi.updateProfile(formData)

            // For now, just show success message
            setMessage({ type: 'success', text: 'Cập nhật thông tin thành công!' })
            setIsEditing(false)

            // Update user in context (simulated)
            // In real implementation, you should re-fetch user data
            // const updatedUser = await authApi.getMe()
            // login(updatedUser, localStorage.getItem('token'))
        } catch (error) {
            setMessage({ type: 'error', text: error.message || 'Có lỗi xảy ra' })
        }
    }

    const handleChangePassword = async (e) => {
        e.preventDefault()
        setMessage({ type: '', text: '' })

        if (passwordData.newPassword !== passwordData.confirmPassword) {
            setMessage({ type: 'error', text: 'Mật khẩu mới không khớp' })
            return
        }

        if (passwordData.newPassword.length < 6) {
            setMessage({ type: 'error', text: 'Mật khẩu phải có ít nhất 6 ký tự' })
            return
        }

        try {
            // TODO: Implement change password API
            // await authApi.changePassword(passwordData)

            setMessage({ type: 'success', text: 'Đổi mật khẩu thành công!' })
            setIsChangingPassword(false)
            setPasswordData({
                currentPassword: '',
                newPassword: '',
                confirmPassword: '',
            })
        } catch (error) {
            setMessage({ type: 'error', text: error.message || 'Có lỗi xảy ra' })
        }
    }

    const handleCancel = () => {
        setIsEditing(false)
        setIsChangingPassword(false)
        setFormData({
            name: user?.name || '',
            email: user?.email || '',
        })
        setPasswordData({
            currentPassword: '',
            newPassword: '',
            confirmPassword: '',
        })
        setMessage({ type: '', text: '' })
    }

    const getRoleBadgeClass = (role) => {
        return role === 'admin' ? 'role-badge admin' : 'role-badge user'
    }

    const getRoleLabel = (role) => {
        return role === 'admin' ? 'Quản trị viên' : 'Người dùng'
    }

    return (
        <div className="profile-page">
            <div className="profile-container">
                {/* Nút quay lại */}
                <button className="back-button" onClick={() => navigate('/')}>
                    <svg viewBox="0 0 24 24" fill="currentColor">
                        <path d="M20 11H7.83l5.59-5.59L12 4l-8 8 8 8 1.41-1.41L7.83 13H20v-2z" />
                    </svg>
                    Quay lại Dashboard
                </button>

                <div className="profile-header">
                    <div className="profile-avatar">
                        <svg viewBox="0 0 24 24" fill="currentColor">
                            <path d="M12 12c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4zm0 2c-2.67 0-8 1.34-8 4v2h16v-2c0-2.66-5.33-4-8-4z" />
                        </svg>
                    </div>
                    <h1>Trang cá nhân</h1>
                    <span className={getRoleBadgeClass(user?.role)}>
                        {getRoleLabel(user?.role)}
                    </span>
                </div>

                {message.text && (
                    <div className={`profile-message ${message.type}`}>
                        {message.text}
                    </div>
                )}

                <div className="profile-content">
                    {/* Thông tin cá nhân */}
                    <section className="profile-section">
                        <div className="section-header">
                            <h2>Thông tin cá nhân</h2>
                            {!isEditing && !isChangingPassword && (
                                <button
                                    className="btn-edit"
                                    onClick={() => setIsEditing(true)}
                                >
                                    <svg viewBox="0 0 24 24" fill="currentColor">
                                        <path d="M3 17.25V21h3.75L17.81 9.94l-3.75-3.75L3 17.25zM20.71 7.04c.39-.39.39-1.02 0-1.41l-2.34-2.34c-.39-.39-1.02-.39-1.41 0l-1.83 1.83 3.75 3.75 1.83-1.83z" />
                                    </svg>
                                    Chỉnh sửa
                                </button>
                            )}
                        </div>

                        {isEditing ? (
                            <form onSubmit={handleUpdateProfile} className="profile-form">
                                <div className="form-group">
                                    <label htmlFor="name">Họ tên</label>
                                    <input
                                        type="text"
                                        id="name"
                                        value={formData.name}
                                        onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                                        required
                                    />
                                </div>

                                <div className="form-group">
                                    <label htmlFor="email">Email</label>
                                    <input
                                        type="email"
                                        id="email"
                                        value={formData.email}
                                        onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                                        required
                                        disabled
                                        title="Email không thể thay đổi"
                                    />
                                    <small>Email không thể thay đổi</small>
                                </div>

                                <div className="form-actions">
                                    <button type="submit" className="btn-primary">
                                        Lưu thay đổi
                                    </button>
                                    <button type="button" className="btn-secondary" onClick={handleCancel}>
                                        Hủy
                                    </button>
                                </div>
                            </form>
                        ) : (
                            <div className="profile-info">
                                <div className="info-item">
                                    <span className="info-label">Họ tên:</span>
                                    <span className="info-value">{user?.name}</span>
                                </div>
                                <div className="info-item">
                                    <span className="info-label">Email:</span>
                                    <span className="info-value">{user?.email}</span>
                                </div>
                                <div className="info-item">
                                    <span className="info-label">Vai trò:</span>
                                    <span className="info-value">{getRoleLabel(user?.role)}</span>
                                </div>
                            </div>
                        )}
                    </section>

                    {/* Đổi mật khẩu */}
                    <section className="profile-section">
                        <div className="section-header">
                            <h2>Bảo mật</h2>
                            {!isChangingPassword && !isEditing && (
                                <button
                                    className="btn-edit"
                                    onClick={() => setIsChangingPassword(true)}
                                >
                                    <svg viewBox="0 0 24 24" fill="currentColor">
                                        <path d="M18 8h-1V6c0-2.76-2.24-5-5-5S7 3.24 7 6v2H6c-1.1 0-2 .9-2 2v10c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V10c0-1.1-.9-2-2-2zm-6 9c-1.1 0-2-.9-2-2s.9-2 2-2 2 .9 2 2-.9 2-2 2zm3.1-9H8.9V6c0-1.71 1.39-3.1 3.1-3.1 1.71 0 3.1 1.39 3.1 3.1v2z" />
                                    </svg>
                                    Đổi mật khẩu
                                </button>
                            )}
                        </div>

                        {isChangingPassword ? (
                            <form onSubmit={handleChangePassword} className="profile-form">
                                <div className="form-group">
                                    <label htmlFor="currentPassword">Mật khẩu hiện tại</label>
                                    <input
                                        type="password"
                                        id="currentPassword"
                                        value={passwordData.currentPassword}
                                        onChange={(e) => setPasswordData({ ...passwordData, currentPassword: e.target.value })}
                                        required
                                    />
                                </div>

                                <div className="form-group">
                                    <label htmlFor="newPassword">Mật khẩu mới</label>
                                    <input
                                        type="password"
                                        id="newPassword"
                                        value={passwordData.newPassword}
                                        onChange={(e) => setPasswordData({ ...passwordData, newPassword: e.target.value })}
                                        required
                                        minLength={6}
                                    />
                                </div>

                                <div className="form-group">
                                    <label htmlFor="confirmPassword">Xác nhận mật khẩu mới</label>
                                    <input
                                        type="password"
                                        id="confirmPassword"
                                        value={passwordData.confirmPassword}
                                        onChange={(e) => setPasswordData({ ...passwordData, confirmPassword: e.target.value })}
                                        required
                                        minLength={6}
                                    />
                                </div>

                                <div className="form-actions">
                                    <button type="submit" className="btn-primary">
                                        Đổi mật khẩu
                                    </button>
                                    <button type="button" className="btn-secondary" onClick={handleCancel}>
                                        Hủy
                                    </button>
                                </div>
                            </form>
                        ) : (
                            <div className="profile-info">
                                <div className="info-item">
                                    <span className="info-label">Mật khẩu:</span>
                                    <span className="info-value">••••••••</span>
                                </div>
                                <small className="security-note">
                                    Để bảo mật tài khoản, bạn nên đổi mật khẩu thường xuyên
                                </small>
                            </div>
                        )}
                    </section>
                </div>
            </div>
        </div>
    )
}

export default ProfilePage
