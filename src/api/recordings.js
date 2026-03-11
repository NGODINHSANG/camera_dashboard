import apiClient from './client';

export const recordingsApi = {
    // Start recording a camera (auto-saves to /recordings/<project>/<camera>)
    start: (cameraId) =>
        apiClient.post('/recordings/start', { cameraId }),

    // Stop a recording
    stop: (recordingId) =>
        apiClient.post(`/recordings/${recordingId}/stop`),

    // Get all recordings (DB-based)
    getAll: () =>
        apiClient.get('/recordings'),

    // Get active recordings
    getActive: () =>
        apiClient.get('/recordings/active'),

    // Get a specific recording (DB-based)
    getById: (id) =>
        apiClient.get(`/recordings/${id}`),

    // Check if camera is recording
    getCameraStatus: (cameraId) =>
        apiClient.get(`/recordings/camera/${cameraId}/status`),

    // Delete a recording (DB-based)
    delete: (id) =>
        apiClient.delete(`/recordings/${id}`),

    // Get download URL (DB-based)
    getDownloadUrl: (id) =>
        `${apiClient.baseUrl}/recordings/${id}/download`,

    // Get stream URL (DB-based)
    getStreamUrl: (id) =>
        `${apiClient.baseUrl}/recordings/${id}/stream`,

    // Get all completed recordings for a camera (DB-based)
    getByCamera: (cameraId) =>
        apiClient.get(`/recordings/camera/${cameraId}/list`),

    // ========== File-based methods (new) ==========

    // List video files in camera's recording directory
    listFiles: (projectName, cameraName) =>
        apiClient.get(`/recordings/files/${encodeURIComponent(projectName)}/${encodeURIComponent(cameraName)}`),

    // Get stream URL for a video file (direct streaming - fallback)
    getFileStreamUrl: (filePath) =>
        `${apiClient.baseUrl}/recordings/files/stream?path=${encodeURIComponent(filePath)}`,

    // Get restream URL - FFmpeg generates HLS on-the-fly (instant, bandwidth efficient)
    getRestreamUrl: (filePath) =>
        `${apiClient.baseUrl}/recordings/restream?path=${encodeURIComponent(filePath)}`,

    // Get download URL for a video file
    getFileDownloadUrl: (filePath) =>
        `${apiClient.baseUrl}/recordings/files/download?path=${encodeURIComponent(filePath)}`,

    // Delete a video file
    deleteFile: (filePath) =>
        apiClient.delete(`/recordings/files/delete?path=${encodeURIComponent(filePath)}`),

    // Browse server directories (legacy, for custom path selection)
    browse: (path = '') => {
        const encodedPath = encodeURIComponent(path)
        return apiClient.get(`/recordings/browse?path=${encodedPath}`)
    },

    // Validate a path is writable (legacy)
    validatePath: (path) =>
        apiClient.post('/recordings/validate-path', { path }),
};

export default recordingsApi;
