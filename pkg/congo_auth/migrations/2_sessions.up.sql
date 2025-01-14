CREATE TABLE sessions (
    id          TEXT PRIMARY KEY,
    identity_id TEXT REFERENCES identities (id),
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);