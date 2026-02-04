import { createContext, useContext, useState, useEffect, useCallback } from 'react';
import { authApi } from '../api/auth';
import apiClient from '../api/client';

const AuthContext = createContext(null);

export function AuthProvider({ children }) {
    const [user, setUser] = useState(null);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const token = apiClient.getToken();
        if (token) {
            authApi.getMe()
                .then(response => setUser(response.data))
                .catch(() => apiClient.removeToken())
                .finally(() => setLoading(false));
        } else {
            setLoading(false);
        }
    }, []);

    const login = useCallback(async (email, password) => {
        const response = await authApi.login(email, password);
        apiClient.setToken(response.data.token);
        setUser(response.data.user);
        return response.data.user;
    }, []);

    const register = useCallback(async (email, password, name) => {
        const response = await authApi.register(email, password, name);
        apiClient.setToken(response.data.token);
        setUser(response.data.user);
        return response.data.user;
    }, []);

    const logout = useCallback(() => {
        apiClient.removeToken();
        setUser(null);
    }, []);

    const value = {
        user,
        loading,
        isAuthenticated: !!user,
        isAdmin: user?.role === 'admin',
        login,
        register,
        logout,
    };

    return (
        <AuthContext.Provider value={value}>
            {children}
        </AuthContext.Provider>
    );
}

export function useAuth() {
    const context = useContext(AuthContext);
    if (!context) {
        throw new Error('useAuth must be used within an AuthProvider');
    }
    return context;
}

export default AuthContext;
