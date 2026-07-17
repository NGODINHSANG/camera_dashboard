import { useState, useEffect, useRef, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'
import { chatApi, createChatWebSocket } from '../api/chat'
import ChatSidebar from '../components/Chat/ChatSidebar'
import ChatWindow from '../components/Chat/ChatWindow'
import { CreateGroupModal, RenameGroupModal, ManageMembersModal } from '../components/Chat/ChatModals'
import './MessagesPage.css'

function MessagesPage() {
    const navigate = useNavigate()
    const { user } = useAuth()
    const [conversations, setConversations] = useState([])
    const [activeConv, setActiveConv] = useState(null)
    const [onlineUserIds, setOnlineUserIds] = useState([])
    const [loading, setLoading] = useState(true)
    const [showCreateGroup, setShowCreateGroup] = useState(false)
    const [showRename, setShowRename] = useState(false)
    const [showManageMembers, setShowManageMembers] = useState(false)
    const wsRef = useRef(null)
    const activeConvRef = useRef(null)

    useEffect(() => { activeConvRef.current = activeConv }, [activeConv])

    const loadConversations = useCallback(async () => {
        try {
            const data = await chatApi.getConversations()
            const list = Array.isArray(data) ? data : []
            setConversations(list)
            // Sync activeConv with fresh data (picks up otherUser, name changes, etc.)
            setActiveConv(prev => {
                if (!prev) return prev
                return list.find(c => c.id === prev.id) || prev
            })
        } catch (_) {}
        setLoading(false)
    }, [])

    useEffect(() => {
        loadConversations()
    }, [loadConversations])

    // Kết nối WebSocket
    useEffect(() => {
        const ws = createChatWebSocket(handleWsEvent)
        wsRef.current = ws
        return () => { ws.close() }
    }, [])

    const handleWsEvent = (event) => {
        switch (event.type) {
            case 'user_online':
                setOnlineUserIds(prev => [...new Set([...prev, event.payload.userId])])
                break
            case 'user_offline':
                setOnlineUserIds(prev => prev.filter(id => id !== event.payload.userId))
                break
            case 'new_message':
                // Dispatch cho ChatWindow
                window.dispatchEvent(new CustomEvent('chat:new_message', { detail: event.payload }))
                // Cập nhật lastMessage trong danh sách
                setConversations(prev => prev.map(c =>
                    c.id === event.payload.conversationId
                        ? { ...c, lastMessage: event.payload, unreadCount: c.id === activeConvRef.current?.id ? 0 : c.unreadCount + 1 }
                        : c
                ).sort((a, b) => new Date(b.updatedAt) - new Date(a.updatedAt)))
                break
            case 'conversation_created':
                setConversations(prev => [event.payload, ...prev])
                break
            case 'group_renamed':
                setConversations(prev => prev.map(c =>
                    c.id === event.payload.conversationId ? { ...c, name: event.payload.name } : c
                ))
                setActiveConv(prev => prev?.id === event.payload.conversationId
                    ? { ...prev, name: event.payload.name } : prev)
                break
            case 'member_added':
            case 'member_removed':
                loadConversations()
                break
        }
    }

    const handleSelectConv = (conv) => {
        setActiveConv(conv)
        setConversations(prev => prev.map(c => c.id === conv.id ? { ...c, unreadCount: 0 } : c))
        chatApi.markRead(conv.id).catch(() => {})
    }

    const handleCreateDM = async (targetUser) => {
        if (!targetUser) return
        try {
            const conv = await chatApi.createConversation({ type: 'dm', userId: targetUser.id })
            // Ensure otherUser is set — backend sets it, but use targetUser as fallback
            const enriched = conv.otherUser
                ? conv
                : { ...conv, otherUser: { id: targetUser.id, name: targetUser.name, email: targetUser.email } }
            setConversations(prev => {
                if (prev.find(c => c.id === enriched.id)) return prev
                return [enriched, ...prev]
            })
            setActiveConv(enriched)
        } catch (_) {}
    }

    const handleGroupCreated = (conv) => {
        setConversations(prev => [conv, ...prev])
        setActiveConv(conv)
    }

    const handleRenameGroup = (newName) => {
        setActiveConv(prev => ({ ...prev, name: newName }))
        setConversations(prev => prev.map(c => c.id === activeConv?.id ? { ...c, name: newName } : c))
    }

    const totalUnread = conversations.reduce((sum, c) => sum + (c.unreadCount || 0), 0)

    return (
        <div className="messages-page">
            <div className="messages-header">
                <button className="back-btn" onClick={() => navigate('/')}>
                    <svg viewBox="0 0 24 24" fill="currentColor">
                        <path d="M20 11H7.83l5.59-5.59L12 4l-8 8 8 8 1.41-1.41L7.83 13H20v-2z" />
                    </svg>
                    Quay lại
                </button>
                <h1 className="messages-title">
                    💬 TIN NHẮN
                    {totalUnread > 0 && <span className="header-unread-badge">{totalUnread}</span>}
                </h1>
            </div>

            <div className="messages-body">
                <ChatSidebar
                    currentUserId={user?.id}
                    activeConvId={activeConv?.id}
                    onSelectConv={handleSelectConv}
                    onCreateDM={handleCreateDM}
                    onCreateGroup={() => setShowCreateGroup(true)}
                    onlineUserIds={onlineUserIds}
                    conversations={conversations}
                    loading={loading}
                />
                <ChatWindow
                    conversation={activeConv}
                    currentUserId={user?.id}
                    onRenameGroup={() => setShowRename(true)}
                    onManageMembers={() => setShowManageMembers(true)}
                />
            </div>

            {showCreateGroup && (
                <CreateGroupModal
                    currentUserId={user?.id}
                    onClose={() => setShowCreateGroup(false)}
                    onCreated={handleGroupCreated}
                />
            )}
            {showRename && activeConv && (
                <RenameGroupModal
                    conversation={activeConv}
                    onClose={() => setShowRename(false)}
                    onRenamed={handleRenameGroup}
                />
            )}
            {showManageMembers && activeConv && (
                <ManageMembersModal
                    conversation={activeConv}
                    currentUserId={user?.id}
                    onClose={() => setShowManageMembers(false)}
                    onUpdated={loadConversations}
                />
            )}
        </div>
    )
}

export default MessagesPage
