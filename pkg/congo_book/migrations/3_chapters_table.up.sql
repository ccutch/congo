CREATE TABLE chatpers (
    id         TEXT PRIMARY KEY,
    book_id    TEXT REFERENCES books (id),
    number     INTEGER NOT NULL,
    title      TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (book_id, number)
);