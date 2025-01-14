CREATE TABLE contents (
    id         TEXT PRIMARY KEY,
    book_id    TEXT REFERENCES books (id),
    chapter_id TEXT REFERENCES chapters (id),
    position   INTEGER NOT NULL,
    media      TEXT NOT NULL,
    content    BLOB NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (book_id, chapter_id, position)
);