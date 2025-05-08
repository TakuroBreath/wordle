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
            console.log('Telegram WebApp API инициализирован');

            // Устанавливаем данные в состояние
            setTg(webApp);
            setUser(webApp.initDataUnsafe?.user || null);

            // Получаем и логируем initData
            const initDataValue = webApp.initData || null;
            console.log('Получены данные инициализации Telegram:', {
                initDataLength: initDataValue ? initDataValue.length : 0,
                user: webApp.initDataUnsafe?.user
            });
            setInitData(initDataValue);

            // Устанавливаем инициализационные данные в API клиент
            if (initDataValue) {
                console.log('Установка initData в API клиент');
                api.setTelegramInitData(initDataValue);
            } else {
                console.warn('initData отсутствует или пуст');
            }

            // Настраиваем внешний вид приложения
            webApp.expand();

            // Устанавливаем основной цвет кнопки в соответствии с темой Telegram
            webApp.MainButton.setParams({
                color: webApp.themeParams.button_color,
                text_color: webApp.themeParams.button_text_color,
            });

            setIsReady(true);
        } else {
            console.warn('Telegram WebApp API не доступен, использую режим разработки');
            // Для разработки - эмулируем пользователя и окружение
            if (process.env.NODE_ENV === 'development') {
                console.log('Инициализация эмуляции Telegram WebApp для разработки');
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

                // Создаем тестового пользователя
                const testUser = {
                    id: 12345678,
                    first_name: 'Test',
                    last_name: 'User',
                    username: 'testuser',
                    language_code: 'ru'
                };
                setUser(testUser);

                // Проверяем, есть ли сохраненный токен
                const savedToken = localStorage.getItem('token');
                if (savedToken && savedToken.startsWith('dev_')) {
                    console.log('Найден сохраненный тестовый токен для режима разработки');
                    // Если есть тестовый токен, используем его
                    api.setAuthToken(savedToken);
                    setInitData(null); // Не используем initData в этом случае
                } else {
                    // Создаем тестовые данные инициализации в формате, который ожидает сервер
                    // Это должно быть в формате query-string, который Telegram отправляет в initData
                    console.log('Создание тестовых данных инициализации для режима разработки');

                    // Генерируем тестовый токен
                    const devToken = 'dev_test_token_for_development_only';
                    localStorage.setItem('token', devToken);
                    api.setAuthToken(devToken);

                    // Не используем initData, так как у нас есть тестовый токен
                    setInitData(null);
                }

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