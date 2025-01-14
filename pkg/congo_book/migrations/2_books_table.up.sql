CREATE TABLE books (
    id          TEXT PRIMARY KEY,
    library_id  TEXT REFERENCES libraries (id),
    public      BOOLEAN DEFAULT FALSE,
    title       TEXT NOT NULL,
    author      TEXT NOT NULL,
    publish_at  TIMESTAMP,
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);