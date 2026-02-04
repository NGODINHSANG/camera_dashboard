const API_BASE_URL = import.meta.env.VITE_API_URL || '/api';

class ApiClient {
    constructor() {
        this.baseUrl = API_BASE_URL;
    }

    getToken() {
        return localStorage.getItem('auth_token');
    }

    setToken(token) {
        localStorage.setItem('auth_token', token);
    }

    removeToken() {
        localStorage.removeItem('auth_token');
    }

    async request(endpoint, options = {}) {
        const url = `${this.baseUrl}${endpoint}`;
        const token = this.getToken();

        const config = {
            ...options,
            headers: {
                'Content-Type': 'application/json',
                ...(token && { Authorization: `Bearer ${token}` }),
                ...options.headers,
            },
        };

        try {
            const response = await fetch(url, config);

            // Xử lý response rỗng (204 No Content hoặc body rỗng)
            const contentType = response.headers.get('content-type');
            const text = await response.text();
            let data = null;

            if (text && contentType?.includes('application/json')) {
                try {
                    data = JSON.parse(text);
                } catch {
                    data = { message: text };
                }
            } else if (text) {
                data = { message: text };
            }

            if (!response.ok) {
                if (response.status === 401) {
                    this.removeToken();
                    window.location.href = '/login';
                }
                throw new Error(data?.error?.message || data?.message || 'Request failed');
            }

            return data || {};
        } catch (error) {
            if (error.message === 'Failed to fetch') {
                throw new Error('Không thể kết nối đến server');
            }
            throw error;
        }
    }

    get(endpoint) {
        return this.request(endpoint, { method: 'GET' });
    }

    post(endpoint, body) {
        return this.request(endpoint, { method: 'POST', body: JSON.stringify(body) });
    }

    put(endpoint, body) {
        return this.request(endpoint, { method: 'PUT', body: JSON.stringify(body) });
    }

    delete(endpoint) {
        return this.request(endpoint, { method: 'DELETE' });
    }
}

export const apiClient = new ApiClient();
export default apiClient;
