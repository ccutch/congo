CREATE TABLE workspaces (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    port        INTEGER NOT NULL,
    image       TEXT NOT NULL,
    tag         TEXT NOT NULL,
    ready       BOOLEAN DEFAULT FALSE,
    repo_id     TEXT REFERENCES repositories (id),
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);