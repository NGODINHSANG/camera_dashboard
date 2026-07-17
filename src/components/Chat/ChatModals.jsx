import { useState, useEffect } from 'react'
import { chatApi } from '../../api/chat'

// Modal tạo nhóm
export function CreateGroupModal({ onClose, onCreated, currentUserId }) {
    const [name, setName] = useState('')
    const [search, setSearch] = useState('')
    const [results, setResults] = useState([])
    const [selected, setSelected] = useState([])
    const [loading, setLoading] = useState(false)

    useEffect(() => {
        if (!search.trim()) { setResults([]); return }
        const t = setTimeout(async () => {
            const users = await chatApi.searchUsers(search).catch(() => [])
            setResults(users.filter(u => u.id !== currentUserId))
        }, 300)
        return () => clearTimeout(t)
    }, [search, currentUserId])

    const toggleUser = (u) => {
        setSelected(prev =>
            prev.find(s => s.id === u.id) ? prev.filter(s => s.id !== u.id) : [...prev, u]
        )
    }

    const handleCreate = async () => {
        if (!name.trim() || selected.length === 0) return
        setLoading(true)
        try {
            const conv = await chatApi.createConversation({
                type: 'group',
                name: name.trim(),
                userIds: selected.map(u => u.id),
            })
            onCreated(conv)
            onClose()
        } catch (_) {}
        setLoading(false)
    }

    return (
        <div className="modal-overlay" onClick={onClose}>
            <div className="chat-modal" onClick={e => e.stopPropagation()}>
                <div className="chat-modal-header">
                    <h3>Tạo nhóm chat</h3>
                    <button onClick={onClose}>✕</button>
                </div>
                <div className="chat-modal-body">
                    <input
                        className="chat-modal-input"
                        placeholder="Tên nhóm..."
                        value={name}
                        onChange={e => setName(e.target.value)}
                        autoFocus
                    />
                    <input
                        className="chat-modal-input"
                        placeholder="Tìm thành viên..."
                        value={search}
                        onChange={e => setSearch(e.target.value)}
                    />
                    {selected.length > 0 && (
                        <div className="selected-members">
                            {selected.map(u => (
                                <span key={u.id} className="selected-tag">
                                    {u.name}
                                    <button onClick={() => toggleUser(u)}>✕</button>
                                </span>
                            ))}
                        </div>
                    )}
                    <div className="user-search-results">
                        {results.map(u => (
                            <button
                                key={u.id}
                                className={`user-result-item ${selected.find(s => s.id === u.id) ? 'selected' : ''}`}
                                onClick={() => toggleUser(u)}
                            >
                                <div className="user-result-avatar">{u.name[0]?.toUpperCase()}</div>
                                <div>
                                    <div className="user-result-name">{u.name}</div>
                                    <div className="user-result-email">{u.email}</div>
                                </div>
                                {selected.find(s => s.id === u.id) && <span className="check-mark">✓</span>}
                            </button>
                        ))}
                    </div>
                </div>
                <div className="chat-modal-footer">
                    <button className="btn btn-secondary" onClick={onClose}>Hủy</button>
                    <button
                        className="btn btn-primary"
                        onClick={handleCreate}
                        disabled={!name.trim() || selected.length === 0 || loading}
                    >
                        {loading ? 'Đang tạo...' : `Tạo nhóm (${selected.length} thành viên)`}
                    </button>
                </div>
            </div>
        </div>
    )
}

