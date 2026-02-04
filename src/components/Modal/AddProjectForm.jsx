import { useState } from 'react'

function AddProjectForm({ onSubmit, onCancel }) {
    const [formData, setFormData] = useState({
        name: '',
    })

    const handleChange = (e) => {
        const { name, value } = e.target
        setFormData(prev => ({
            ...prev,
            [name]: value
        }))
    }

    const handleSubmit = (e) => {
        e.preventDefault()

        if (!formData.name.trim()) {
            alert('Vui lòng nhập tên dự án')
            return
        }

        onSubmit({
            id: Date.now(),
            name: formData.name.trim().toUpperCase(),
            cameras: []
        })
    }

    return (
        <form onSubmit={handleSubmit}>
            <div className="form-group">
                <label className="form-label">
                    Tên Dự Án <span style={{ color: 'var(--status-offline)' }}>*</span>
                </label>
                <input
                    type="text"
                    name="name"
                    className="form-input"
                    placeholder="VD: HUAJIN NAM ĐỊNH"
                    value={formData.name}
                    onChange={handleChange}
                    autoFocus
                />
            </div>

            <div className="modal-actions">
                <button type="button" className="btn btn-secondary" onClick={onCancel}>
                    Hủy
                </button>
                <button type="submit" className="btn btn-primary">
                    <svg viewBox="0 0 24 24" fill="currentColor">
                        <path d="M19 13h-6v6h-2v-6H5v-2h6V5h2v6h6v2z" />
                    </svg>
                    Thêm Dự Án
                </button>
            </div>
        </form>
    )
}

export default AddProjectForm
