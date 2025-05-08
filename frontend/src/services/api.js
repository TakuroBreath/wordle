import axios from 'axios';

// Создаем экземпляр axios с базовым URL для API
const apiClient = axios.create({
    baseURL: process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1',
    headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
        'X-Requested-With': 'XMLHttpRequest'
    },
    withCredentials: false // Отключаем отправку куки для решения проблем с CORS
});

// Добавляем перехватчик для обработки ошибок
apiClient.interceptors.response.use(
    (response) => {
        return response;
    },
    (error) => {
        if (error.message === 'Network Error') {
            console.error('Ошибка сети (возможно CORS):', error);
        }
        return Promise.reject(error);
    }
);

// Функция для проверки, находимся ли мы в режиме разработки
const isDevelopmentMode = () => {
    return process.env.NODE_ENV === 'development';
};

// Функция для проверки, используем ли мы тестовый токен
const isUsingDevToken = () => {
    const token = localStorage.getItem('token');
    return isDevelopmentMode() && token && token.startsWith('dev_');
};

// Функция для установки токена авторизации
const setAuthToken = (token) => {
    console.log('Установка токена авторизации:', token ? 'токен установлен' : 'токен удален');
    if (token) {
        // Проверяем, начинается ли токен с "Bearer "
        const bearerToken = token.startsWith('Bearer ') ? token : `Bearer ${token}`;
        apiClient.defaults.headers.common['Authorization'] = bearerToken;
        console.log('Установлен токен:', bearerToken.substring(0, 15) + '...');

        // Сохраняем токен в localStorage для последующих запросов
        localStorage.setItem('token', token);
    } else {
        delete apiClient.defaults.headers.common['Authorization'];
        console.log('Токен авторизации удален');

        // Удаляем токен из localStorage
        localStorage.removeItem('token');
    }
    console.log('Текущие заголовки:', apiClient.defaults.headers);
};

// Функция для установки инициализационных данных Telegram Mini App
const setTelegramInitData = (initData) => {
    if (initData) {
        // Сохраняем текущий токен, если он есть
        const currentToken = localStorage.getItem('token');
        if (currentToken) {
            // Если у нас уже есть токен, используем его вместо TMA данных
            setAuthToken(currentToken);
        } else {
            // Иначе используем TMA данные
            apiClient.defaults.headers.common['Authorization'] = `tma ${initData}`;
            console.log('Установлены TMA данные в заголовок Authorization');
        }
    } else {
        // Проверяем, есть ли сохраненный токен
        const savedToken = localStorage.getItem('token');
        if (savedToken) {
            // Если есть сохраненный токен, используем его
            setAuthToken(savedToken);
        } else {
            // Иначе удаляем заголовок авторизации
            delete apiClient.defaults.headers.common['Authorization'];
            console.log('Заголовок авторизации удален (нет TMA данных и сохраненного токена)');
        }
    }
};

// API для работы с аутентификацией
const auth = {
    // Аутентификация через Telegram Mini App
    telegramAuth: async (data) => {
        console.log('Отправка запроса на аутентификацию:', data);
        // Создаем отдельный экземпляр axios для запроса аутентификации
        const authClient = axios.create({
            baseURL: process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1',
            headers: {
                'Content-Type': 'application/json',
                'Accept': 'application/json',
                'X-Requested-With': 'XMLHttpRequest'
            },
            withCredentials: false
        });

        try {
            const response = await authClient.post('/auth/telegram', data);
            console.log('Успешный ответ при аутентификации:', {
                status: response.status,
                data: response.data
            });
            return response;
        } catch (error) {
            console.error('Ошибка при аутентификации:', {
                message: error.message,
                response: error.response?.data,
                status: error.response?.status
            });
            throw error;
        }
    },

    // Проверка валидности токена
    verifyToken: (data) => {
        return apiClient.post('/auth/verify', data);
    },

    // Выход из аккаунта
    logout: (data) => {
        return apiClient.post('/auth/logout', data);
    },
};

