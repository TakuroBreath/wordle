import axios from 'axios';

// API базовый URL
const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8000/api/v1';

// Создаем инстанс axios с базовыми настройками
const api = axios.create({
    baseURL: API_BASE_URL,
    headers: {
        'Content-Type': 'application/json',
    },
    // Добавляем таймаут для запросов
    timeout: 10000,
});

// Добавляем интерцептор для авторизации
api.interceptors.request.use((config) => {
    // Получаем токен из localStorage
    const token = localStorage.getItem('token');

    // Если токен есть, добавляем его в заголовок
    if (token) {
        config.headers.Authorization = `Bearer ${token}`;
    } else {
        // Иначе используем Telegram Mini App данные
        const initData = localStorage.getItem('telegram_init_data');
        if (initData) {
            // Добавляем данные Telegram в заголовок в формате tma <token>
            config.headers['Authorization'] = "tma " + initData;
            console.log('Отправляем данные Telegram:', initData);
        } else {
            console.warn('Данные Telegram не найдены в localStorage');
        }
    }

    // Логируем все запросы
    console.log('Отправляем запрос:', {
        url: config.url,
        method: config.method,
        headers: config.headers,
        data: config.data
    });

    return config;
});

// Добавляем интерцептор для обработки ошибок
api.interceptors.response.use(
    (response) => {
        console.log('Получен ответ:', response.data);
        return response;
    },
    async (error) => {
        console.error('Ошибка запроса:', {
            message: error.message,
            code: error.code,
            response: error.response?.data,
            status: error.response?.status
        });

        if (error.code === 'ERR_NETWORK') {
            console.error('Ошибка подключения к серверу. Проверьте, запущен ли бэкенд.');
        }
        return Promise.reject(error);
    }
);

// Функция для повторных попыток запроса
const retryRequest = async (fn, maxRetries = 3, delay = 1000) => {
    for (let i = 0; i < maxRetries; i++) {
        try {
            return await fn();
        } catch (error) {
            console.log(`Попытка ${i + 1} из ${maxRetries} не удалась:`, error.message);
            if (i === maxRetries - 1) throw error;
            await new Promise(resolve => setTimeout(resolve, delay * (i + 1)));
        }
    }
};

// API методы для авторизации
export const authAPI = {
    // Инициализация через Telegram
    telegramAuth: async () => {
        const initData = localStorage.getItem('telegram_init_data');
        if (!initData) {
            throw new Error('No Telegram init data available');
        }
        try {
            console.log('Начинаем авторизацию через Telegram с данными:', initData);
            const response = await retryRequest(() =>
                api.post('/auth/telegram', { init_data: initData })
            );
            if (response.data.token) {
                localStorage.setItem('token', response.data.token);
                console.log('Токен успешно сохранен');
            }
            return response.data;
        } catch (error) {
            console.error('Ошибка авторизации через Telegram:', error);
            throw error;
        }
    },

    // Проверка токена
    verifyAuth: async (token) => {
        return retryRequest(() => api.post('/auth/verify', { token }));
    },

    // Выход
    logout: async () => {
        const token = localStorage.getItem('token');
        if (token) {
            try {
                await retryRequest(() => api.post('/auth/logout', { token }));
            } finally {
                localStorage.removeItem('token');
            }
        }
    }
};

// API методы для пользователя
export const userAPI = {
    // Получение текущего пользователя
    getCurrentUser: async () => {
        return retryRequest(() => api.get('/users/me'));
    },

    // Получение баланса
    getBalance: async () => {
        return retryRequest(() => api.get('/users/balance'));
    },

    // Подключение кошелька
    connectWallet: async (wallet) => {
        return retryRequest(() => api.post('/users/wallet', { wallet }));
    },

    // Получение информации для депозита
    getDepositInfo: async () => {
        return retryRequest(() => api.get('/users/deposit'));
    },

    // Запрос на вывод средств
    requestWithdraw: async (amount, currency, toAddress) => {
        return retryRequest(() => api.post('/users/withdraw', { amount, currency, to_address: toAddress }));
    },

    // История выводов
    getWithdrawHistory: async () => {
        return retryRequest(() => api.get('/users/withdrawals'));
    },

    // История транзакций
    getTransactionHistory: async (limit = 50, offset = 0) => {
        return retryRequest(() => api.get('/users/transactions', { params: { limit, offset } }));
    },

    // Статистика пользователя
    getStats: async () => {
        return retryRequest(() => api.get('/users/stats'));
    }
};

