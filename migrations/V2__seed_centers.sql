INSERT INTO avalanche_centers (center_id, name, base_url)
VALUES
  ('IPAC', 'Idaho Panhandle Avalanche Center', 'https://www.idahopanhandleavalanche.org/'),
  ('NWAC', 'Northwest Avalanche Center', 'https://nwac.us/')
ON CONFLICT (center_id) DO NOTHING;