// API для работы с пользователями
const user = {
    // Получение данных текущего пользователя
    getCurrent: () => {
        return apiClient.get('/users/me');
    },

    // Получение баланса пользователя
    getBalance: () => {
        return apiClient.get('/users/balance');
    },

    // Запрос на вывод средств
    requestWithdraw: (data) => {
        return apiClient.post('/users/withdraw', data);
    },

    // Получение истории выводов
    getWithdrawHistory: (params) => {
        return apiClient.get('/users/withdrawals', { params });
    },

    // Генерация адреса кошелька
    generateWallet: () => {
        return apiClient.post('/users/wallet');
    },
};

// API для работы с играми
const game = {
    // Создание новой игры
    create: (data) => {
        return apiClient.post('/games', data);
    },

    // Получение информации об игре по ID
    get: (id) => {
        return apiClient.get(`/games/${id}`);
    },

    // Получение списка активных игр
    getActive: (params) => {
        return apiClient.get('/games', { params });
    },

    // Получение списка игр, созданных пользователем
    getMy: (params) => {
        return apiClient.get('/games/my', { params });
    },

    // Удаление игры
    delete: (id) => {
        return apiClient.delete(`/games/${id}`);
    },

    // Пополнение reward pool игры
    addReward: (id, data) => {
        return apiClient.post(`/games/${id}/reward`, data);
    },

    // Активация игры
    activate: (id) => {
        return apiClient.post(`/games/${id}/activate`);
    },

    // Деактивация игры
    deactivate: (id) => {
        return apiClient.post(`/games/${id}/deactivate`);
    },
};

