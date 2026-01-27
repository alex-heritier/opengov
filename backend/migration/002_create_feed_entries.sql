-- 002_create_feed_entries.sql
-- feed_entries

CREATE TABLE IF NOT EXISTS feed_entries (
    id BIGSERIAL PRIMARY KEY,
    source_type TEXT NOT NULL,
    title TEXT NOT NULL,
    short_text TEXT NOT NULL,
    key_points JSONB,
    political_score INTEGER,
    impact_score TEXT,
    source_url TEXT NOT NULL,
    published_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_feed_entries_published_at ON feed_entries(published_at DESC);
CREATE INDEX IF NOT EXISTS idx_feed_entries_source_type ON feed_entries(source_type);
