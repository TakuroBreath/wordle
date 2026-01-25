-- Откат поддержки Base Mini App и USDC

DROP INDEX IF EXISTS idx_users_base_fid;

ALTER TABLE users DROP CONSTRAINT IF EXISTS check_users_auth_provider;

ALTER TABLE users
    DROP COLUMN IF EXISTS balance_usdc,
    DROP COLUMN IF EXISTS base_smart_wallet_address,
    DROP COLUMN IF EXISTS base_custody_address,
    DROP COLUMN IF EXISTS base_pfp_url,
    DROP COLUMN IF EXISTS base_display_name,
    DROP COLUMN IF EXISTS base_username,
    DROP COLUMN IF EXISTS base_fid,
    DROP COLUMN IF EXISTS auth_provider;

ALTER TABLE games DROP CONSTRAINT IF EXISTS check_currency;
ALTER TABLE games ADD CONSTRAINT check_currency CHECK (currency IN ('TON', 'USDT'));

ALTER TABLE transactions DROP CONSTRAINT IF EXISTS check_transaction_currency;
ALTER TABLE transactions ADD CONSTRAINT check_transaction_currency CHECK (currency IN ('TON', 'USDT'));
