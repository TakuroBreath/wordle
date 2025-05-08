import axios from 'axios';

// Создаем экземпляр axios с базовым URL для API
const apiClient = axios.create({
    baseURL: process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1',
    headers: {
        'Content-Type': 'application/json',
    },
});

// Функция для установки токена авторизации
const setAuthToken = (token) => {
    if (token) {
        apiClient.defaults.headers.common['Authorization'] = `Bearer ${token}`;
    } else {
        delete apiClient.defaults.headers.common['Authorization'];
    }
};

// Функция для установки инициализационных данных Telegram Mini App
const setTelegramInitData = (initData) => {
    if (initData) {
        apiClient.defaults.headers.common['Authorization'] = `tma ${initData}`;
    } else {
        delete apiClient.defaults.headers.common['Authorization'];
    }
};

// API для работы с аутентификацией
const auth = {
    // Аутентификация через Telegram Mini App
    telegramAuth: (data) => {
        return apiClient.post('/auth/telegram', data);
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
    join: (data) => {
        return apiClient.post('/lobbies', data);
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

export default {
    setAuthToken,
    setTelegramInitData,
    auth,
    user,
    game,
    lobby,
    transaction,
}; 