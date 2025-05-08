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
            if (!isReady) {
                console.log('Telegram WebApp не готов');
                return;
            }

            if (!initData) {
                console.log('Отсутствуют данные инициализации Telegram');
                // В режиме разработки можно продолжить без initData
                if (process.env.NODE_ENV === 'development' && user) {
                    console.log('Режим разработки: эмуляция пользователя без initData');

                    // Генерируем тестовый токен для режима разработки
                    const devToken = 'dev_test_token_for_development_only';
                    console.log('Генерация тестового токена для режима разработки');

                    // Сохраняем токен
                    localStorage.setItem('token', devToken);
                    setToken(devToken);

                    // Устанавливаем токен в API
                    api.setAuthToken(devToken);
                    console.log('Тестовый токен установлен для режима разработки');

                    // Устанавливаем тестового пользователя
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
                }
                return;
            }

            try {
                setLoading(true);
                console.log('Начало аутентификации с данными Telegram');

                // Проверяем, есть ли уже токен в localStorage
                const savedToken = localStorage.getItem('token');
                if (savedToken) {
                    console.log('Найден сохраненный токен, проверяем его валидность...');
                    try {
                        // Устанавливаем токен в заголовки API запросов
                        api.setAuthToken(savedToken);

                        // Проверяем валидность токена, получая данные пользователя
                        const userResponse = await api.user.getCurrent();
                        console.log('Токен валиден, получены данные пользователя:', userResponse.data);
                        setCurrentUser(userResponse.data);
                        setToken(savedToken);
                        setLoading(false);
                        return;
                    } catch (tokenErr) {
                        console.error('Сохраненный токен недействителен:', tokenErr);
                        localStorage.removeItem('token');
                        api.setAuthToken(null);
                        // Продолжаем процесс аутентификации
                    }
                }

                // Аутентификация с использованием данных Telegram Mini App
                console.log('Отправка запроса на аутентификацию с данными Telegram Mini App');
                try {
                    const response = await api.auth.telegramAuth({ init_data: initData });
                    console.log('Ответ сервера при аутентификации:', response.data);

                    // Сохраняем токен в localStorage и состояние
                    if (response.data?.token) {
                        console.log('Получен токен авторизации');
                        localStorage.setItem('token', response.data.token);
                        setToken(response.data.token);

                        // Добавляем токен в заголовки API запросов
                        api.setAuthToken(response.data.token);
                        console.log('Токен установлен в заголовки API');

                        // Получаем данные текущего пользователя
                        try {
                            const userResponse = await api.user.getCurrent();
                            console.log('Данные пользователя:', userResponse.data);
                            setCurrentUser(userResponse.data);
                        } catch (userErr) {
                            console.error('Ошибка при получении данных пользователя:', userErr);
                            throw new Error('Не удалось получить данные пользователя');
                        }
                    } else {
                        console.error('Токен не получен в ответе сервера');
                        throw new Error('Токен не получен');
                    }
                } catch (authErr) {
                    console.error('Ошибка при аутентификации через Telegram:', authErr);

                    // В режиме разработки можно эмулировать пользователя при ошибке аутентификации
                    if (process.env.NODE_ENV === 'development' && user) {
                        console.log('Режим разработки: эмуляция пользователя после ошибки аутентификации');
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
                    } else {
                        throw authErr;
                    }
                }
            } catch (err) {
                console.error('Ошибка аутентификации:', err);
                setError(err.message || 'Ошибка аутентификации');
                // Очищаем токен при ошибке
                localStorage.removeItem('token');
                setToken(null);
                api.setAuthToken(null);
            } finally {
                setLoading(false);
            }
        };

        // Вызываем инициализацию аутентификации
        initAuth();
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
                console.log('Проверка валидности токена:', token.substring(0, 15) + '...');

                // В режиме разработки, если токен начинается с 'dev_', считаем его валидным
                if (process.env.NODE_ENV === 'development' && token.startsWith('dev_')) {
                    console.log('Режим разработки: пропускаем проверку тестового токена');
                    if (!currentUser) {
                        setCurrentUser({
                            telegram_id: 12345678,
                            username: 'testuser',
                            first_name: 'Test',
                            last_name: 'User',
                            balance_ton: 10.5,
                            balance_usdt: 25.0,
                            wins: 5,
                            losses: 2
                        });
                    }
                    setLoading(false);
                    return;
                }

                // Добавляем токен в заголовки API запросов
                api.setAuthToken(token);

                // Проверяем валидность токена и получаем данные пользователя
                try {
                    // Сначала пробуем получить данные пользователя напрямую
                    console.log('Получение данных пользователя с текущим токеном...');
                    const userResponse = await api.user.getCurrent();
                    console.log('Данные пользователя получены:', userResponse.data);
                    setCurrentUser(userResponse.data);
                } catch (userErr) {
                    console.error('Ошибка при получении данных пользователя:', userErr);

                    // Если не удалось получить данные пользователя, пробуем верифицировать токен
                    console.log('Попытка верификации токена через API...');
                    const response = await api.auth.verifyToken({ token });
                    console.log('Результат верификации токена:', response.data);

                    if (response.data?.authenticated) {
                        setCurrentUser(response.data.user);
                    } else {
                        console.error('Токен недействителен по данным API верификации');
                        throw new Error('Токен недействителен');
                    }
                }
            } catch (err) {
                console.error('Ошибка проверки токена:', err);
                localStorage.removeItem('token');
                setToken(null);
                setCurrentUser(null);
                api.setAuthToken(null);
            } finally {
                setLoading(false);
            }
        };

        if (process.env.NODE_ENV !== 'development' || (token && !currentUser && !token.startsWith('dev_'))) {
            verifyToken();
        }
    }, [token, currentUser]);

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