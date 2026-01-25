-- Миграция для поддержки Base Mini App и USDC

-- Обновление таблицы users
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS auth_provider VARCHAR(32) NOT NULL DEFAULT 'telegram',
    ADD COLUMN IF NOT EXISTS base_fid BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS base_username VARCHAR(255) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS base_display_name VARCHAR(255) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS base_pfp_url TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS base_custody_address VARCHAR(255) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS base_smart_wallet_address VARCHAR(255) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS balance_usdc DECIMAL(18, 6) NOT NULL DEFAULT 0;

-- Ограничение на провайдера авторизации
ALTER TABLE users DROP CONSTRAINT IF EXISTS check_users_auth_provider;
ALTER TABLE users ADD CONSTRAINT check_users_auth_provider CHECK (auth_provider IN ('telegram', 'base'));

-- Уникальный индекс на Base FID (только если указан)
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_base_fid ON users(base_fid) WHERE base_fid > 0;

-- Обновление ограничений валют
ALTER TABLE games DROP CONSTRAINT IF EXISTS check_currency;
ALTER TABLE games ADD CONSTRAINT check_currency CHECK (currency IN ('TON', 'USDT', 'USDC'));

ALTER TABLE transactions DROP CONSTRAINT IF EXISTS check_transaction_currency;
ALTER TABLE transactions ADD CONSTRAINT check_transaction_currency CHECK (currency IN ('TON', 'USDT', 'USDC'));
