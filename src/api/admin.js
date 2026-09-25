import apiClient from './client';

export const adminApi = {
    getUsers: () =>
        apiClient.get('/admin/users'),

    getUser: (id) =>
        apiClient.get(`/admin/users/${id}`),

    deleteUser: (id) =>
        apiClient.delete(`/admin/users/${id}`),

    updateUserRole: (id, role) =>
        apiClient.put(`/admin/users/${id}/role`, { role }),

    getUserProjectPermissions: (userId) =>
        apiClient.get(`/admin/users/${userId}/project-permissions`),

    getAllProjects: () =>
        apiClient.get('/admin/projects'),

    getProjectMembers: (projectId) =>
        apiClient.get(`/admin/projects/${projectId}/members`),

    grantProjectPermission: (projectId, userId) =>
        apiClient.post(`/admin/projects/${projectId}/members/${userId}`),

    revokeProjectPermission: (projectId, userId) =>
        apiClient.delete(`/admin/projects/${projectId}/members/${userId}`),
};

export default adminApi;
