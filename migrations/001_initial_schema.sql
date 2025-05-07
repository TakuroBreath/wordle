-- Миграция для создания начальной схемы базы данных

-- Создание таблицы пользователей
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    telegram_id BIGINT UNIQUE NOT NULL,
    username VARCHAR(255) NOT NULL,
    wallet VARCHAR(255),
    balance DECIMAL(18, 6) DEFAULT 0,
    wins INTEGER DEFAULT 0,
    losses INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Создание таблицы игр
CREATE TABLE IF NOT EXISTS games (
    id UUID PRIMARY KEY,
    title VARCHAR(255),
    word VARCHAR(50) NOT NULL,
    length INTEGER NOT NULL,
    creator_id BIGINT NOT NULL REFERENCES users(id),
    difficulty INTEGER NOT NULL,
    max_tries INTEGER NOT NULL,
    reward_multiplier DECIMAL(8, 2) NOT NULL,
    currency VARCHAR(10) NOT NULL, -- "TON" или "USDT"
    prize_pool DECIMAL(18, 6) NOT NULL,
    min_bet DECIMAL(18, 6) NOT NULL,
    max_bet DECIMAL(18, 6) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- "active" или "inactive"
    CONSTRAINT check_currency CHECK (currency IN ('TON', 'USDT'))
);

-- Создание таблицы лобби
CREATE TABLE IF NOT EXISTS lobbies (
    id UUID PRIMARY KEY,
    game_id UUID NOT NULL REFERENCES games(id),
    user_id BIGINT NOT NULL REFERENCES users(id),
    max_tries INTEGER NOT NULL,
    tries_used INTEGER DEFAULT 0,
    bet DECIMAL(18, 6) NOT NULL,
    potential_reward DECIMAL(18, 6) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- "active" или "inactive"
    CONSTRAINT unique_game_user UNIQUE (game_id, user_id)
);

-- Создание таблицы попыток
CREATE TABLE IF NOT EXISTS attempts (
    id UUID PRIMARY KEY,
    lobby_id UUID NOT NULL REFERENCES lobbies(id),
    word VARCHAR(50) NOT NULL,
    feedback TEXT NOT NULL, -- Результат проверки
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Создание таблицы истории игр
CREATE TABLE IF NOT EXISTS history (
    id UUID PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    game_id UUID NOT NULL REFERENCES games(id),
    lobby_id UUID NOT NULL REFERENCES lobbies(id),
    status VARCHAR(20) NOT NULL, -- "creator_win" или "player_win"
    reward DECIMAL(18, 6) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Создание таблицы транзакций
CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    amount DECIMAL(18, 6) NOT NULL,
    currency VARCHAR(10) NOT NULL,
    type VARCHAR(20) NOT NULL, -- "deposit", "withdraw", "win", "loss"
    status VARCHAR(20) NOT NULL, -- "pending", "completed", "failed"
    tx_hash VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT check_transaction_currency CHECK (currency IN ('TON', 'USDT')),
    CONSTRAINT check_transaction_type CHECK (type IN ('deposit', 'withdraw', 'win', 'loss')),
    CONSTRAINT check_transaction_status CHECK (status IN ('pending', 'completed', 'failed'))
);

-- Создание индексов для оптимизации запросов
CREATE INDEX idx_games_creator_id ON games(creator_id);
CREATE INDEX idx_games_status ON games(status);
CREATE INDEX idx_lobbies_game_id ON lobbies(game_id);
CREATE INDEX idx_lobbies_user_id ON lobbies(user_id);
CREATE INDEX idx_lobbies_status ON lobbies(status);
CREATE INDEX idx_attempts_lobby_id ON attempts(lobby_id);
CREATE INDEX idx_history_user_id ON history(user_id);
CREATE INDEX idx_history_game_id ON history(game_id);
CREATE INDEX idx_transactions_user_id ON transactions(user_id);
CREATE INDEX idx_transactions_status ON transactions(status);