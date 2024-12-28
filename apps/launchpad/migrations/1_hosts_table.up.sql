CREATE TABLE hosts (
    -- columns provided by congo.Model
    id         TEXT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- columns matching our Post model
    owner_id  TEXT NOT NULL,
    name      TEXT NOT NULL,
    size      TEXT NOT NULL CHECK (size IN ('SM', 'MD', 'LG', 'XL')),
    region    TEXT NOT NULL CHECK (region IN ('sfo2', 'nyc2')),
    error     TEXT DEFAULT '',
    ip_addr   TEXT DEFAULT '',
    domain    TEXT DEFAULT ''
);