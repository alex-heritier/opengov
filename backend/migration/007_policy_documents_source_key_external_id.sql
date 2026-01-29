-- 007_policy_documents_source_key_external_id.sql
-- Consolidate policy_documents upstream identifiers to match raw_policy_documents:
--   policy_documents(source_key, external_id)
--
-- This migration:
-- - Adds source_key/external_id to policy_documents
-- - Backfills them from legacy columns (source/source_id)
-- - Enforces NOT NULL + UNIQUE(source_key, external_id)
-- - Drops legacy identifier columns (source, source_id, unique_key, document_number)

-- 1) Add new columns (nullable for backfill)
ALTER TABLE policy_documents
    ADD COLUMN IF NOT EXISTS source_key TEXT;

ALTER TABLE policy_documents
    ADD COLUMN IF NOT EXISTS external_id TEXT;

-- 2) Backfill from legacy columns
UPDATE policy_documents
SET
    source_key = source,
    external_id = source_id
WHERE source_key IS NULL OR external_id IS NULL;

-- 3) Enforce NOT NULL
ALTER TABLE policy_documents
    ALTER COLUMN source_key SET NOT NULL;

ALTER TABLE policy_documents
    ALTER COLUMN external_id SET NOT NULL;

-- 4) Enforce uniqueness for idempotent ingestion
CREATE UNIQUE INDEX IF NOT EXISTS idx_policy_documents_source_key_external_id
    ON policy_documents(source_key, external_id);

-- 5) Drop legacy identifier columns (replaced by source_key/external_id)
ALTER TABLE policy_documents
    DROP COLUMN IF EXISTS unique_key;

ALTER TABLE policy_documents
    DROP COLUMN IF EXISTS document_number;

ALTER TABLE policy_documents
    DROP COLUMN IF EXISTS source_id;

ALTER TABLE policy_documents
    DROP COLUMN IF EXISTS source;

-- 6) Drop now-redundant legacy index (if it still exists)
DROP INDEX IF EXISTS idx_policy_documents_unique_key;

