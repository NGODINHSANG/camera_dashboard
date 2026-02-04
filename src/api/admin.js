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

    getAllProjects: () =>
        apiClient.get('/admin/projects'),
};

export default adminApi;
