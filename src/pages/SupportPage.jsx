import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import './SupportPage.css'

// ============================================================
// DỮ LIỆU TÀI LIỆU - Cập nhật đường dẫn file tại đây
// ============================================================

const PDF_DOCS = [
    { id: 1, title: 'Hướng dẫn sử dụng Dashboard', description: 'Hướng dẫn toàn bộ các tính năng của hệ thống Camera Dashboard', size: '1.7 MB', url: '/docs/huong-dan-dashboard.pdf' },
    { id: 2, title: 'Hướng dẫn Livestream trên thiết bị Drone', description: 'Cách cấu hình và phát livestream từ thiết bị Drone', size: '1.7 MB', url: '/docs/huong-dan-drone.pdf' },
    { id: 3, title: 'Hướng dẫn Livestream Action 4', description: 'Hướng dẫn sử dụng camera Action 4 để livestream', size: '961 KB', url: '/docs/huong-dan-action4.pdf' },
    { id: 4, title: 'Hướng dẫn Livestream bằng điện thoại', description: 'Phát livestream trực tiếp từ điện thoại di động', size: '753 KB', url: '/docs/huong-dan-dien-thoai.pdf' },
]

const VIDEO_GUIDES = [
    // { id: 1, title: 'Video hướng dẫn cài đặt', description: 'Xem từng bước cài đặt hệ thống', duration: '12:35', url: '/docs/video-cai-dat.mp4', thumbnail: '' },
]

const DOC_FILES = [
    // { id: 1, title: 'Tài liệu kỹ thuật hệ thống', description: 'Thông số và kiến trúc kỹ thuật', size: '3.1 MB', url: '/docs/tai-lieu-ky-thuat.docx' },
]

const FAQ_ITEMS = [
    {
        id: 1,
        question: 'Camera không kết nối được, phải làm gì?',
        answer: 'Kiểm tra lại đường truyền mạng và URL stream. Đảm bảo MediaMTX đang chạy và camera đang phát tín hiệu RTSP. Kiểm tra log tại /api/health để xác nhận backend đang hoạt động.',
    },
    {
        id: 2,
        question: 'Làm sao để xem lại video đã ghi?',
        answer: 'Nhấn vào icon "Xem lại" trên từng camera trong dashboard. Chọn video từ danh sách và nhấn Play. Video hỗ trợ phát lại toàn bộ danh sách liên tiếp.',
    },
    {
        id: 3,
        question: 'Cách thay đổi mật khẩu tài khoản?',
        answer: 'Vào mục Hồ sơ cá nhân (icon user ở góc phải header), chọn "Đổi mật khẩu". Nhập mật khẩu hiện tại và mật khẩu mới, sau đó xác nhận.',
    },
    {
        id: 4,
        question: 'Quản lý dung lượng lưu trữ như thế nào?',
        answer: 'Video được lưu trên server tại thư mục recordings. Admin có thể xóa các video cũ qua giao diện Xem lại. Hệ thống hỗ trợ cấu hình tự động xóa video quá số ngày nhất định cho từng camera.',
    },
    {
        id: 5,
        question: 'Làm sao để thêm camera mới vào dự án?',
        answer: 'Chọn dự án trong sidebar, nhấn nút "+" ở góc trên để thêm camera. Điền tên, vị trí và URL stream HLS/RTSP của camera. Bật "Tự động ghi hình" nếu muốn ghi liên tục.',
    },
    {
        id: 6,
        question: 'Phân quyền người dùng hoạt động như thế nào?',
        answer: 'Admin toàn hệ thống có thể cấp quyền quản lý từng dự án cho người dùng cụ thể. Người dùng được phân quyền có thể thêm/sửa/xóa camera và upload video trong dự án đó.',
    },
]

// ============================================================

