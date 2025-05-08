import React, { createContext, useState, useEffect, useContext } from 'react';
import { authAPI, userAPI } from '../api';

// Создаем контекст
const AuthContext = createContext(null);

// Провайдер контекста
export const AuthProvider = ({ children }) => {
    const [user, setUser] = useState(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);

    // Инициализация пользователя при загрузке страницы
    useEffect(() => {
        const initAuth = async () => {
            try {
                setLoading(true);

                // Проверяем наличие токена
                const token = localStorage.getItem('token');
                if (token) {
                    try {
                        // Проверяем валидность токена
                        await authAPI.verifyAuth(token);
                        // Получаем данные пользователя
                        const userData = await userAPI.getCurrentUser();
                        setUser(userData.data);
                        setLoading(false);
                        return;
                    } catch (err) {
                        // Если токен невалиден, удаляем его
                        localStorage.removeItem('token');
                    }
                }

                // Если нет токена или он невалиден, пробуем авторизацию через Telegram
                if (window.Telegram && window.Telegram.WebApp) {
                    const initData = localStorage.getItem('telegram_init_data');
                    if (initData) {
                        const authData = await authAPI.telegramAuth();
                        if (authData.token) {
                            localStorage.setItem('token', authData.token);
                            const userData = await userAPI.getCurrentUser();
                            setUser(userData.data);
                        }
                    }
                }
            } catch (err) {
                console.error('Auth initialization error:', err);
                setError(err.message || 'Ошибка авторизации');
            } finally {
                setLoading(false);
            }
        };

        initAuth();
    }, []);

    // Функция выхода
    const logout = async () => {
        try {
            await authAPI.logout();
            localStorage.removeItem('token');
            setUser(null);
        } catch (err) {
            console.error('Logout error:', err);
            setError(err.message || 'Ошибка при выходе из системы');
        }
    };

    // Обновление данных пользователя
    const refreshUserData = async () => {
        try {
            setLoading(true);
            const userData = await userAPI.getCurrentUser();
            setUser(userData.data);
        } catch (err) {
            console.error('Error refreshing user data:', err);
            setError(err.message || 'Ошибка обновления данных пользователя');
        } finally {
            setLoading(false);
        }
    };

    const value = {
        user,
        loading,
        error,
        logout,
        refreshUserData,
        isAuthenticated: !!user,
    };

    return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

// Хук для использования контекста
export const useAuth = () => {
    const context = useContext(AuthContext);
    if (!context) {
        throw new Error('useAuth must be used within an AuthProvider');
    }
    return context;
};

export default AuthContext; 