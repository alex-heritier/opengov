# Federal Register Data Processing Pipeline

## Overview

| Stage            | Table                  | Purpose                                     |
| ---------------- | ---------------------- | ------------------------------------------- |
| Raw ingestion    | `raw_policy_documents` | Staging area for scraped data               |
| Canonicalization | `policy_documents`     | Authoritative source with document identity |
| Enrichment       | `policy_documents`     | AI processing (summary/keypoints/scores)    |
| Materialization  | `feed_entries`         | Feed-ready record derived from policy docs  |

## Stages

### 1) Raw ingestion (scraper)

- **Input**: Federal Register API
- **Output**: `raw_policy_documents`
- **Idempotency**: UPSERT on (`source_key`, `external_id`) so re-scrapes update the same raw record.

### 2) Canonicalization (service)

- **Output**: `policy_documents`
- **Idempotency**: UPSERT on `unique_key` so the same document identity is updated, not duplicated.

### 3) Enrichment (AI service)

- **Output**: updates AI fields on `policy_documents`
- **Idempotency**: for each enrichment field, only write when the target field is NULL/empty; if already set, skip that enrichment step.

### 4) Materialization (service)

- **Output**: creates `feed_entries` record
- **Idempotency**: UPSERT on `policy_document_id` so each policy document has at most one feed entry.

## Architecture

Each stage will have its own `backend/cmd/.../` directory:

- Scraper: `backend/cmd/scraper/` (exists)
- Canonicalization: `backend/cmd/canonicalize/` (new)
- Enrichment: `backend/cmd/enrichment/` (new)
- Materialization: `backend/cmd/materialize/` (new)

## Usage

The pipeline will be orchestrated by `backend/scripts/run-pipeline.sh`, which runs each executable in sequence.
