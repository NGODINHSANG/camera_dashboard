import { useState, useRef, useEffect } from 'react'
import './Footer.css'

function Footer({ isLive, gridLayout, onGridLayoutChange, aspectRatio, onAspectRatioChange }) {
    const [showAspectMenu, setShowAspectMenu] = useState(false)
    const menuRef = useRef(null)
    const btnRef = useRef(null)

    const gridOptions = [
        { value: '2x2', label: '2×2', cols: 2 },
        { value: '3x3', label: '3×3', cols: 3 },
        { value: '4x4', label: '4×4', cols: 4 },
    ]

    const aspectOptions = [
        { value: '16:9', label: '16:9' },
        { value: '4:3', label: '4:3' },
        { value: '1:1', label: '1:1' },
        { value: 'auto', label: 'Auto' },
    ]

    // Close menu when clicking outside
    useEffect(() => {
        const handleClickOutside = (e) => {
            if (menuRef.current && !menuRef.current.contains(e.target) &&
                btnRef.current && !btnRef.current.contains(e.target)) {
                setShowAspectMenu(false)
            }
        }
        if (showAspectMenu) {
            document.addEventListener('mousedown', handleClickOutside)
        }
        return () => document.removeEventListener('mousedown', handleClickOutside)
    }, [showAspectMenu])

    return (
        <footer className="footer">
            <div className="footer-left">
                <div className="live-indicator">
                    <span className="live-text">LIVE</span>
                    <span className={`live-dot ${isLive ? 'active' : ''}`}></span>
                </div>
            </div>

            <div className="footer-center"></div>

            <div className="footer-right">
                {/* Grid selector */}
                <div className="grid-selector-inline">
                    {gridOptions.map((option) => (
                        <button
                            key={option.value}
                            className={`grid-option-btn ${gridLayout === option.value ? 'active' : ''}`}
                            onClick={() => onGridLayoutChange(option.value)}
                            title={`Lưới ${option.label}`}
                        >
                            <div className={`grid-preview grid-preview-${option.cols}`}>
                                {Array.from({ length: option.cols * option.cols }).map((_, i) => (
                                    <div key={i} className="grid-cell"></div>
                                ))}
                            </div>
                            <span>{option.label}</span>
                        </button>
                    ))}
                </div>

                {/* Aspect ratio selector */}
                <div className="aspect-selector">
                    <button
                        ref={btnRef}
                        className={`aspect-btn ${showAspectMenu ? 'active' : ''}`}
                        onClick={() => setShowAspectMenu(!showAspectMenu)}
                        title="Tỷ lệ khung hình"
                    >
                        <svg viewBox="0 0 24 24" fill="currentColor">
                            <path d="M19 12h-2v3h-3v2h5v-5zM7 9h3V7H5v5h2V9zm14-6H3c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h18c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2zm0 16.01H3V4.99h18v14.02z" />
                        </svg>
                        <span>{aspectRatio}</span>
                    </button>

                    {showAspectMenu && (
                        <div className="aspect-menu" ref={menuRef}>
                            {aspectOptions.map((option) => (
                                <button
                                    key={option.value}
                                    className={`aspect-menu-item ${aspectRatio === option.value ? 'active' : ''}`}
                                    onClick={() => {
                                        onAspectRatioChange(option.value)
                                        setShowAspectMenu(false)
                                    }}
                                >
                                    {option.label}
                                </button>
                            ))}
                        </div>
                    )}
                </div>
            </div>
        </footer>
    )
}

export default Footer
