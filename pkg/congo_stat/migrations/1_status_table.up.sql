CREATE TABLE system_status (
    id           TEXT PRIMARY KEY,
    cpu_usage    INTEGER NOT NULL,
    ram_usage    INTEGER NOT NULL,
    storage_used INTEGER NOT NULL,
    volume_used  INTEGER NOT NULL,
    created_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);