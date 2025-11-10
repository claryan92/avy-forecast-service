CREATE TABLE IF NOT EXISTS subscriptions (
    id SERIAL PRIMARY KEY,
    zone_id TEXT NOT NULL,
    email TEXT NOT NULL,
    last_notified TIMESTAMP DEFAULT NULL
);
