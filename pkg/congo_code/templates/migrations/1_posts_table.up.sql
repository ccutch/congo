CREATE TABLE posts (
    -- columns provided by congo.Model
    id         TEXT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- columns matching our Post model
    title      TEXT NOT NULL UNIQUE,
    content    TEXT NOT NULL
);