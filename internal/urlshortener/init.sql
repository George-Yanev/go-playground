-- Enable Write-Ahead Logging for better concurrency
PRAGMA journal_mode = WAL;

-- Seeds Table
CREATE TABLE IF NOT EXISTS seeds (
    seed TEXT PRIMARY KEY, -- 24-bit value
    counter_size INTEGER NOT NULL DEFAULT 4096,
    counter_used INTEGER NOT NULL DEFAULT 0,
    lease_holder TEXT DEFAULT '',
    lease_taken DATETIME NULL,
    status INTEGER NOT NULL DEFAULT 0 CHECK (status IN (0, 1, 2)) -- available|used|exhausted
);

-- URL Mappings
CREATE TABLE IF NOT EXISTS url_mapping (
    original_url TEXT NOT NULL COLLATE NOCASE,
    short_url TEXT PRIMARY KEY, -- base62 encoded
    seed INTEGER NOT NULL,
    counter INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for fast lookups
CREATE UNIQUE INDEX IF NOT EXISTS idx_seed_counter ON url_mapping (seed, counter);

CREATE INDEX IF NOT EXISTS idx_original_url ON url_mapping (original_url);
