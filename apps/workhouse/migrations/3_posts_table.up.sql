CREATE TABLE posts (
  id TEXT PRIMARY KEY,
  title TEXT UNIQUE NOT NULL,
  content TEXT NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);