const GUIDE_CARDS = [
    {
        key: 'pdf',
        icon: '📄',
        label: 'TÀI LIỆU HƯỚNG DẪN PDF',
        description: 'Hướng dẫn sử dụng và tài liệu kỹ thuật dạng PDF',
        color: '#e53935',
        count: PDF_DOCS.length,
        unit: 'tài liệu',
    },
    {
        key: 'video',
        icon: '🎬',
        label: 'VIDEO HƯỚNG DẪN',
        description: 'Video minh hoạ từng bước cài đặt và vận hành',
        color: '#1e88e5',
        count: VIDEO_GUIDES.length,
        unit: 'video',
    },
    {
        key: 'doc',
        icon: '📝',
        label: 'TÀI LIỆU DOC',
        description: 'Văn bản quy trình, quy định và tài liệu nội bộ',
        color: '#2e7d32',
        count: DOC_FILES.length,
        unit: 'tài liệu',
    },
    {
        key: 'faq',
        icon: '❓',
        label: 'CÂU HỎI THƯỜNG GẶP (FAQ)',
        description: 'Giải đáp nhanh các thắc mắc phổ biến nhất',
        color: '#f57c00',
        count: FAQ_ITEMS.length,
        unit: 'câu hỏi',
    },
]

function PDFContent({ onBack }) {
    const [viewing, setViewing] = useState(null)
    return (
        <div className="guide-sub-section">
            <button className="guide-back-btn" onClick={onBack}>← Quay lại</button>
            <h3 className="guide-sub-title">📄 TÀI LIỆU HƯỚNG DẪN PDF</h3>
            {viewing && (
                <div className="pdf-viewer-overlay" onClick={() => setViewing(null)}>
                    <div className="pdf-viewer-container" onClick={e => e.stopPropagation()}>
                        <div className="pdf-viewer-header">
                            <span>{viewing.title}</span>
                            <button onClick={() => setViewing(null)}>✕</button>
                        </div>
                        <iframe src={viewing.url} title={viewing.title} className="pdf-iframe" />
                    </div>
                </div>
            )}
            {PDF_DOCS.length === 0 ? (
                <div className="guide-empty">Chưa có tài liệu PDF nào. Tài liệu sẽ được cập nhật sớm.</div>
            ) : (
                <div className="doc-list">
                    {PDF_DOCS.map(doc => (
                        <div key={doc.id} className="doc-item">
                            <span className="doc-item-icon">📄</span>
                            <div className="doc-item-info">
                                <strong>{doc.title}</strong>
                                <span>{doc.description} • {doc.size}</span>
                            </div>
                            <div className="doc-item-actions">
                                <button className="doc-action-btn" onClick={() => setViewing(doc)}>Xem</button>
                                <a className="doc-action-btn" href={doc.url} download>Tải về</a>
                            </div>
                        </div>
                    ))}
                </div>
            )}
        </div>
    )
}

function VideoContent({ onBack }) {
    const [playing, setPlaying] = useState(null)
    return (
        <div className="guide-sub-section">
            <button className="guide-back-btn" onClick={onBack}>← Quay lại</button>
            <h3 className="guide-sub-title">🎬 VIDEO HƯỚNG DẪN</h3>
            {playing && (
                <div className="pdf-viewer-overlay" onClick={() => setPlaying(null)}>
                    <div className="video-player-container" onClick={e => e.stopPropagation()}>
                        <div className="pdf-viewer-header">
                            <span>{playing.title}</span>
                            <button onClick={() => setPlaying(null)}>✕</button>
                        </div>
                        <video controls autoPlay className="video-player" src={playing.url}>
                            Trình duyệt không hỗ trợ video.
                        </video>
                    </div>
                </div>
            )}
            {VIDEO_GUIDES.length === 0 ? (
                <div className="guide-empty">Chưa có video hướng dẫn nào. Video sẽ được cập nhật sớm.</div>
            ) : (
                <div className="video-grid">
                    {VIDEO_GUIDES.map(video => (
                        <div key={video.id} className="video-card" onClick={() => setPlaying(video)}>
                            <div className="video-thumb">
                                {video.thumbnail
                                    ? <img src={video.thumbnail} alt={video.title} />
                                    : <div className="video-thumb-placeholder">▶</div>
                                }
                                <span className="video-duration">{video.duration}</span>
                            </div>
                            <div className="video-card-info">
                                <strong>{video.title}</strong>
                                <span>{video.description}</span>
                            </div>
                        </div>
                    ))}
                </div>
            )}
        </div>
    )
}

