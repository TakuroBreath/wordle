-- Миграция для интеграции с TON блокчейном

-- Обновление таблицы users
ALTER TABLE users 
    ADD COLUMN IF NOT EXISTS pending_withdrawal DECIMAL(18, 6) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS withdrawal_lock_until TIMESTAMP WITH TIME ZONE,
    ADD COLUMN IF NOT EXISTS total_deposited DECIMAL(18, 6) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS total_withdrawn DECIMAL(18, 6) DEFAULT 0;

-- Обновление таблицы games
ALTER TABLE games
    ADD COLUMN IF NOT EXISTS short_id VARCHAR(10) UNIQUE,
    ADD COLUMN IF NOT EXISTS time_limit INTEGER DEFAULT 5,
    ADD COLUMN IF NOT EXISTS deposit_amount DECIMAL(18, 6) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS reserved_amount DECIMAL(18, 6) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS deposit_tx_hash VARCHAR(255);

-- Обновляем статус игры для поддержки pending
ALTER TABLE games DROP CONSTRAINT IF EXISTS check_game_status;
ALTER TABLE games ADD CONSTRAINT check_game_status CHECK (status IN ('pending', 'active', 'inactive', 'closed'));

-- Обновление таблицы lobbies
ALTER TABLE lobbies
    ADD COLUMN IF NOT EXISTS game_short_id VARCHAR(10),
    ADD COLUMN IF NOT EXISTS payment_tx_hash VARCHAR(255),
    ADD COLUMN IF NOT EXISTS currency VARCHAR(10),
    ADD COLUMN IF NOT EXISTS started_at TIMESTAMP WITH TIME ZONE;

-- Обновляем статус лобби
ALTER TABLE lobbies DROP CONSTRAINT IF EXISTS check_lobby_status;
ALTER TABLE lobbies ADD CONSTRAINT check_lobby_status CHECK (status IN ('pending', 'active', 'success', 'failed_tries', 'failed_expired', 'failed_internal', 'canceled'));

-- Обновление таблицы transactions
ALTER TABLE transactions
    ADD COLUMN IF NOT EXISTS fee DECIMAL(18, 6) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS blockchain_lt BIGINT,
    ADD COLUMN IF NOT EXISTS from_address VARCHAR(255),
    ADD COLUMN IF NOT EXISTS to_address VARCHAR(255),
    ADD COLUMN IF NOT EXISTS comment VARCHAR(255),
    ADD COLUMN IF NOT EXISTS game_short_id VARCHAR(10),
    ADD COLUMN IF NOT EXISTS confirmations INTEGER DEFAULT 0,
    ADD COLUMN IF NOT EXISTS processed_at TIMESTAMP WITH TIME ZONE;

-- Обновляем типы транзакций
ALTER TABLE transactions DROP CONSTRAINT IF EXISTS check_transaction_type;
ALTER TABLE transactions ADD CONSTRAINT check_transaction_type CHECK (type IN ('deposit', 'withdraw', 'reward', 'bet', 'commission', 'refund', 'game_deposit', 'game_refund', 'reserve', 'release_reserve'));

-- Обновляем статусы транзакций
ALTER TABLE transactions DROP CONSTRAINT IF EXISTS check_transaction_status;
ALTER TABLE transactions ADD CONSTRAINT check_transaction_status CHECK (status IN ('pending', 'confirming', 'completed', 'failed', 'canceled'));

-- Создаем таблицу для хранения состояния воркера
CREATE TABLE IF NOT EXISTS blockchain_state (
    id INTEGER PRIMARY KEY DEFAULT 1,
    last_processed_lt BIGINT DEFAULT 0,
    last_processed_hash VARCHAR(255),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT single_row CHECK (id = 1)
);

-- Вставляем начальную запись состояния
INSERT INTO blockchain_state (id, last_processed_lt, last_processed_hash)
VALUES (1, 0, '')
ON CONFLICT (id) DO NOTHING;

-- Создаем таблицу для платежных ссылок (pending payments)
CREATE TABLE IF NOT EXISTS pending_payments (
    id UUID PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(telegram_id),
    game_id UUID REFERENCES games(id),
    game_short_id VARCHAR(10),
    lobby_id UUID REFERENCES lobbies(id),
    payment_type VARCHAR(20) NOT NULL, -- 'game_deposit', 'lobby_bet'
    amount DECIMAL(18, 6) NOT NULL,
    currency VARCHAR(10) NOT NULL,
    comment VARCHAR(255) NOT NULL UNIQUE, -- Уникальный комментарий для идентификации
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT check_payment_type CHECK (payment_type IN ('game_deposit', 'lobby_bet')),
    CONSTRAINT check_payment_status CHECK (status IN ('pending', 'completed', 'expired', 'canceled'))
);

-- Индексы для pending_payments
CREATE INDEX IF NOT EXISTS idx_pending_payments_comment ON pending_payments(comment);
CREATE INDEX IF NOT EXISTS idx_pending_payments_user_id ON pending_payments(user_id);
CREATE INDEX IF NOT EXISTS idx_pending_payments_status ON pending_payments(status);
CREATE INDEX IF NOT EXISTS idx_pending_payments_expires_at ON pending_payments(expires_at);

-- Индекс для games.short_id
CREATE INDEX IF NOT EXISTS idx_games_short_id ON games(short_id);

-- Индекс для transactions
CREATE INDEX IF NOT EXISTS idx_transactions_tx_hash ON transactions(tx_hash);
CREATE INDEX IF NOT EXISTS idx_transactions_blockchain_lt ON transactions(blockchain_lt);
CREATE INDEX IF NOT EXISTS idx_transactions_game_short_id ON transactions(game_short_id);

-- Индекс для lobbies
CREATE INDEX IF NOT EXISTS idx_lobbies_game_short_id ON lobbies(game_short_id);
CREATE INDEX IF NOT EXISTS idx_lobbies_payment_tx_hash ON lobbies(payment_tx_hash);

-- Генерация коротких ID для существующих игр
CREATE OR REPLACE FUNCTION generate_short_id() RETURNS VARCHAR(8) AS $$
DECLARE
    chars VARCHAR(36) := 'ABCDEFGHJKLMNPQRSTUVWXYZ23456789';
    result VARCHAR(8) := '';
    i INTEGER;
BEGIN
    FOR i IN 1..8 LOOP
        result := result || substr(chars, floor(random() * length(chars) + 1)::integer, 1);
    END LOOP;
    RETURN result;
END;
$$ LANGUAGE plpgsql;

-- Обновляем существующие игры с short_id
UPDATE games SET short_id = generate_short_id() WHERE short_id IS NULL;

-- Делаем short_id NOT NULL после заполнения
ALTER TABLE games ALTER COLUMN short_id SET NOT NULL;
