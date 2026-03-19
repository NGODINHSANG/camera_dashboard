import { useState } from 'react'

function AddCameraForm({ onSubmit, onCancel }) {
    const [formData, setFormData] = useState({
        name: '',
        location: '',
        streamUrl: '',
        isRecording: true,
        autoRecord: false,
        aspectRatio: 'auto',
    })

    const aspectOptions = [
        { value: 'auto', label: 'Tự động' },
        { value: '16:9', label: '16:9' },
        { value: '4:3', label: '4:3' },
        { value: '1:1', label: '1:1' },
    ]

    const handleChange = (e) => {
        const { name, value, type, checked } = e.target
        setFormData(prev => ({
            ...prev,
            [name]: type === 'checkbox' ? checked : value
        }))
    }

    const handleSubmit = (e) => {
        e.preventDefault()

        // Validate
        if (!formData.name.trim() || !formData.streamUrl.trim()) {
            alert('Vui lòng điền đầy đủ thông tin bắt buộc')
            return
        }

        onSubmit({
            ...formData,
            id: Date.now(), // Tạo ID unique
        })
    }

    return (
        <form onSubmit={handleSubmit}>
            <div className="form-group">
                <label className="form-label">
                    Tên Camera <span style={{ color: 'var(--status-offline)' }}>*</span>
                </label>
                <input
                    type="text"
                    name="name"
                    className="form-input"
                    placeholder="VD: Camera 1"
                    value={formData.name}
                    onChange={handleChange}
                    autoFocus
                />
            </div>

            <div className="form-group">
                <label className="form-label">Vị trí</label>
                <input
                    type="text"
                    name="location"
                    className="form-input"
                    placeholder="VD: Khu xưởng A"
                    value={formData.location}
                    onChange={handleChange}
                />
            </div>

            <div className="form-group">
                <label className="form-label">
                    URL Stream (HLS) <span style={{ color: 'var(--status-offline)' }}>*</span>
                </label>
                <input
                    type="text"
                    name="streamUrl"
                    className="form-input"
                    placeholder="http://example.com/stream/index.m3u8"
                    value={formData.streamUrl}
                    onChange={handleChange}
                />
            </div>

            <div className="form-group">
                <label className="form-label">Tỷ lệ khung hình</label>
                <div className="aspect-ratio-options">
                    {aspectOptions.map((option) => (
                        <label key={option.value} className="aspect-ratio-option">
                            <input
                                type="radio"
                                name="aspectRatio"
                                value={option.value}
                                checked={formData.aspectRatio === option.value}
                                onChange={handleChange}
                            />
                            <span className="aspect-ratio-label">{option.label}</span>
                        </label>
                    ))}
                </div>
            </div>

            <div className="form-group">
                <label className="checkbox-label">
                    <input
                        type="checkbox"
                        name="autoRecord"
                        checked={formData.autoRecord}
                        onChange={handleChange}
                    />
                    <span className="checkbox-text">Tự động ghi hình</span>
                </label>
                <p className="form-hint">Tự động ghi hình khi có tín hiệu stream (mỗi file 15 phút)</p>
            </div>

            <div className="modal-actions">
                <button type="button" className="btn btn-secondary" onClick={onCancel}>
                    Hủy
                </button>
                <button type="submit" className="btn btn-primary">
                    <svg viewBox="0 0 24 24" fill="currentColor">
                        <path d="M19 13h-6v6h-2v-6H5v-2h6V5h2v6h6v2z" />
                    </svg>
                    Thêm Camera
                </button>
            </div>
        </form>
    )
}

export default AddCameraForm
