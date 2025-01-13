CREATE TABLE libraries (
    id          TEXT PRIMARY KEY,
    owner_id    TEXT NOT NULL,
    public      BOOLEAN DEFAULT FALSE,
    name        TEXT NOT NULL,
    description TEXT NOT NULL,
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);