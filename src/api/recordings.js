import apiClient from './client';

export const recordingsApi = {
    // Start recording a camera
    start: (cameraId, outputDir) =>
        apiClient.post('/recordings/start', { cameraId, outputDir }),

    // Stop a recording
    stop: (recordingId) =>
        apiClient.post(`/recordings/${recordingId}/stop`),

    // Get all recordings
    getAll: () =>
        apiClient.get('/recordings'),

    // Get active recordings
    getActive: () =>
        apiClient.get('/recordings/active'),

    // Get a specific recording
    getById: (id) =>
        apiClient.get(`/recordings/${id}`),

    // Check if camera is recording
    getCameraStatus: (cameraId) =>
        apiClient.get(`/recordings/camera/${cameraId}/status`),

    // Delete a recording
    delete: (id) =>
        apiClient.delete(`/recordings/${id}`),

    // Get download URL
    getDownloadUrl: (id) =>
        `${apiClient.defaults.baseURL}/recordings/${id}/download`,

    // Browse server directories
    browse: (path = '') => {
        // Manually encode path to handle Windows backslashes
        const encodedPath = encodeURIComponent(path)
        return apiClient.get(`/recordings/browse?path=${encodedPath}`)
    },

    // Validate a path is writable
    validatePath: (path) =>
        apiClient.post('/recordings/validate-path', { path }),
};

export default recordingsApi;
