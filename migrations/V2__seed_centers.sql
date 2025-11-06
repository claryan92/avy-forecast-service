INSERT INTO avalanche_centers (center_id, name, base_url)
VALUES
  ('IPAC', 'Intermountain Avalanche Center', 'https://api.avalanche.org/v2/public'),
  ('NWAC', 'Northwest Avalanche Center', 'https://api.avalanche.org/v2/public')
ON CONFLICT (center_id) DO NOTHING;
