CREATE TABLE hosts (
  id          TEXT PRIMARY KEY,
  owner_id    TEXT NOT NULL,
  server_id   TEXT NOT NULL,
  payment_id  TEXT NOT NULL,
  name        TEXT NOT NULL,
  domain_name TEXT NOT NULL,
  status      TEXT DEFAULT 'starting',
  error       TEXT DEFAULT '',
  created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);