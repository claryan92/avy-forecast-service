-- Undo V5__extend_forecast_cache: drop added columns and trigger, revert last_issued to TIMESTAMP

DROP TRIGGER IF EXISTS forecast_cache_set_updated_at ON forecast_cache;
DROP FUNCTION IF EXISTS set_updated_at_forecast_cache();

ALTER TABLE forecast_cache
    ALTER COLUMN last_issued TYPE TIMESTAMP USING last_issued,
    DROP COLUMN IF EXISTS updated_at,
    DROP COLUMN IF EXISTS created_at;