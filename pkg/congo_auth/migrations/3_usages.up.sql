CREATE TABLE usages (
    id          TEXT PRIMARY KEY,
    identity_id TEXT REFERENCES identities (id),
    resource    TEXT NOT NULL,
    allowed     BOOLEAN NOT NULL,
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP 
);