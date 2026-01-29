-- 008_raw_policy_documents_nullable_policy_document_id.sql
-- Allow raw ingestion before canonicalization by making policy_document_id nullable.

ALTER TABLE raw_policy_documents
    ALTER COLUMN policy_document_id DROP NOT NULL;