// Modal đổi tên nhóm
export function RenameGroupModal({ conversation, onClose, onRenamed }) {
    const [name, setName] = useState(conversation?.name || '')
    const [loading, setLoading] = useState(false)

    const handleRename = async () => {
        if (!name.trim()) return
        setLoading(true)
        try {
            await chatApi.renameGroup(conversation.id, name.trim())
            onRenamed(name.trim())
            onClose()
        } catch (_) {}
        setLoading(false)
    }

    return (
        <div className="modal-overlay" onClick={onClose}>
            <div className="chat-modal small" onClick={e => e.stopPropagation()}>
                <div className="chat-modal-header">
                    <h3>Đổi tên nhóm</h3>
                    <button onClick={onClose}>✕</button>
                </div>
                <div className="chat-modal-body">
                    <input
                        className="chat-modal-input"
                        value={name}
                        onChange={e => setName(e.target.value)}
                        autoFocus
                        onKeyDown={e => e.key === 'Enter' && handleRename()}
                    />
                </div>
                <div className="chat-modal-footer">
                    <button className="btn btn-secondary" onClick={onClose}>Hủy</button>
                    <button className="btn btn-primary" onClick={handleRename} disabled={!name.trim() || loading}>
                        {loading ? 'Đang lưu...' : 'Lưu'}
                    </button>
                </div>
            </div>
        </div>
    )
}

// Modal quản lý thành viên nhóm
export function ManageMembersModal({ conversation, currentUserId, onClose, onUpdated }) {
    const [search, setSearch] = useState('')
    const [results, setResults] = useState([])
    const [loading, setLoading] = useState(false)
    const members = conversation?.members || []
    const isCreator = conversation?.createdBy === currentUserId

    useEffect(() => {
        if (!search.trim()) { setResults([]); return }
        const t = setTimeout(async () => {
            const users = await chatApi.searchUsers(search).catch(() => [])
            setResults(users.filter(u => !members.find(m => m.userId === u.id)))
        }, 300)
        return () => clearTimeout(t)
    }, [search, members])

    const addMember = async (userId) => {
        setLoading(true)
        try {
            await chatApi.addMember(conversation.id, userId)
            onUpdated()
        } catch (_) {}
        setLoading(false)
    }

    const removeMember = async (userId) => {
        setLoading(true)
        try {
            await chatApi.removeMember(conversation.id, userId)
            onUpdated()
        } catch (_) {}
        setLoading(false)
    }

    return (
        <div className="modal-overlay" onClick={onClose}>
            <div className="chat-modal" onClick={e => e.stopPropagation()}>
                <div className="chat-modal-header">
                    <h3>Quản lý thành viên</h3>
                    <button onClick={onClose}>✕</button>
                </div>
                <div className="chat-modal-body">
                    <div className="members-list">
                        <p className="members-label">Thành viên hiện tại ({members.length})</p>
                        {members.map(m => (
                            <div key={m.userId} className="member-item">
                                <div className="user-result-avatar">{m.name?.[0]?.toUpperCase()}</div>
                                <div className="member-info">
                                    <div className="user-result-name">{m.name} {m.userId === currentUserId && '(Bạn)'}</div>
                                    <div className="user-result-email">{m.email}</div>
                                </div>
                                {isCreator && m.userId !== currentUserId && (
                                    <button className="remove-member-btn" onClick={() => removeMember(m.userId)} disabled={loading}>
                                        Xóa
                                    </button>
                                )}
                            </div>
                        ))}
                    </div>
                    <div className="add-member-section">
                        <p className="members-label">Thêm thành viên</p>
                        <input
                            className="chat-modal-input"
                            placeholder="Tìm người dùng..."
                            value={search}
                            onChange={e => setSearch(e.target.value)}
                        />
                        <div className="user-search-results">
                            {results.map(u => (
                                <button key={u.id} className="user-result-item" onClick={() => addMember(u.id)} disabled={loading}>
                                    <div className="user-result-avatar">{u.name[0]?.toUpperCase()}</div>
                                    <div>
                                        <div className="user-result-name">{u.name}</div>
                                        <div className="user-result-email">{u.email}</div>
                                    </div>
                                    <span className="add-mark">+ Thêm</span>
                                </button>
                            ))}
                        </div>
                    </div>
                </div>
                <div className="chat-modal-footer">
                    <button className="btn btn-primary" onClick={onClose}>Xong</button>
                </div>
            </div>
        </div>
    )
}
