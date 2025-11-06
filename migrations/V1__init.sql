CREATE TABLE IF NOT EXISTS avalanche_centers (
    id SERIAL PRIMARY KEY,
    center_id TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    base_url TEXT NOT NULL,
    active BOOLEAN DEFAULT TRUE
);
