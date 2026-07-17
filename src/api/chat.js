import apiClient from './client'

const API_BASE = import.meta.env.VITE_API_URL || ''

export const chatApi = {
    searchUsers: (q) => apiClient.get(`/chat/users/search?q=${encodeURIComponent(q)}`),
    getOnlineUsers: () => apiClient.get('/chat/online'),

    getConversations: () => apiClient.get('/chat/conversations'),
    createConversation: (data) => apiClient.post('/chat/conversations', data),
    renameGroup: (id, name) => apiClient.put(`/chat/conversations/${id}/name`, { name }),
    addMember: (convId, userId) => apiClient.post(`/chat/conversations/${convId}/members/${userId}`, {}),
    removeMember: (convId, userId) => apiClient.delete(`/chat/conversations/${convId}/members/${userId}`),

    getMessages: (convId, offset = 0) => apiClient.get(`/chat/conversations/${convId}/messages?offset=${offset}`),
    sendMessage: (convId, content) => apiClient.post(`/chat/conversations/${convId}/messages`, { content, type: 'text' }),

    uploadAttachment: async (convId, file) => {
        const token = localStorage.getItem('auth_token')
        const fd = new FormData()
        fd.append('file', file)
        const res = await fetch(`${API_BASE}/api/chat/conversations/${convId}/attachments`, {
            method: 'POST',
            headers: token ? { Authorization: `Bearer ${token}` } : {},
            body: fd,
        })
        if (!res.ok) throw new Error('Upload failed')
        return res.json()
    },

    markRead: (convId) => apiClient.post(`/chat/conversations/${convId}/read`, {}),
    getAttachmentUrl: (url) => `${API_BASE}${url}`,
}

export function createChatWebSocket(onEvent) {
    const token = localStorage.getItem('auth_token')
    const wsBase = (API_BASE || window.location.origin).replace(/^http/, 'ws')
    const ws = new WebSocket(`${wsBase}/api/chat/ws?token=${token}`)
    ws.onmessage = (e) => {
        try { onEvent(JSON.parse(e.data)) } catch (_) {}
    }
    ws.onerror = (e) => console.error('[WS] error', e)
    return ws
}
