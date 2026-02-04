import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import './SupportPage.css'

function SupportPage() {
    const navigate = useNavigate()
    const [activeSection, setActiveSection] = useState('guide')

    const sections = {
        guide: { icon: 'ℹ️', label: 'Trung tâm hướng dẫn' },
        support: { icon: '🎧', label: 'Yêu cầu hỗ trợ kỹ thuật' },
        contact: { icon: '📞', label: 'Thông tin liên hệ hỗ trợ' },
    }

    const guideItems = [
        {
            id: 'basic',
            icon: '📖',
            title: 'Hướng dẫn sử dụng cơ bản',
            description: 'Hướng dẫn chi tiết cách sử dụng hệ thống từ A-Z',
            items: [
                'Đăng nhập và quản lý tài khoản',
                'Thêm và cấu hình camera',
                'Quản lý dự án',
                'Xem live stream và phát lại',
            ]
        },
        {
            id: 'common',
            icon: '⚡',
            title: 'Các thao tác phổ biến',
            description: 'Thao tác thường dùng hàng ngày',
            items: [
                'Thay đổi bố cục hiển thị camera',
                'Chụp ảnh và quay video',
                'Xuất dữ liệu ghi hình',
                'Cấu hình cảnh báo',
            ]
        },
        {
            id: 'faq',
            icon: '❓',
            title: 'Câu hỏi thường gặp (FAQ)',
            description: 'Giải đáp các thắc mắc phổ biến',
            items: [
                'Camera không kết nối được?',
                'Làm sao để xem lại video đã ghi?',
                'Cách thay đổi mật khẩu?',
                'Quản lý dung lượng lưu trữ',
            ]
        },
        {
            id: 'troubleshoot',
            icon: '🔧',
            title: 'Hướng dẫn xử lý lỗi thông dụng',
            description: 'Khắc phục các sự cố thường gặp',
            items: [
                'Khắc phục lỗi kết nối mạng',
                'Xử lý lỗi không xem được video',
                'Khắc phục lỗi đầy ổ cứng',
                'Reset hệ thống về mặc định',
            ]
        },
    ]

    const supportTickets = [
        { id: 1, title: 'Camera 3 không hiển thị hình ảnh', status: 'pending', date: '2026-01-28' },
        { id: 2, title: 'Yêu cầu nâng cấp tài khoản', status: 'processing', date: '2026-01-27' },
        { id: 3, title: 'Lỗi đăng nhập trên mobile', status: 'resolved', date: '2026-01-25' },
    ]

    const contactInfo = [
        {
            type: 'Hotline',
            icon: '📞',
            value: '1900-xxxx',
            description: 'Hỗ trợ 24/7'
        },
        {
            type: 'Email',
            icon: '✉️',
            value: 'support@dashboard.com',
            description: 'Phản hồi trong 24h'
        },
        {
            type: 'Địa chỉ',
            icon: '📍',
            value: 'Tầng 5, Tòa nhà ABC, Hà Nội',
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
                            <h2 className="section-title">📚 TRUNG TÂM HƯỚNG DẪN</h2>
                            <div className="guide-list">
                                {guideItems.map((guide) => (
                                    <div key={guide.id} className="guide-card">
                                        <div className="guide-header">
                                            <span className="guide-icon">{guide.icon}</span>
                                            <div>
                                                <h3 className="guide-title">{guide.title}</h3>
                                                <p className="guide-description">{guide.description}</p>
                                            </div>
                                        </div>
                                        <ul className="guide-items">
                                            {guide.items.map((item, index) => (
                                                <li key={index} className="guide-item">
                                                    <svg viewBox="0 0 24 24" fill="currentColor">
                                                        <path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z" />
                                                    </svg>
                                                    <span>{item}</span>
                                                </li>
                                            ))}
                                        </ul>
                                    </div>
                                ))}
                            </div>
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
