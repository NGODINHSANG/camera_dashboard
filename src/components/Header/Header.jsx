import { useState, useEffect, useRef } from 'react'
import { Link } from 'react-router-dom'
import './Header.css'
import batgroupLogo from '../../assets/batgroup.jpg'

function Header({ title, onlineCount, offlineCount, showBackButton, onBackClick, user, onLogout, isAdmin, onAddCamera, hasProject }) {
    const [currentTime, setCurrentTime] = useState(new Date())

    useEffect(() => {
        const timer = setInterval(() => {
            setCurrentTime(new Date())
        }, 1000)
        return () => clearInterval(timer)
    }, [])

    const formatTime = (date) => {
        return date.toLocaleTimeString('vi-VN', {
            hour: '2-digit',
            minute: '2-digit',
            second: '2-digit',
            hour12: false
        })
    }

    const formatDate = (date) => {
        return date.toLocaleDateString('vi-VN', {
            day: '2-digit',
            month: '2-digit',
            year: 'numeric'
        }).replace(/\//g, '/')
    }

    return (
        <header className="header">
            <div className="header-left">
                {/* Back Button - hiện khi đang xem fullscreen */}
                {showBackButton && (
                    <button className="back-btn" onClick={onBackClick} title="Quay lại">
                        <svg viewBox="0 0 24 24" fill="currentColor">
                            <path d="M20 11H7.83l5.59-5.59L12 4l-8 8 8 8 1.41-1.41L7.83 13H20v-2z" />
                        </svg>
                        <span>Quay lại</span>
                    </button>
                )}
                <div className="logo">
                    <img src={batgroupLogo} alt="Batgroup Logo" className="logo-icon" />
                </div>
                <h1 className="header-title">{title}</h1>
            </div>

            <div className="header-center">
                <div className="status-group">
                    <div className="status-item status-online">
                        <span className="status-dot online"></span>
                        <span className="status-label">Online:</span>
                        <span className="status-value">{onlineCount}</span>
                    </div>
                    <div className="status-item status-offline">
                        <span className="status-dot offline"></span>
                        <span className="status-label">Offline:</span>
                        <span className="status-value">{offlineCount}</span>
                    </div>
                </div>
                <div className="system-info">
                    <span className="info-item network">
                        <svg viewBox="0 0 24 24" fill="currentColor" className="info-icon">
                            <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-1 17.93c-3.95-.49-7-3.85-7-7.93 0-.62.08-1.21.21-1.79L9 15v1c0 1.1.9 2 2 2v1.93zm6.9-2.54c-.26-.81-1-1.39-1.9-1.39h-1v-3c0-.55-.45-1-1-1H8v-2h2c.55 0 1-.45 1-1V7h2c1.1 0 2-.9 2-2v-.41c2.93 1.19 5 4.06 5 7.41 0 2.08-.8 3.97-2.1 5.39z" />
                        </svg>
                        Network
                    </span>
                    <span className="info-item system">
                        <svg viewBox="0 0 24 24" fill="currentColor" className="info-icon">
                            <path d="M20 18c1.1 0 1.99-.9 1.99-2L22 6c0-1.1-.9-2-2-2H4c-1.1 0-2 .9-2 2v10c0 1.1.9 2 2 2H0v2h24v-2h-4zM4 6h16v10H4V6z" />
                        </svg>
                        System
                    </span>
                </div>
                {/* Nút Thêm Camera - chỉ hiện khi không ở fullscreen */}
                {hasProject && isAdmin && !showBackButton && (
                    <button
                        className="add-camera-btn"
                        onClick={onAddCamera}
                        title="Thêm camera mới"
                    >
                        <svg viewBox="0 0 24 24" fill="currentColor">
                            <path d="M19 13h-6v6h-2v-6H5v-2h6V5h2v6h6v2z" />
                        </svg>
                        <span>Thêm Camera</span>
                    </button>
                )}
            </div>

            <div className="header-right">
                <div className="datetime">
                    <span className="time">{formatTime(currentTime)}</span>
                    <span className="date">{formatDate(currentTime)}</span>
                </div>
                {user && (
                    <UserMenu
                        user={user}
                        isAdmin={isAdmin}
                        onLogout={onLogout}
                    />
                )}
            </div>
        </header>
    )
}

function UserMenu({ user, isAdmin, onLogout }) {
    const [isOpen, setIsOpen] = useState(false)
    const menuRef = useRef(null)

    // Đóng menu khi click bên ngoài
    useEffect(() => {
        const handleClickOutside = (event) => {
            if (menuRef.current && !menuRef.current.contains(event.target)) {
                setIsOpen(false)
            }
        }

        if (isOpen) {
            document.addEventListener('mousedown', handleClickOutside)
        }
        return () => {
            document.removeEventListener('mousedown', handleClickOutside)
        }
    }, [isOpen])

    return (
        <div className="user-menu" ref={menuRef}>
            <button
                className={`user-menu-trigger ${isOpen ? 'active' : ''}`}
                onClick={() => setIsOpen(!isOpen)}
            >
                <div className="user-avatar">
                    {user.name?.charAt(0).toUpperCase() || 'U'}
                </div>
                <span className="user-name">{user.name}</span>
                <svg className="dropdown-arrow" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M7 10l5 5 5-5z" />
                </svg>
            </button>

            {isOpen && (
                <div className="user-menu-dropdown">
                    <div className="menu-header">
                        <div className="user-avatar large">
                            {user.name?.charAt(0).toUpperCase() || 'U'}
                        </div>
                        <div className="user-info">
                            <span className="user-fullname">{user.name}</span>
                            <span className="user-email">{user.email || 'user@example.com'}</span>
                        </div>
                    </div>
                    <div className="menu-divider"></div>
                    <div className="menu-items">
                        <Link to="/profile" className="menu-item" onClick={() => setIsOpen(false)}>
                            <svg viewBox="0 0 24 24" fill="currentColor">
                                <path d="M12 12c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4zm0 2c-2.67 0-8 1.34-8 4v2h16v-2c0-2.66-5.33-4-8-4z" />
                            </svg>
                            <span>Trang cá nhân</span>
                        </Link>
                        <Link to="/settings" className="menu-item" onClick={() => setIsOpen(false)}>
                            <svg viewBox="0 0 24 24" fill="currentColor">
                                <path d="M19.14 12.94c.04-.31.06-.63.06-.94 0-.31-.02-.63-.06-.94l2.03-1.58c.18-.14.23-.41.12-.61l-1.92-3.32c-.12-.22-.37-.29-.59-.22l-2.39.96c-.5-.38-1.03-.7-1.62-.94l-.36-2.54c-.04-.24-.24-.41-.48-.41h-3.84c-.24 0-.43.17-.47.41l-.36 2.54c-.59.24-1.13.57-1.62.94l-2.39-.96c-.22-.08-.47 0-.59.22L2.74 8.87c-.12.21-.08.47.12.61l2.03 1.58c-.04.31-.06.63-.06.94s.02.63.06.94l-2.03 1.58c-.18.14-.23.41-.12.61l1.92 3.32c.12.22.37.29.59.22l2.39-.96c.5.38 1.03.7 1.62.94l.36 2.54c.05.24.24.41.48.41h3.84c.24 0 .44-.17.47-.41l.36-2.54c.59-.24 1.13-.56 1.62-.94l2.39.96c.22.08.47 0 .59-.22l1.92-3.32c.12-.22.07-.47-.12-.61l-2.01-1.58zM12 15.6c-1.98 0-3.6-1.62-3.6-3.6s1.62-3.6 3.6-3.6 3.6 1.62 3.6 3.6-1.62 3.6-3.6 3.6z" />
                            </svg>
                            <span>Cài đặt</span>
                        </Link>
                        <Link to="/support" className="menu-item" onClick={() => setIsOpen(false)}>
                            <svg viewBox="0 0 24 24" fill="currentColor">
                                <path d="M11 18h2v-2h-2v2zm1-16C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 18c-4.41 0-8-3.59-8-8s3.59-8 8-8 8 3.59 8 8-3.59 8-8 8zm0-14c-2.21 0-4 1.79-4 4h2c0-1.1.9-2 2-2s2 .9 2 2c0 2-3 1.75-3 5h2c0-2.25 3-2.5 3-5 0-2.21-1.79-4-4-4z" />
                            </svg>
                            <span>Hỗ trợ</span>
                        </Link>
                        {isAdmin && (
                            <Link to="/admin" className="menu-item" onClick={() => setIsOpen(false)}>
                                <svg viewBox="0 0 24 24" fill="currentColor">
                                    <path d="M12 1L3 5v6c0 5.55 3.84 10.74 9 12 5.16-1.26 9-6.45 9-12V5l-9-4zm0 10.99h7c-.53 4.12-3.28 7.79-7 8.94V12H5V6.3l7-3.11v8.8z" />
                                </svg>
                                <span>Quản trị</span>
                            </Link>
                        )}
                    </div>
                    <div className="menu-divider"></div>
                    <button className="menu-item logout" onClick={onLogout}>
                        <svg viewBox="0 0 24 24" fill="currentColor">
                            <path d="M17 7l-1.41 1.41L18.17 11H8v2h10.17l-2.58 2.58L17 17l5-5zM4 5h8V3H4c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h8v-2H4V5z" />
                        </svg>
                        <span>Đăng xuất</span>
                    </button>
                </div>
            )}
        </div>
    )
}

export default Header
