CREATE TABLE domains (
    id          TEXT PRIMARY KEY,
    server_id   TEXT REFERENCES servers (id),
    domain_name TEXT UNIQUE NOT NULL,
    verified    BOOLEAN DEFAULT FALSE,
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP 
);