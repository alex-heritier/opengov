
| Stage            | Table                  | Purpose                                     |
| ---------------- | ---------------------- | ------------------------------------------- |
| Raw ingestion    | `raw_policy_documents` | Staging area for scraped data               |
| Canonicalization | `policy_documents`     | Authoritative source with document identity |
| Enrichment       | `policy_documents`     | AI processing (summary/keypoints/scores)    |
| Materialization  | `feed_entries`         | Feed-ready record derived from policy docs  |


---

# Federal Register Data Processing Pipeline

## Stages

Federal Register API

```
    ↓
```

Raw ingestion (scraper)
→ `raw_policy_documents` table

- Idempotency check: UPSERT on (`source_key`, `external_id`) so re-scrapes update the same raw record.

```
    ↓
```

canonicalization service
→ `policy_documents` table

- Idempotency check: UPSERT on `unique_key` so the same document identity is updated, not duplicated.

```
    ↓
```

enrichment service (AI)
→ updates AI fields on `policy_documents`

- Idempotency check: for each enrichment field, only write when the target field is NULL/empty; if already set, skip that enrichment step.

```
    ↓
```

materialization service
→ creates `feed_entries` record

- Idempotency check: UPSERT on `policy_document_id` so each policy document has at most one feed entry.

## Architecture

Each stage will have its own backend/cmd/.../ directory

- Scraper: backend/cmd/scraper/ (exists)
- Canonicalization: backend/cmd/canonicalize/ (new)
- Enrichment: backend/cmd/enrichment/ (new)
- Materialization: backend/cmd/materialize/ (new)

## Usage

The pipeline will be orchestrated by a backend/scripts/run-pipeline.sh script that handles running each executable in sequence.

