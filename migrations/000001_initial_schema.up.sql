-- Миграция для создания начальной схемы базы данных

-- Создание таблицы пользователей
CREATE TABLE IF NOT EXISTS users (
    telegram_id BIGINT PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255),
    wallet VARCHAR(255),
    balance_ton DECIMAL(18, 6) DEFAULT 0,
    balance_usdt DECIMAL(18, 6) DEFAULT 0,
    wins INTEGER DEFAULT 0,
    losses INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Создание таблицы игр
CREATE TABLE IF NOT EXISTS games (
    id UUID PRIMARY KEY,
    creator_id BIGINT NOT NULL REFERENCES users(telegram_id),
    word VARCHAR(50) NOT NULL,
    length INTEGER NOT NULL,
    difficulty VARCHAR(20) NOT NULL,
    max_tries INTEGER NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    min_bet DECIMAL(18, 6) NOT NULL,
    max_bet DECIMAL(18, 6) NOT NULL,
    reward_multiplier DECIMAL(8, 2) NOT NULL,
    currency VARCHAR(10) NOT NULL, -- "TON" или "USDT"
    reward_pool_ton DECIMAL(18, 6) DEFAULT 0,
    reward_pool_usdt DECIMAL(18, 6) DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- "active" или "inactive"
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT check_currency CHECK (currency IN ('TON', 'USDT'))
);

-- Создание таблицы лобби
CREATE TABLE IF NOT EXISTS lobbies (
    id UUID PRIMARY KEY,
    game_id UUID NOT NULL REFERENCES games(id),
    user_id BIGINT NOT NULL REFERENCES users(telegram_id),
    max_tries INTEGER NOT NULL,
    tries_used INTEGER DEFAULT 0,
    bet_amount DECIMAL(18, 6) NOT NULL,
    potential_reward DECIMAL(18, 6) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- "active", "success", "failed"
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Создание таблицы попыток
CREATE TABLE IF NOT EXISTS attempts (
    id UUID PRIMARY KEY,
    lobby_id UUID NOT NULL REFERENCES lobbies(id),
    game_id UUID NOT NULL REFERENCES games(id),
    user_id BIGINT NOT NULL REFERENCES users(telegram_id),
    word VARCHAR(50) NOT NULL,
    result INTEGER[] NOT NULL, -- Массив результатов (0, 1, 2) для каждой буквы
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Создание таблицы истории игр
CREATE TABLE IF NOT EXISTS history (
    id UUID PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(telegram_id),
    game_id UUID NOT NULL REFERENCES games(id),
    lobby_id UUID NOT NULL REFERENCES lobbies(id),
    status VARCHAR(20) NOT NULL, -- "success" или "failed"
    reward DECIMAL(18, 6) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Создание таблицы транзакций
CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(telegram_id),
    amount DECIMAL(18, 6) NOT NULL,
    currency VARCHAR(10) NOT NULL, -- "TON" или "USDT"
    type VARCHAR(20) NOT NULL, -- "deposit", "withdraw", "reward", "bet"
    status VARCHAR(20) NOT NULL, -- "pending", "completed", "failed"
    tx_hash VARCHAR(255),
    wallet_address VARCHAR(255),
    network VARCHAR(20),
    game_id UUID REFERENCES games(id),
    lobby_id UUID REFERENCES lobbies(id),
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT check_transaction_currency CHECK (currency IN ('TON', 'USDT')),
    CONSTRAINT check_transaction_type CHECK (type IN ('deposit', 'withdraw', 'reward', 'bet')),
    CONSTRAINT check_transaction_status CHECK (status IN ('pending', 'completed', 'failed'))
);

-- Создание индексов для оптимизации запросов
CREATE INDEX idx_games_creator_id ON games(creator_id);
CREATE INDEX idx_games_status ON games(status);
CREATE INDEX idx_lobbies_game_id ON lobbies(game_id);
CREATE INDEX idx_lobbies_user_id ON lobbies(user_id);
CREATE INDEX idx_lobbies_status ON lobbies(status);
CREATE INDEX idx_attempts_lobby_id ON attempts(lobby_id);
CREATE INDEX idx_attempts_game_id ON attempts(game_id);
CREATE INDEX idx_attempts_user_id ON attempts(user_id);
CREATE INDEX idx_history_user_id ON history(user_id);
CREATE INDEX idx_history_game_id ON history(game_id);
CREATE INDEX idx_history_lobby_id ON history(lobby_id);
CREATE INDEX idx_transactions_user_id ON transactions(user_id);
CREATE INDEX idx_transactions_status ON transactions(status);
CREATE INDEX idx_transactions_type ON transactions(type);