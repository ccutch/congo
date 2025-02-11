CREATE TABLE servers (
  id           TEXT PRIMARY KEY,
  user_id      TEXT NOT NULL,
  host_id      TEXT NOT NULL,
  checkout_id  TEXT DEFAULT '',
  checkout_url TEXT DEFAULT '',
  name         TEXT NOT NULL,
  size         TEXT NOT NULL,
  status       TEXT DEFAULT 'created',
  ip_addr      TEXT DEFAULT '',
  domain       TEXT DEFAULT '',
  error        TEXT DEFAULT '',
  created_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);