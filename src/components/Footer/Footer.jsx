import './Footer.css'

function Footer({ isLive, gridLayout, onGridLayoutChange }) {
    const gridOptions = [
        { value: '2x2', label: '2×2', cols: 2 },
        { value: '3x3', label: '3×3', cols: 3 },
        { value: '4x4', label: '4×4', cols: 4 },
    ]

    return (
        <footer className="footer">
            <div className="footer-left">
                <div className="live-indicator">
                    <span className="live-text">LIVE</span>
                    <span className={`live-dot ${isLive ? 'active' : ''}`}></span>
                </div>
            </div>

            <div className="footer-center">
                {/* Empty center */}
            </div>

            {/* Grid selector directly visible at bottom right */}
            <div className="footer-right">
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
            </div>
        </footer>
    )
}

export default Footer