function DocContent({ onBack }) {
    return (
        <div className="guide-sub-section">
            <button className="guide-back-btn" onClick={onBack}>← Quay lại</button>
            <h3 className="guide-sub-title">📝 TÀI LIỆU DOC</h3>
            {DOC_FILES.length === 0 ? (
                <div className="guide-empty">Chưa có tài liệu DOC nào. Tài liệu sẽ được cập nhật sớm.</div>
            ) : (
                <div className="doc-list">
                    {DOC_FILES.map(doc => (
                        <div key={doc.id} className="doc-item">
                            <span className="doc-item-icon">📝</span>
                            <div className="doc-item-info">
                                <strong>{doc.title}</strong>
                                <span>{doc.description} • {doc.size}</span>
                            </div>
                            <div className="doc-item-actions">
                                <a className="doc-action-btn" href={doc.url} download>Tải về</a>
                            </div>
                        </div>
                    ))}
                </div>
            )}
        </div>
    )
}

function FAQContent({ onBack }) {
    const [openId, setOpenId] = useState(null)
    return (
        <div className="guide-sub-section">
            <button className="guide-back-btn" onClick={onBack}>← Quay lại</button>
            <h3 className="guide-sub-title">❓ CÂU HỎI THƯỜNG GẶP</h3>
            <div className="faq-list">
                {FAQ_ITEMS.map(item => (
                    <div key={item.id} className={`faq-item ${openId === item.id ? 'open' : ''}`}>
                        <button className="faq-question" onClick={() => setOpenId(openId === item.id ? null : item.id)}>
                            <span>{item.question}</span>
                            <span className="faq-chevron">{openId === item.id ? '▲' : '▼'}</span>
                        </button>
                        {openId === item.id && (
                            <div className="faq-answer"><p>{item.answer}</p></div>
                        )}
                    </div>
                ))}
            </div>
        </div>
    )
}

