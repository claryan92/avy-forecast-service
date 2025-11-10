-- Undo V6__add_timestamps_to_subscriptions
DROP TRIGGER IF EXISTS subscriptions_set_updated_at ON subscriptions;
DROP FUNCTION IF EXISTS set_updated_at_subscriptions();
ALTER TABLE subscriptions
    DROP COLUMN IF EXISTS updated_at,
    DROP COLUMN IF EXISTS created_at;