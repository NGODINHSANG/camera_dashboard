import apiClient from './client';

export const camerasApi = {
    getByProject: (projectId) =>
        apiClient.get(`/projects/${projectId}/cameras`),

    create: (projectId, cameraData) =>
        apiClient.post(`/projects/${projectId}/cameras`, cameraData),

    update: (projectId, cameraId, cameraData) =>
        apiClient.put(`/projects/${projectId}/cameras/${cameraId}`, cameraData),

    delete: (projectId, cameraId) =>
        apiClient.delete(`/projects/${projectId}/cameras/${cameraId}`),
};

export default camerasApi;
