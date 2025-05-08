import React, { createContext, useContext, useEffect, useState } from 'react';
import api from '../services/api';

const TelegramContext = createContext(null);

export const useTelegram = () => {
    const context = useContext(TelegramContext);
    if (!context) {
        throw new Error('useTelegram должен быть использован внутри TelegramProvider');
    }
    return context;
};

export function TelegramProvider({ children }) {
    const [tg, setTg] = useState(null);
    const [user, setUser] = useState(null);
    const [initData, setInitData] = useState(null);
    const [isReady, setIsReady] = useState(false);

    useEffect(() => {
        // Получаем доступ к Telegram WebApp API
        const webApp = window.Telegram?.WebApp;

        if (webApp) {
            // Инициализация завершена - API доступно
            webApp.ready();

            // Устанавливаем данные в состояние
            setTg(webApp);
            setUser(webApp.initDataUnsafe?.user || null);
            setInitData(webApp.initData || null);

            // Устанавливаем инициализационные данные в API клиент
            api.setTelegramInitData(webApp.initData);

            // Настраиваем внешний вид приложения
            webApp.expand();

            // Устанавливаем основной цвет кнопки в соответствии с темой Telegram
            webApp.MainButton.setParams({
                color: webApp.themeParams.button_color,
                text_color: webApp.themeParams.button_text_color,
            });

            setIsReady(true);
        } else {
            console.error('Telegram WebApp API не доступен');
            // Для разработки - эмулируем пользователя и окружение
            if (process.env.NODE_ENV === 'development') {
                setTg({
                    MainButton: {
                        show: () => console.log('MainButton.show'),
                        hide: () => console.log('MainButton.hide'),
                        setParams: (params) => console.log('MainButton.setParams', params),
                        onClick: (callback) => console.log('MainButton.onClick', callback)
                    },
                    BackButton: {
                        show: () => console.log('BackButton.show'),
                        hide: () => console.log('BackButton.hide'),
                        onClick: (callback) => console.log('BackButton.onClick', callback)
                    },
                    themeParams: {
                        bg_color: '#ffffff',
                        text_color: '#000000',
                        hint_color: '#999999',
                        button_color: '#007aff',
                        button_text_color: '#ffffff',
                    },
                    isExpanded: true,
                    expand: () => console.log('WebApp.expand'),
                    close: () => console.log('WebApp.close'),
                    ready: () => console.log('WebApp.ready')
                });

                setUser({
                    id: 12345678,
                    first_name: 'Test',
                    last_name: 'User',
                    username: 'testuser',
                    language_code: 'ru'
                });

                // Создаем тестовые данные инициализации в формате, который ожидает сервер
                const testInitData = 'user=%7B%22id%22%3A12345678%2C%22first_name%22%3A%22Test%22%2C%22last_name%22%3A%22User%22%2C%22username%22%3A%22testuser%22%2C%22language_code%22%3A%22ru%22%7D&auth_date=1234567890&hash=1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef';
                setInitData(testInitData);
                api.setTelegramInitData(testInitData);
                setIsReady(true);
            }
        }
    }, []);

    const value = {
        tg,
        user,
        initData,
        isReady
    };

    return (
        <TelegramContext.Provider value={value}>
            {children}
        </TelegramContext.Provider>
    );
} 