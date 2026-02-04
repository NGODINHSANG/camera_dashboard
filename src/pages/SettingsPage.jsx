import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import './SettingsPage.css'

function SettingsPage() {
    const navigate = useNavigate()
    const [activeSection, setActiveSection] = useState('system')

    const sections = {
        system: { icon: '⚙️', label: 'Hệ thống' },
        camera: { icon: '📹', label: 'Camera' },
        storage: { icon: '💾', label: 'Lưu trữ' },
    }

    const systemSettings = [
        { id: 'display', icon: '🖥️', label: 'Màn hình', description: 'Cấu hình độ phân giải, tần số làm tươi' },
        { id: 'sound', icon: '🔊', label: 'Âm thanh', description: 'Điều chỉnh âm lượng, cảnh báo âm thanh' },
        { id: 'brightness', icon: '☀️', label: 'Độ sáng', description: 'Điều chỉnh độ sáng màn hình' },
        { id: 'theme', icon: '🎨', label: 'Giao diện', description: 'Tùy chỉnh theme, màu sắc' },
        { id: 'device', icon: '📱', label: 'Thiết bị điều khiển', description: 'Quản lý thiết bị kết nối' },
        { id: 'info', icon: 'ℹ️', label: 'Thông tin hệ thống', description: 'Phiên bản, thông số kỹ thuật' },
    ]

    const cameraSettings = [
        { id: 'quality', icon: '📊', label: 'Chất lượng video', description: 'Cấu hình độ phân giải, bitrate' },
        { id: 'record', icon: '⏺️', label: 'Cài đặt ghi hình', description: 'Lịch ghi, chế độ ghi hình' },
        { id: 'motion', icon: '🏃', label: 'Phát hiện chuyển động', description: 'Độ nhạy, vùng phát hiện' },
        { id: 'alert', icon: '🔔', label: 'Cảnh báo', description: 'Thông báo qua email, push notification' },
    ]

    const storageSettings = [
        { id: 'disk', icon: '💿', label: 'Quản lý ổ đĩa', description: 'Dung lượng, phân vùng' },
        { id: 'retention', icon: '📅', label: 'Thời gian lưu trữ', description: 'Cấu hình thời gian giữ video' },
        { id: 'backup', icon: '☁️', label: 'Sao lưu', description: 'Tự động sao lưu, cloud storage' },
        { id: 'cleanup', icon: '🗑️', label: 'Dọn dẹp tự động', description: 'Xóa video cũ tự động' },
    ]

    const getCurrentSettings = () => {
        if (activeSection === 'system') return systemSettings
        if (activeSection === 'camera') return cameraSettings
        if (activeSection === 'storage') return storageSettings
        return []
    }

    return (
        <div className="settings-page">
            <div className="settings-header">
                <button className="back-btn" onClick={() => navigate('/')}>
                    <svg viewBox="0 0 24 24" fill="currentColor">
                        <path d="M20 11H7.83l5.59-5.59L12 4l-8 8 8 8 1.41-1.41L7.83 13H20v-2z" />
                    </svg>
                    Quay lại
                </button>
                <h1 className="settings-title">⚙️ CÀI ĐẶT</h1>
            </div>

            <div className="settings-content">
                <aside className="settings-sidebar">
                    <div className="settings-nav">
                        {Object.entries(sections).map(([key, { icon, label }]) => (
                            <button
                                key={key}
                                className={`nav-item ${activeSection === key ? 'active' : ''}`}
                                onClick={() => setActiveSection(key)}
                            >
                                <span className="nav-icon">{icon}</span>
                                <span className="nav-label">{label}</span>
                            </button>
                        ))}
                    </div>
                </aside>

                <main className="settings-main">
                    <div className="settings-section-header">
                        <h2>
                            {sections[activeSection].icon} {sections[activeSection].label.toUpperCase()}
                        </h2>
                    </div>

                    <div className="settings-list">
                        {getCurrentSettings().map((setting) => (
                            <div key={setting.id} className="setting-item">
                                <div className="setting-icon">{setting.icon}</div>
                                <div className="setting-content">
                                    <div className="setting-label">{setting.label}</div>
                                    <div className="setting-description">{setting.description}</div>
                                </div>
                                <button className="setting-action">
                                    <svg viewBox="0 0 24 24" fill="currentColor">
                                        <path d="M8.59 16.59L13.17 12 8.59 7.41 10 6l6 6-6 6-1.41-1.41z" />
                                    </svg>
                                </button>
                            </div>
                        ))}
                    </div>
                </main>
            </div>
        </div>
    )
}

export default SettingsPage
