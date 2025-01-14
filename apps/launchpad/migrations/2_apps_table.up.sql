CREATE TABLE apps (
    -- columns provided by congo.Model
    id         TEXT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- columns matching our Post model
    owner_id  TEXT NOT NULL,
    name      TEXT NOT NULL,
    binary    BLOB NOT NULL
);


ALTER TABLE hosts ADD COLUMN app_id TEXT REFERENCES apps (id);