// API методы для игр
export const gameAPI = {
    // Получение всех активных игр
    getAllGames: async () => {
        return retryRequest(() => api.get('/games'));
    },

    // Получение игр пользователя
    getMyGames: async () => {
        return retryRequest(() => api.get('/games/my'));
    },

    // Получение игры по ID или ShortID
    getGame: async (id) => {
        return retryRequest(() => api.get(`/games/${id}`));
    },

    // Создание игры (возвращает payment_info для депозита)
    createGame: async (gameData) => {
        return retryRequest(() => api.post('/games', gameData));
    },

    // Получение платежной информации для игры
    getPaymentInfo: async (id) => {
        return retryRequest(() => api.get(`/games/${id}/payment`));
    },

    // Присоединение к игре (возвращает lobby или payment_info)
    joinGame: async (gameId, betAmount) => {
        return retryRequest(() => api.post('/games/join', { game_id: gameId, bet_amount: betAmount }));
    },

    // Удаление игры
    deleteGame: async (id) => {
        return retryRequest(() => api.delete(`/games/${id}`));
    },

    // Поиск игр
    searchGames: async (params) => {
        return retryRequest(() => api.get('/games/search', { params }));
    },

    // Пополнение reward pool
    addToRewardPool: async (id, amount) => {
        return retryRequest(() => api.post(`/games/${id}/reward`, { amount }));
    },

    // Активация игры
    activateGame: async (id) => {
        return retryRequest(() => api.post(`/games/${id}/activate`));
    },

    // Деактивация игры
    deactivateGame: async (id) => {
        return retryRequest(() => api.post(`/games/${id}/deactivate`));
    }
};

// API методы для лобби
export const lobbyAPI = {
    // Присоединение к игре
    joinGame: async (gameId, betAmount) => {
        return retryRequest(() => api.post('/lobbies', { game_id: gameId, bet_amount: betAmount }));
    },

    // Получение лобби по ID
    getLobby: async (id) => {
        return retryRequest(() => api.get(`/lobbies/${id}`));
    },

    // Получение активного лобби
    getActiveLobby: async () => {
        return retryRequest(() => api.get('/lobbies/active'));
    },

    // Получение всех лобби пользователя
    getUserLobbies: async () => {
        return retryRequest(() => api.get('/lobbies'));
    },

    // Отправка попытки
    makeAttempt: async (lobbyId, word) => {
        return retryRequest(() => api.post(`/lobbies/${lobbyId}/attempt`, { word }));
    },

    // Получение попыток
    getAttempts: async (lobbyId) => {
        return retryRequest(() => api.get(`/lobbies/${lobbyId}/attempts`));
    },

    // Продление времени лобби
    extendLobbyTime: async (lobbyId) => {
        return retryRequest(() => api.post(`/lobbies/${lobbyId}/extend`));
    }
};

// API методы для транзакций
export const transactionAPI = {
    // Получение транзакций пользователя
    getUserTransactions: async () => {
        return api.get('/transactions');
    },

    // Получение транзакции по ID
    getTransaction: async (id) => {
        return api.get(`/transactions/${id}`);
    },

    // Создание депозита
    createDeposit: async (amount, currency) => {
        return api.post('/transactions/deposit', { amount, currency });
    },

    // Проверка депозита
    verifyDeposit: async (txHash) => {
        return api.post('/transactions/verify', { tx_hash: txHash });
    },

    // Получение статистики транзакций
    getTransactionStats: async () => {
        return api.get('/transactions/stats');
    }
};

export default api; 