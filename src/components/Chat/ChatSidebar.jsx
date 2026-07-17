import { useState, useEffect } from 'react'
import { chatApi } from '../../api/chat'

function ConversationItem({ conv, isActive, onClick, currentUserId, onlineUserIds }) {
    const isGroup = conv.type === 'group'
    const displayName = isGroup ? (conv.name || 'Nhóm không tên') : (conv.otherUser?.name || '...')
    const lastMsg = conv.lastMessage
    const isOtherOnline = !isGroup && onlineUserIds.includes(conv.otherUser?.id)

    return (
        <button className={`conv-item ${isActive ? 'active' : ''}`} onClick={onClick}>
            <div className="conv-avatar">
                {isGroup ? '👥' : displayName[0]?.toUpperCase()}
                {isOtherOnline && <span className="online-dot" />}
            </div>
            <div className="conv-info">
                <div className="conv-header-row">
                    <span className="conv-name">{displayName}</span>
                    {conv.unreadCount > 0 && <span className="unread-badge">{conv.unreadCount}</span>}
                </div>
                {lastMsg && (
                    <span className="conv-last-msg">
                        {lastMsg.senderId === currentUserId ? 'Bạn: ' : ''}
                        {lastMsg.type === 'text' ? lastMsg.content : (lastMsg.type === 'image' ? '🖼 Hình ảnh' : '📎 File')}
                    </span>
                )}
            </div>
        </button>
    )
}

function ChatSidebar({ currentUserId, activeConvId, onSelectConv, onCreateDM, onCreateGroup, onlineUserIds, conversations, loading }) {
    const [search, setSearch] = useState('')
    const [searchResults, setSearchResults] = useState([])
    const [searching, setSearching] = useState(false)

    useEffect(() => {
        if (!search.trim()) { setSearchResults([]); return }
        const t = setTimeout(async () => {
            setSearching(true)
            try {
                const users = await chatApi.searchUsers(search.trim())
                setSearchResults(users || [])
            } catch (_) {}
            setSearching(false)
        }, 300)
        return () => clearTimeout(t)
    }, [search])

    return (
        <aside className="chat-sidebar">
            <div className="chat-sidebar-header">
                <h2>Tin nhắn</h2>
                <div className="chat-sidebar-actions">
                    <button className="chat-icon-btn" title="Nhắn tin cá nhân" onClick={onCreateDM}>
                        <svg viewBox="0 0 24 24" fill="currentColor"><path d="M20 4H4c-1.1 0-2 .9-2 2v18l4-4h14c1.1 0 2-.9 2-2V6c0-1.1-.9-2-2-2z"/></svg>
                    </button>
                    <button className="chat-icon-btn" title="Tạo nhóm chat" onClick={onCreateGroup}>
                        <svg viewBox="0 0 24 24" fill="currentColor"><path d="M16 11c1.66 0 2.99-1.34 2.99-3S17.66 5 16 5c-1.66 0-3 1.34-3 3s1.34 3 3 3zm-8 0c1.66 0 2.99-1.34 2.99-3S9.66 5 8 5C6.34 5 5 6.34 5 8s1.34 3 3 3zm0 2c-2.33 0-7 1.17-7 3.5V19h14v-2.5c0-2.33-4.67-3.5-7-3.5zm8 0c-.29 0-.62.02-.97.05 1.16.84 1.97 1.97 1.97 3.45V19h6v-2.5c0-2.33-4.67-3.5-7-3.5z"/></svg>
                    </button>
                </div>
            </div>

            <div className="chat-search-box">
                <svg viewBox="0 0 24 24" fill="currentColor"><path d="M15.5 14h-.79l-.28-.27A6.471 6.471 0 0 0 16 9.5 6.5 6.5 0 1 0 9.5 16c1.61 0 3.09-.59 4.23-1.57l.27.28v.79l5 4.99L20.49 19l-4.99-5zm-6 0C7.01 14 5 11.99 5 9.5S7.01 5 9.5 5 14 7.01 14 9.5 11.99 14 9.5 14z"/></svg>
                <input
                    placeholder="Tìm người dùng..."
                    value={search}
                    onChange={e => setSearch(e.target.value)}
                />
                {search && <button className="clear-search" onClick={() => setSearch('')}>✕</button>}
            </div>

            <div className="conv-list">
                {search.trim() ? (
                    <>
                        {searching && <div className="conv-loading">Đang tìm...</div>}
                        {!searching && searchResults.length === 0 && <div className="conv-empty">Không tìm thấy</div>}
                        {searchResults.map(u => (
                            <button key={u.id} className="conv-item search-result" onClick={() => { onCreateDM(u); setSearch('') }}>
                                <div className="conv-avatar">
                                    {u.name[0]?.toUpperCase()}
                                    {onlineUserIds.includes(u.id) && <span className="online-dot" />}
                                </div>
                                <div className="conv-info">
                                    <span className="conv-name">{u.name}</span>
                                    <span className="conv-last-msg">{u.email}</span>
                                </div>
                            </button>
                        ))}
                    </>
                ) : (
                    <>
                        {loading && <div className="conv-loading">Đang tải...</div>}
                        {!loading && conversations.length === 0 && (
                            <div className="conv-empty">Chưa có cuộc trò chuyện nào</div>
                        )}
                        {conversations.map(conv => (
                            <ConversationItem
                                key={conv.id}
                                conv={conv}
                                isActive={conv.id === activeConvId}
                                currentUserId={currentUserId}
                                onlineUserIds={onlineUserIds}
                                onClick={() => onSelectConv(conv)}
                            />
                        ))}
                    </>
                )}
            </div>
        </aside>
    )
}

export default ChatSidebar
