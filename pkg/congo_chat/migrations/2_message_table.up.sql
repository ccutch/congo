CREATE TABLE messages (
    id           TEXT PRIMARY KEY,
    to_mailbox   TEXT NOT NULL REFERENCES mailboxes (id),
    from_mailbox TEXT NOT NULL REFERENCES mailboxes (id),
    content      TEXT NOT NULL,
    created_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);