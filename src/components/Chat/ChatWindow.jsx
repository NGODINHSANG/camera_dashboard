import { useState, useEffect, useRef } from 'react'
import { chatApi } from '../../api/chat'

const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080'

function MessageBubble({ msg, isOwn }) {
    const isImage = msg.type === 'image'
    const isFile = msg.type === 'file'
    const att = msg.attachments?.[0]

    return (
        <div className={`msg-row ${isOwn ? 'own' : ''}`}>
            {!isOwn && <div className="msg-avatar">{msg.senderName?.[0]?.toUpperCase()}</div>}
            <div className="msg-bubble-wrap">
                {!isOwn && <span className="msg-sender">{msg.senderName}</span>}
                <div className={`msg-bubble ${isOwn ? 'own' : ''}`}>
                    {isImage && att && (
                        <a href={`${API_BASE}${att.url}`} target="_blank" rel="noreferrer">
                            <img src={`${API_BASE}${att.url}`} alt={att.originalName} className="msg-image" />
                        </a>
                    )}
                    {isFile && att && (
                        <a href={`${API_BASE}${att.url}`} download={att.originalName} className="msg-file-link">
                            <svg viewBox="0 0 24 24" fill="currentColor"><path d="M19 9h-4V3H9v6H5l7 7 7-7zM5 18v2h14v-2H5z"/></svg>
                            <span>{att.originalName}</span>
                            <small>{(att.fileSize / 1024).toFixed(1)} KB</small>
                        </a>
                    )}
                    {msg.type === 'text' && <p>{msg.content}</p>}
                </div>
                <span className="msg-time">
                    {new Date(msg.createdAt).toLocaleTimeString('vi-VN', { hour: '2-digit', minute: '2-digit' })}
                </span>
            </div>
        </div>
    )
}

