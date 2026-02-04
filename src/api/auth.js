import apiClient from './client';

export const authApi = {
    login: (email, password) =>
        apiClient.post('/auth/login', { email, password }),

    register: (email, password, name) =>
        apiClient.post('/auth/register', { email, password, name }),

    getMe: () =>
        apiClient.get('/auth/me'),
};

export default authApi;
