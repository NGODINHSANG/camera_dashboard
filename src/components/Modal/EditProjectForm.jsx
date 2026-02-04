import { useState, useEffect } from 'react'

function EditProjectForm({ project, onSubmit, onCancel }) {
    const [name, setName] = useState(project?.name || '')

    useEffect(() => {
        setName(project?.name || '')
    }, [project])

    const handleSubmit = (e) => {
        e.preventDefault()

        if (!name.trim()) {
            alert('Vui lòng nhập tên dự án')
            return
        }

        onSubmit({ name: name.trim() })
    }

    return (
        <form onSubmit={handleSubmit}>
            <div className="form-group">
                <label className="form-label">
                    Tên Dự Án <span style={{ color: 'var(--status-offline)' }}>*</span>
                </label>
                <input
                    type="text"
                    className="form-input"
                    placeholder="VD: Dự án ABC"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    autoFocus
                />
            </div>

            <div className="modal-actions">
                <button type="button" className="btn btn-secondary" onClick={onCancel}>
                    Hủy
                </button>
                <button type="submit" className="btn btn-primary">
                    <svg viewBox="0 0 24 24" fill="currentColor">
                        <path d="M17 3H5c-1.11 0-2 .9-2 2v14c0 1.1.89 2 2 2h14c1.1 0 2-.9 2-2V7l-4-4zm-5 16c-1.66 0-3-1.34-3-3s1.34-3 3-3 3 1.34 3 3-1.34 3-3 3zm3-10H5V5h10v4z" />
                    </svg>
                    Lưu thay đổi
                </button>
            </div>
        </form>
    )
}

export default EditProjectForm
