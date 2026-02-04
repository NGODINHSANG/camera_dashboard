import apiClient from './client';

export const projectsApi = {
    getAll: () =>
        apiClient.get('/projects'),

    getById: (id) =>
        apiClient.get(`/projects/${id}`),

    create: (name) =>
        apiClient.post('/projects', { name }),

    update: (id, name) =>
        apiClient.put(`/projects/${id}`, { name }),

    delete: (id) =>
        apiClient.delete(`/projects/${id}`),
};

export default projectsApi;
