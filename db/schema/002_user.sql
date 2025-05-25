-- +goose Up
 CREATE TABLE IF NOT EXISTS users (
     id TEXT PRIMARY KEY,
     email TEXT UNIQUE NOT NULL,
     password_hash TEXT NOT NULL, -- "" for OAuth-only users
     name TEXT NOT NULL,
     avatar_url TEXT NOT NULL DEFAULT '',
     provider TEXT NOT NULL DEFAULT 'local', -- 'local', 'google'
     provider_id TEXT NOT NULL DEFAULT '', -- Google user ID
     email_verified BOOLEAN NOT NULL DEFAULT FALSE,
     disabled BOOLEAN NOT NULL DEFAULT FALSE,
     created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
     updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
 );

-- +goose Down
DROP TABLE IF EXISTS users;