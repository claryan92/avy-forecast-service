CREATE TABLE IF NOT EXISTS forecast_cache (
    zone_id TEXT PRIMARY KEY,
    last_issued TIMESTAMP NOT NULL
);