function SupportPage() {
    const navigate = useNavigate()
    const [activeSection, setActiveSection] = useState('guide')
    const [activeGuideCard, setActiveGuideCard] = useState(null)

    const sections = {
        guide: { icon: 'ℹ️', label: 'Trung tâm hướng dẫn' },
        support: { icon: '🎧', label: 'Yêu cầu hỗ trợ kỹ thuật' },
        contact: { icon: '📞', label: 'Thông tin liên hệ hỗ trợ' },
    }

    const supportTickets = [
        { id: 1, title: 'Camera 3 không hiển thị hình ảnh', status: 'pending', date: '2026-01-28' },
        { id: 2, title: 'Yêu cầu nâng cấp tài khoản', status: 'processing', date: '2026-01-27' },
        { id: 3, title: 'Lỗi đăng nhập trên mobile', status: 'resolved', date: '2026-01-25' },
    ]

    const contactInfo = [
        {
            type: 'Email',
            icon: '✉️',
            value: 'info@batgroup.vn',
            description: 'Phản hồi trong 24h'
        },
        {
            type: 'Địa chỉ',
            icon: '📍',
            value: 'Tầng 2 – Eco Lakeview, 32 Đại Từ, Phường Định Công',
            description: 'Văn phòng chính'
        },
        {
            type: 'Giờ làm việc',
            icon: '🕐',
            value: 'T2 - T6: 8:00 - 17:30',
            description: 'T7: 8:00 - 12:00'
        },
    ]

    return (
        <div className="support-page">
            <div className="support-header">
                <button className="back-btn" onClick={() => navigate('/')}>
                    <svg viewBox="0 0 24 24" fill="currentColor">
                        <path d="M20 11H7.83l5.59-5.59L12 4l-8 8 8 8 1.41-1.41L7.83 13H20v-2z" />
                    </svg>
                    Quay lại
                </button>
                <h1 className="support-title">ℹ️ HỖ TRỢ</h1>
            </div>

            <div className="support-content">
                <aside className="support-sidebar">
                    <div className="support-nav">
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

                <main className="support-main">
                    {activeSection === 'guide' && (
                        <>
                            {!activeGuideCard && (
                                <>
                                    <h2 className="section-title">📚 TRUNG TÂM HƯỚNG DẪN</h2>
                                    <div className="guide-cards-grid">
                                        {GUIDE_CARDS.map(card => (
                                            <button
                                                key={card.key}
                                                className="guide-main-card"
                                                style={{ '--card-color': card.color }}
                                                onClick={() => setActiveGuideCard(card.key)}
                                            >
                                                <span className="guide-main-card-icon">{card.icon}</span>
                                                <h3 className="guide-main-card-label">{card.label}</h3>
                                                <p className="guide-main-card-desc">{card.description}</p>
                                                <span className="guide-main-card-count">{card.count} {card.unit}</span>
                                            </button>
                                        ))}
                                    </div>
                                </>
                            )}
                            {activeGuideCard === 'pdf' && <PDFContent onBack={() => setActiveGuideCard(null)} />}
                            {activeGuideCard === 'video' && <VideoContent onBack={() => setActiveGuideCard(null)} />}
                            {activeGuideCard === 'doc' && <DocContent onBack={() => setActiveGuideCard(null)} />}
                            {activeGuideCard === 'faq' && <FAQContent onBack={() => setActiveGuideCard(null)} />}
                        </>
                    )}

                    {activeSection === 'support' && (
                        <>
                            <h2 className="section-title">🎧 YÊU CẦU HỖ TRỢ KỸ THUẬT</h2>
                            <button className="create-ticket-btn">
                                <svg viewBox="0 0 24 24" fill="currentColor">
                                    <path d="M19 13h-6v6h-2v-6H5v-2h6V5h2v6h6v2z" />
                                </svg>
                                Tạo yêu cầu mới
                            </button>

                            <div className="tickets-list">
                                <h3 className="tickets-header">Danh sách yêu cầu của bạn</h3>
                                {supportTickets.map((ticket) => (
                                    <div key={ticket.id} className="ticket-item">
                                        <div className="ticket-info">
                                            <div className="ticket-title">{ticket.title}</div>
                                            <div className="ticket-date">#{ticket.id} • {ticket.date}</div>
                                        </div>
                                        <span className={`ticket-status status-${ticket.status}`}>
                                            {ticket.status === 'pending' && 'Chờ xử lý'}
                                            {ticket.status === 'processing' && 'Đang xử lý'}
                                            {ticket.status === 'resolved' && 'Đã giải quyết'}
                                        </span>
                                    </div>
                                ))}
                            </div>
                        </>
                    )}

                    {activeSection === 'contact' && (
                        <>
                            <h2 className="section-title">📞 THÔNG TIN LIÊN HỆ HỖ TRỢ</h2>
                            <div className="contact-grid">
                                {contactInfo.map((contact, index) => (
                                    <div key={index} className="contact-card">
                                        <div className="contact-icon">{contact.icon}</div>
                                        <div className="contact-content">
                                            <div className="contact-type">{contact.type}</div>
                                            <div className="contact-value">{contact.value}</div>
                                            <div className="contact-description">{contact.description}</div>
                                        </div>
                                    </div>
                                ))}
                            </div>

                            <div className="contact-form-section">
                                <h3>Gửi tin nhắn cho chúng tôi</h3>
                                <form className="contact-form">
                                    <div className="form-row">
                                        <input type="text" placeholder="Họ và tên" className="form-input" />
                                        <input type="email" placeholder="Email" className="form-input" />
                                    </div>
                                    <input type="text" placeholder="Chủ đề" className="form-input" />
                                    <textarea placeholder="Nội dung tin nhắn..." className="form-textarea" rows="5"></textarea>
                                    <button type="submit" className="submit-btn">
                                        <svg viewBox="0 0 24 24" fill="currentColor">
                                            <path d="M2.01 21L23 12 2.01 3 2 10l15 2-15 2z" />
                                        </svg>
                                        Gửi tin nhắn
                                    </button>
                                </form>
                            </div>
                        </>
                    )}
                </main>
            </div>
        </div>
    )
}

export default SupportPage
