-- Add new columns to games table
ALTER TABLE games ADD COLUMN IF NOT EXISTS short_id VARCHAR(20);
ALTER TABLE games ADD COLUMN IF NOT EXISTS deposit_amount DECIMAL(18, 6);
ALTER TABLE games ADD COLUMN IF NOT EXISTS commission_rate DECIMAL(5, 2) DEFAULT 5.0;
ALTER TABLE games ADD COLUMN IF NOT EXISTS time_limit_minutes INTEGER DEFAULT 60;
ALTER TABLE games ADD CONSTRAINT games_short_id_unique UNIQUE (short_id);

-- Add comment column to transactions
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS comment VARCHAR(255);

-- Update status check constraint if needed (postgres usually allows new values if not restricted by enum type, here it is a VARCHAR with check constraint usually, but the original schema used VARCHAR with CHECK)
-- Dropping old constraint and adding new one to support 'pending_activation'
ALTER TABLE games DROP CONSTRAINT IF EXISTS games_status_check;
ALTER TABLE games ADD CONSTRAINT games_status_check CHECK (status IN ('active', 'inactive', 'pending_activation', 'finished'));

-- Update transaction type check
ALTER TABLE transactions DROP CONSTRAINT IF EXISTS check_transaction_type;
ALTER TABLE transactions ADD CONSTRAINT check_transaction_type CHECK (type IN ('deposit', 'withdraw', 'reward', 'bet', 'refund', 'commission'));
