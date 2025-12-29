-- Откат миграции для интеграции с TON блокчейном

-- Удаляем таблицу pending_payments
DROP TABLE IF EXISTS pending_payments;

-- Удаляем таблицу blockchain_state
DROP TABLE IF EXISTS blockchain_state;

-- Удаляем функцию генерации short_id
DROP FUNCTION IF EXISTS generate_short_id();

-- Откат изменений в transactions
ALTER TABLE transactions DROP CONSTRAINT IF EXISTS check_transaction_type;
ALTER TABLE transactions ADD CONSTRAINT check_transaction_type CHECK (type IN ('deposit', 'withdraw', 'reward', 'bet'));

ALTER TABLE transactions DROP CONSTRAINT IF EXISTS check_transaction_status;
ALTER TABLE transactions ADD CONSTRAINT check_transaction_status CHECK (status IN ('pending', 'completed', 'failed'));

ALTER TABLE transactions
    DROP COLUMN IF EXISTS fee,
    DROP COLUMN IF EXISTS blockchain_lt,
    DROP COLUMN IF EXISTS from_address,
    DROP COLUMN IF EXISTS to_address,
    DROP COLUMN IF EXISTS comment,
    DROP COLUMN IF EXISTS game_short_id,
    DROP COLUMN IF EXISTS confirmations,
    DROP COLUMN IF EXISTS processed_at;

-- Откат изменений в lobbies
ALTER TABLE lobbies DROP CONSTRAINT IF EXISTS check_lobby_status;

ALTER TABLE lobbies
    DROP COLUMN IF EXISTS game_short_id,
    DROP COLUMN IF EXISTS payment_tx_hash,
    DROP COLUMN IF EXISTS currency,
    DROP COLUMN IF EXISTS started_at;

-- Откат изменений в games
ALTER TABLE games DROP CONSTRAINT IF EXISTS check_game_status;
ALTER TABLE games ADD CONSTRAINT check_game_status CHECK (status IN ('active', 'inactive'));

ALTER TABLE games
    DROP COLUMN IF EXISTS short_id,
    DROP COLUMN IF EXISTS time_limit,
    DROP COLUMN IF EXISTS deposit_amount,
    DROP COLUMN IF EXISTS reserved_amount,
    DROP COLUMN IF EXISTS deposit_tx_hash;

-- Откат изменений в users
ALTER TABLE users
    DROP COLUMN IF EXISTS pending_withdrawal,
    DROP COLUMN IF EXISTS withdrawal_lock_until,
    DROP COLUMN IF EXISTS total_deposited,
    DROP COLUMN IF EXISTS total_withdrawn;

-- Удаляем индексы
DROP INDEX IF EXISTS idx_pending_payments_comment;
DROP INDEX IF EXISTS idx_pending_payments_user_id;
DROP INDEX IF EXISTS idx_pending_payments_status;
DROP INDEX IF EXISTS idx_pending_payments_expires_at;
DROP INDEX IF EXISTS idx_games_short_id;
DROP INDEX IF EXISTS idx_transactions_tx_hash;
DROP INDEX IF EXISTS idx_transactions_blockchain_lt;
DROP INDEX IF EXISTS idx_transactions_game_short_id;
DROP INDEX IF EXISTS idx_lobbies_game_short_id;
DROP INDEX IF EXISTS idx_lobbies_payment_tx_hash;
