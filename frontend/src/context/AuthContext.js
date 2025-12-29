import React, { createContext, useState, useEffect, useContext, useCallback } from 'react';
import { authAPI, userAPI } from '../api';

// Создаем контекст
const AuthContext = createContext(null);

// Провайдер контекста
export const AuthProvider = ({ children }) => {
    const [user, setUser] = useState(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [isInitialized, setIsInitialized] = useState(false);

    // Инициализация пользователя при загрузке страницы
    useEffect(() => {
        const initAuth = async () => {
            try {
                setLoading(true);
                setError(null);

                // Проверяем наличие токена
                const token = localStorage.getItem('token');
                if (token) {
                    try {
                        // Проверяем валидность токена
                        await authAPI.verifyAuth(token);
                        // Получаем данные пользователя
                        const userData = await userAPI.getCurrentUser();
                        setUser(userData.data);
                        setIsInitialized(true);
                        setLoading(false);
                        return;
                    } catch (err) {
                        console.log('Token invalid, trying Telegram auth');
                        // Если токен невалиден, удаляем его
                        localStorage.removeItem('token');
                    }
                }

                // Если нет токена или он невалиден, пробуем авторизацию через Telegram
                const initData = localStorage.getItem('telegram_init_data');
                if (initData) {
                    try {
                        console.log('Attempting Telegram auth...');
                        const authData = await authAPI.telegramAuth();
                        if (authData.token) {
                            localStorage.setItem('token', authData.token);
                            const userData = await userAPI.getCurrentUser();
                            setUser(userData.data);
                            console.log('Telegram auth successful');
                        }
                    } catch (authErr) {
                        console.error('Telegram auth failed:', authErr);
                        setError('Ошибка авторизации через Telegram');
                    }
                } else {
                    console.log('No Telegram init data available');
                }
            } catch (err) {
                console.error('Auth initialization error:', err);
                setError(err.message || 'Ошибка авторизации');
            } finally {
                setLoading(false);
                setIsInitialized(true);
            }
        };

        initAuth();
    }, []);

    // Функция выхода
    const logout = useCallback(async () => {
        try {
            await authAPI.logout();
        } catch (err) {
            console.error('Logout error:', err);
        } finally {
            localStorage.removeItem('token');
            setUser(null);
        }
    }, []);

    // Обновление данных пользователя
    const refreshUserData = useCallback(async () => {
        try {
            const userData = await userAPI.getCurrentUser();
            setUser(userData.data);
            return userData.data;
        } catch (err) {
            console.error('Error refreshing user data:', err);
            // Если ошибка авторизации, выходим
            if (err.response?.status === 401) {
                logout();
            }
            throw err;
        }
    }, [logout]);

    // Повторная попытка авторизации через Telegram
    const retryTelegramAuth = useCallback(async () => {
        const initData = localStorage.getItem('telegram_init_data');
        if (!initData) {
            throw new Error('No Telegram init data');
        }

        try {
            setLoading(true);
            setError(null);
            
            const authData = await authAPI.telegramAuth();
            if (authData.token) {
                localStorage.setItem('token', authData.token);
                const userData = await userAPI.getCurrentUser();
                setUser(userData.data);
                return userData.data;
            }
        } catch (err) {
            console.error('Telegram auth retry failed:', err);
            setError('Ошибка авторизации');
            throw err;
        } finally {
            setLoading(false);
        }
    }, []);

    // Проверка, запущено ли в Telegram
    const isTelegramWebApp = useCallback(() => {
        return !!(window.Telegram && window.Telegram.WebApp && window.Telegram.WebApp.initData);
    }, []);

    const value = {
        user,
        loading,
        error,
        isInitialized,
        logout,
        refreshUserData,
        retryTelegramAuth,
        isTelegramWebApp,
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
