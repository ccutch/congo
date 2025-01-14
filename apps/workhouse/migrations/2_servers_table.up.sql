CREATE TABLE servers (
    id          TEXT PRIMARY KEY,
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    name        TEXT UNIQUE NOT NULL,
    domain_name TEXT DEFAULT ''
);