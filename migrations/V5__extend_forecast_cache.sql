-- Extend forecast_cache with timestamps and trigger (previously attempted in modified V4)
-- Safe additive migration

ALTER TABLE forecast_cache
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ALTER COLUMN last_issued TYPE TIMESTAMPTZ USING last_issued AT TIME ZONE 'UTC';

CREATE OR REPLACE FUNCTION set_updated_at_forecast_cache()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS forecast_cache_set_updated_at ON forecast_cache;
CREATE TRIGGER forecast_cache_set_updated_at
BEFORE UPDATE ON forecast_cache
FOR EACH ROW EXECUTE FUNCTION set_updated_at_forecast_cache();