function ChatWindow({ conversation, currentUserId, onRenameGroup, onManageMembers }) {
    const [messages, setMessages] = useState([])
    const [input, setInput] = useState('')
    const [loading, setLoading] = useState(false)
    const [hasMore, setHasMore] = useState(false)
    const [uploading, setUploading] = useState(false)
    const [error, setError] = useState(null)
    const [showInfo, setShowInfo] = useState(false)
    const bottomRef = useRef(null)
    const fileInputRef = useRef(null)
    const convId = conversation?.id

    // Load lần đầu khi chọn conversation
    useEffect(() => {
        if (!convId) return
        let cancelled = false
        setMessages([])
        setHasMore(false)
        setLoading(true)
        setError(null)

        chatApi.getMessages(convId, 0).then(msgs => {
            if (cancelled) return
            const list = Array.isArray(msgs) ? msgs : []
            setMessages(list)
            setHasMore(list.length >= 30)
            setLoading(false)
        }).catch(err => {
            if (cancelled) return
            console.error('[ChatWindow] getMessages failed:', err)
            setError(err.message || 'Không thể tải tin nhắn')
            setLoading(false)
        })

        chatApi.markRead(convId).catch(() => {})
        return () => { cancelled = true }
    }, [convId])

    // Load thêm khi nhấn "Tải thêm"
    const loadMore = async () => {
        if (loading || !convId) return
        setLoading(true)
        try {
            const msgs = await chatApi.getMessages(convId, messages.length)
            const list = Array.isArray(msgs) ? msgs : []
            setMessages(prev => [...list, ...prev])
            setHasMore(list.length >= 30)
        } catch (_) {}
        setLoading(false)
    }

    useEffect(() => {
        bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
    }, [messages])

    // Nhận tin nhắn mới từ WebSocket (event đẩy lên từ MessagesPage)
    useEffect(() => {
        const handler = (e) => {
            const { detail } = e
            if (detail?.conversationId === convId) {
                setMessages(prev => {
                    if (prev.find(m => m.id === detail.id)) return prev
                    return [...prev, detail]
                })
            }
        }
        window.addEventListener('chat:new_message', handler)
        return () => window.removeEventListener('chat:new_message', handler)
    }, [convId])

    const sendMessage = async () => {
        if (!input.trim()) return
        const content = input.trim()
        setInput('')
        try {
            const msg = await chatApi.sendMessage(convId, content)
            if (msg?.id) {
                setMessages(prev => prev.find(m => m.id === msg.id) ? prev : [...prev, msg])
            }
        } catch (_) {}
    }

    const handleKeyDown = (e) => {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault()
            sendMessage()
        }
    }

    const handleFileSelect = async (e) => {
        const file = e.target.files?.[0]
        if (!file) return
        setUploading(true)
        try {
            const msg = await chatApi.uploadAttachment(convId, file)
            if (msg?.id) {
                setMessages(prev => prev.find(m => m.id === msg.id) ? prev : [...prev, msg])
            }
        } catch (_) {}
        setUploading(false)
        e.target.value = ''
    }

    if (!conversation) {
        return (
            <div className="chat-window chat-window-empty">
                <div className="chat-empty-state">
                    <svg viewBox="0 0 24 24" fill="currentColor"><path d="M20 4H4c-1.1 0-2 .9-2 2v18l4-4h14c1.1 0 2-.9 2-2V6c0-1.1-.9-2-2-2zm-2 12H6v-2h12v2zm0-3H6V11h12v2zm0-3H6V8h12v2z"/></svg>
                    <p>Chọn một cuộc trò chuyện để bắt đầu</p>
                </div>
            </div>
        )
    }

    const isGroup = conversation.type === 'group'
    const displayName = isGroup
        ? (conversation.name || 'Nhóm không tên')
        : conversation.otherUser?.name || '...'

    return (
        <div className="chat-window">
            {/* Header */}
            <div className="chat-window-header">
                <div className="chat-window-info">
                    <div className="chat-window-avatar">{isGroup ? '👥' : displayName[0]?.toUpperCase()}</div>
                    <div>
                        <div className="chat-window-name">{displayName}</div>
                        {isGroup && (
                            <div className="chat-window-sub">{conversation.members?.length || 0} thành viên</div>
                        )}
                    </div>
                </div>
                <div className="chat-window-actions">
                    {isGroup && (
                        <button className="chat-icon-btn" title="Thông tin nhóm" onClick={() => setShowInfo(!showInfo)}>
                            <svg viewBox="0 0 24 24" fill="currentColor"><path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-6h2v6zm0-8h-2V7h2v2z"/></svg>
                        </button>
                    )}
                </div>
            </div>

            {/* Group info panel */}
            {showInfo && isGroup && (
                <div className="chat-info-panel">
                    <div className="chat-info-row">
                        <span>Tên nhóm:</span>
                        <button className="chat-link-btn" onClick={onRenameGroup}>{conversation.name || 'Chưa đặt tên'} ✏️</button>
                    </div>
                    <div className="chat-info-row">
                        <span>Thành viên ({conversation.members?.length}):</span>
                        <button className="chat-link-btn" onClick={onManageMembers}>Quản lý</button>
                    </div>
                </div>
            )}

            {/* Messages */}
            <div className="chat-messages">
                {error && (
                    <div className="chat-error-msg">
                        ⚠ {error}
                        <button className="chat-retry-btn" onClick={() => {
                            setError(null)
                            setLoading(true)
                            chatApi.getMessages(convId, 0).then(msgs => {
                                const list = Array.isArray(msgs) ? msgs : []
                                setMessages(list)
                                setHasMore(list.length >= 30)
                                setLoading(false)
                            }).catch(err => {
                                console.error('[ChatWindow] retry failed:', err)
                                setError(err.message || 'Không thể tải tin nhắn')
                                setLoading(false)
                            })
                        }}>Thử lại</button>
                    </div>
                )}
                {loading && messages.length === 0 && !error && (
                    <div className="chat-loading">Đang tải tin nhắn...</div>
                )}
                {hasMore && !loading && (
                    <button className="load-more-btn" onClick={loadMore}>Tải thêm</button>
                )}
                {loading && messages.length > 0 && (
                    <div className="load-more-btn" style={{ opacity: 0.6, pointerEvents: 'none' }}>Đang tải...</div>
                )}
                {messages.map(msg => (
                    <MessageBubble key={msg.id} msg={msg} isOwn={msg.senderId === currentUserId} />
                ))}
                <div ref={bottomRef} />
            </div>

            {/* Input */}
            <div className="chat-input-bar">
                <input
                    type="file"
                    ref={fileInputRef}
                    style={{ display: 'none' }}
                    onChange={handleFileSelect}
                    accept="image/*,video/*,.pdf,.doc,.docx,.xls,.xlsx,.zip,.rar"
                />
                <button
                    className="chat-attach-btn"
                    onClick={() => fileInputRef.current?.click()}
                    disabled={uploading}
                    title="Đính kèm file"
                >
                    {uploading
                        ? <span className="uploading-spinner" />
                        : <svg viewBox="0 0 24 24" fill="currentColor"><path d="M16.5 6v11.5c0 2.21-1.79 4-4 4s-4-1.79-4-4V5c0-1.38 1.12-2.5 2.5-2.5s2.5 1.12 2.5 2.5v10.5c0 .55-.45 1-1 1s-1-.45-1-1V6H10v9.5c0 1.38 1.12 2.5 2.5 2.5s2.5-1.12 2.5-2.5V5c0-2.21-1.79-4-4-4S7 2.79 7 5v12.5c0 3.04 2.46 5.5 5.5 5.5s5.5-2.46 5.5-5.5V6h-1.5z"/></svg>
                    }
                </button>
                <textarea
                    className="chat-input"
                    placeholder="Nhập tin nhắn..."
                    value={input}
                    onChange={e => setInput(e.target.value)}
                    onKeyDown={handleKeyDown}
                    rows={1}
                />
                <button className="chat-send-btn" onClick={sendMessage} disabled={!input.trim()}>
                    <svg viewBox="0 0 24 24" fill="currentColor"><path d="M2.01 21L23 12 2.01 3 2 10l15 2-15 2z"/></svg>
                </button>
            </div>
        </div>
    )
}

export default ChatWindow
