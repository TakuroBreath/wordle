import React, { createContext, useContext, useEffect, useState } from 'react';
import { useTelegram } from './TelegramContext';
import api from '../services/api';

const AuthContext = createContext(null);

export const useAuth = () => {
    const context = useContext(AuthContext);
    if (!context) {
        throw new Error('useAuth должен быть использован внутри AuthProvider');
    }
    return context;
};

export function AuthProvider({ children }) {
    const { initData, isReady, user } = useTelegram();
    const [token, setToken] = useState(localStorage.getItem('token') || null);
    const [currentUser, setCurrentUser] = useState(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);

    // Инициализация аутентификации при загрузке компонента
    useEffect(() => {
        const initAuth = async () => {
            if (!isReady || !initData) {
                console.log('Аутентификация не инициализирована:', { isReady, initData });
                return;
            }

            try {
                setLoading(true);
                console.log('Начало аутентификации с данными:', initData);

                // Аутентификация с использованием данных Telegram Mini App
                const response = await api.auth.telegramAuth({ init_data: initData });
                console.log('Ответ сервера:', response.data);

                // Сохраняем токен в localStorage и состояние
                if (response.data?.token) {
                    localStorage.setItem('token', response.data.token);
                    setToken(response.data.token);

                    // Добавляем токен в заголовки API запросов
                    api.setAuthToken(response.data.token);

                    // Получаем данные текущего пользователя
                    const userResponse = await api.user.getCurrent();
                    console.log('Данные пользователя:', userResponse.data);
                    setCurrentUser(userResponse.data);
                }
            } catch (err) {
                console.error('Ошибка аутентификации:', err);
                setError(err.message || 'Ошибка аутентификации');
            } finally {
                setLoading(false);
            }
        };

        // Для разработки - эмулируем пользователя без выполнения аутентификации
        if (process.env.NODE_ENV === 'development' && !token && user && !initData) {
            console.log('Режим разработки: эмуляция пользователя');
            setCurrentUser({
                telegram_id: user.id,
                username: user.username,
                first_name: user.first_name,
                last_name: user.last_name,
                balance_ton: 10.5,
                balance_usdt: 25.0,
                wins: 5,
                losses: 2
            });
            setLoading(false);
        } else {
            initAuth();
        }
    }, [initData, isReady, user]);

    // Проверка токена при загрузке приложения
    useEffect(() => {
        const verifyToken = async () => {
            if (!token) {
                setLoading(false);
                return;
            }

            try {
                setLoading(true);

                // Добавляем токен в заголовки API запросов
                api.setAuthToken(token);

                // Проверяем валидность токена и получаем данные пользователя
                const response = await api.auth.verifyToken({ token });

                if (response.data?.authenticated) {
                    setCurrentUser(response.data.user);
                } else {
                    // Токен недействителен - удаляем из хранилища
                    localStorage.removeItem('token');
                    setToken(null);
                    api.setAuthToken(null);
                }
            } catch (err) {
                console.error('Ошибка проверки токена:', err);
                localStorage.removeItem('token');
                setToken(null);
                api.setAuthToken(null);
            } finally {
                setLoading(false);
            }
        };

        if (process.env.NODE_ENV !== 'development' || (token && !currentUser)) {
            verifyToken();
        }
    }, [token]);

    // Функция выхода из аккаунта
    const logout = async () => {
        if (token) {
            try {
                await api.auth.logout({ token });
            } catch (err) {
                console.error('Ошибка при выходе:', err);
            }
        }

        localStorage.removeItem('token');
        setToken(null);
        setCurrentUser(null);
        api.setAuthToken(null);
    };

    const value = {
        isAuthenticated: !!currentUser,
        user: currentUser,
        token,
        loading,
        error,
        logout
    };

    return (
        <AuthContext.Provider value={value}>
            {children}
        </AuthContext.Provider>
    );
} 