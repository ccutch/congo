CREATE TABLE agents (
    id            TEXT PRIMARY KEY,
    name          TEXT NOT NULL UNIQUE,
    model         TEXT NOT NULL,
    system_prompt TEXT NOT NULL,
    created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);