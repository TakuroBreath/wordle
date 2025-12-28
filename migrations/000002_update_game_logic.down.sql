ALTER TABLE games DROP CONSTRAINT IF EXISTS games_short_id_unique;
ALTER TABLE games DROP COLUMN IF NOT EXISTS short_id;
ALTER TABLE games DROP COLUMN IF NOT EXISTS deposit_amount;
ALTER TABLE games DROP COLUMN IF NOT EXISTS commission_rate;
ALTER TABLE games DROP COLUMN IF NOT EXISTS time_limit_minutes;

ALTER TABLE transactions DROP COLUMN IF NOT EXISTS comment;

-- Revert constraints (simplified)
ALTER TABLE games DROP CONSTRAINT IF EXISTS games_status_check;
ALTER TABLE transactions DROP CONSTRAINT IF EXISTS check_transaction_type;