// API для работы с лобби
const lobby = {
    // Присоединение к игре (создание лобби)
    join: async (data) => {
        try {
            // Проверяем наличие токена авторизации в заголовках
            const authHeader = apiClient.defaults.headers.common['Authorization'];
            console.log('Текущий заголовок авторизации:', authHeader ? 'Присутствует' : 'Отсутствует');

            if (!authHeader) {
                // Если токена нет в заголовках, проверяем localStorage
                const token = localStorage.getItem('token');
                console.log('Токен из localStorage:', token ? `${token.substring(0, 15)}...` : 'Отсутствует');

                if (!token) {
                    // Проверяем, находимся ли мы в режиме разработки
                    if (isDevelopmentMode()) {
                        console.log('Режим разработки: используем тестовый токен');
                        // В режиме разработки используем тестовый токен
                        const devToken = 'dev_test_token_for_development_only';
                        setAuthToken(devToken);
                        console.log('Установлен тестовый токен для режима разработки');
                    } else {
                        throw new Error('Требуется авторизация');
                    }
                } else {
                    // Устанавливаем токен в заголовки запроса
                    setAuthToken(token);
                    console.log('Токен установлен из localStorage для запроса на создание лобби');
                }
            } else {
                console.log('Токен авторизации уже присутствует в заголовках');
            }

            // Проверяем наличие необходимых данных
            if (!data.game_id) {
                throw new Error('ID игры не указан');
            }
            if (!data.bet_amount || data.bet_amount <= 0) {
                throw new Error('Некорректная сумма ставки');
            }

            const requestData = {
                game_id: data.game_id,
                bet_amount: data.bet_amount
            };

            // Если мы в режиме разработки и используем тестовый токен, добавляем флаг dev_mode
            if (isUsingDevToken()) {
                console.log('Добавление флага dev_mode в запрос');
                requestData.dev_mode = true;
            }

            // Повторно проверяем заголовки перед отправкой запроса
            const finalAuthHeader = apiClient.defaults.headers.common['Authorization'];
            console.log('Финальный заголовок авторизации перед отправкой:',
                finalAuthHeader ? `${finalAuthHeader.substring(0, 20)}...` : 'Отсутствует');

            console.log('Подготовка запроса на создание лобби:', {
                url: '/lobbies',
                method: 'POST',
                data: requestData,
                headers: {
                    ...apiClient.defaults.headers,
                    Authorization: apiClient.defaults.headers.common['Authorization'] ?
                        apiClient.defaults.headers.common['Authorization'].substring(0, 20) + '...' : undefined
                }
            });

            const response = await apiClient.post('/lobbies', requestData);

            console.log('Успешный ответ сервера:', {
                status: response.status,
                data: response.data,
                headers: response.headers
            });

            // Проверяем формат ответа и нормализуем его
            if (response.data && !response.data.id && response.data.lobby) {
                // Если сервер вернул данные в формате { lobby: { id: ... } }
                response.data = response.data.lobby;
            }

            return response;
        } catch (error) {
            console.error('Ошибка при создании лобби:', {
                message: error.message,
                request: {
                    url: error.config?.url,
                    method: error.config?.method,
                    data: JSON.parse(error.config?.data || '{}'),
                    headers: {
                        ...error.config?.headers,
                        Authorization: error.config?.headers?.Authorization ?
                            error.config?.headers?.Authorization.substring(0, 20) + '...' : undefined
                    }
                },
                response: {
                    status: error.response?.status,
                    statusText: error.response?.statusText,
                    data: error.response?.data,
                    headers: error.response?.headers
                }
            });

            // Если это ошибка 500 и сообщение об отсутствии активного лобби,
            // пробуем создать лобби еще раз
            if (error.response?.status === 500 &&
                error.response?.data?.error?.includes('active lobby not found')) {
                console.log('Повторная попытка создания лобби...');
                try {
                    // Используем те же данные, что и в первом запросе
                    const retryData = {
                        game_id: data.game_id,
                        bet_amount: data.bet_amount
                    };
                    const retryResponse = await apiClient.post('/lobbies', retryData);
                    console.log('Успешный ответ сервера при повторной попытке:', {
                        status: retryResponse.status,
                        data: retryResponse.data,
                        headers: retryResponse.headers
                    });
                    return retryResponse;
                } catch (retryError) {
                    console.error('Ошибка при повторной попытке создания лобби:', {
                        status: retryError.response?.status,
                        data: retryError.response?.data,
                        message: retryError.message,
                        request: {
                            url: retryError.config?.url,
                            method: retryError.config?.method,
                            data: JSON.parse(retryError.config?.data || '{}'),
                            headers: {
                                ...retryError.config?.headers,
                                Authorization: retryError.config?.headers?.Authorization ?
                                    retryError.config?.headers?.Authorization.substring(0, 20) + '...' : undefined
                            }
                        },
                        response: {
                            status: retryError.response?.status,
                            statusText: retryError.response?.statusText,
                            data: retryError.response?.data,
                            headers: retryError.response?.headers
                        }
                    });
                    throw retryError;
                }
            }

            throw error;
        }
    },

    // Получение информации о лобби по ID
    get: (id) => {
        return apiClient.get(`/lobbies/${id}`);
    },

    // Получение активного лобби пользователя
    getActive: () => {
        return apiClient.get('/lobbies/active');
    },

    // Получение всех лобби пользователя
    getAll: (params) => {
        return apiClient.get('/lobbies', { params });
    },

    // Отправка попытки угадать слово
    makeAttempt: (id, data) => {
        return apiClient.post(`/lobbies/${id}/attempt`, data);
    },

    // Получение истории попыток
    getAttempts: (id) => {
        return apiClient.get(`/lobbies/${id}/attempts`);
    },

    // Продление времени лобби
    extendTime: (id, data) => {
        return apiClient.post(`/lobbies/${id}/extend`, data);
    },
};

// API для работы с транзакциями
const transaction = {
    // Получение всех транзакций пользователя
    getAll: (params) => {
        return apiClient.get('/transactions', { params });
    },

    // Получение транзакции по ID
    get: (id) => {
        return apiClient.get(`/transactions/${id}`);
    },

    // Создание депозита
    createDeposit: (data) => {
        return apiClient.post('/transactions/deposit', data);
    },

    // Проверка статуса депозита
    verifyDeposit: (data) => {
        return apiClient.post('/transactions/verify', data);
    },
};

// Собираем все API методы в один объект
const apiService = {
    apiClient,
    setAuthToken,
    setTelegramInitData,
    isDevelopmentMode,
    isUsingDevToken,
    auth,
    user,
    game,
    lobby,
    transaction,
};

export default apiService; 