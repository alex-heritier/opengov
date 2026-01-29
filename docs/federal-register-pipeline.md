# Federal Register Pipeline (Jobs Refactor Plan)

## Context

The legacy pipeline was monolithic (`backend/cmd/scraper/`).

It has been refactored into a single jobs executable that can run one stage at a time:

- Entrypoint: `backend/cmd/jobs/main.go`
- One run (`--job pipeline`) will do: agency sync + scrape + canonicalize + enrich + materialize (**enrichment not implemented yet**)
- Writes: `policy_documents`, `feed_entries`, `raw_policy_documents`

## Goal

- Preserve the pipeline stages/logic (scrape -> canonicalize -> enrich -> materialize).
- Replace `backend/cmd/scraper/` with `backend/cmd/jobs/`.
- One executable, one selector flag: `--job <name>`.
- Keep the API server as-is: `backend/cmd/api/main.go`.

## Data Flow

| Stage | Tables | Purpose |
| --- | --- | --- |
| Raw ingestion | `raw_policy_documents` | Persist upstream JSON payload per upstream document |
| Canonicalization | `policy_documents` | Canonical doc identity + stable source fields |
| Enrichment | `policy_documents` | AI fields (summary/keypoints/scores) |
| Materialization | `feed_entries` | Feed-ready row derived from policy document fields |

Current idempotency invariants (must remain true):

- `policy_documents`: UNIQUE (`source_key`, `external_id`)
- `raw_policy_documents`: UNIQUE (`source_key`, `external_id`)
- `feed_entries`: UPSERT on `policy_document_id`

## CLI Design

New jobs entrypoint:

- `backend/cmd/jobs/main.go`

Jobs are selected via a single flag:

- `./jobs --job migrate`
- `./jobs --job sync-agencies`
- `./jobs --job scrape`
- `./jobs --job canonicalize`
- `./jobs --job enrich`
- `./jobs --job materialize`
- `./jobs --job pipeline` (runs stages in order)

Rule: exactly one job runs per invocation (except `pipeline`, which runs multiple stages sequentially).

## Stage Definitions

### 0) DB migrations (`--job migrate`)

- Runs `internal/db.RunMigrations()` and exits.

### Helper) Agency sync (`--job sync-agencies`)

- Input: Federal Register Agencies API
- Output: `agencies`
- Idempotency: upsert by Federal Register agency ID

### 1) Raw ingestion (`--job scrape`)

- Input: Federal Register Documents API
- Output: `raw_policy_documents`
- Idempotency: UNIQUE (`source_key`, `external_id`); on conflict, treat as already ingested

Design note: raw ingestion must not require a `policy_documents` row.

### 2) Canonicalization (`--job canonicalize`)

- Input: `raw_policy_documents` where `policy_document_id IS NULL`
- Output:
  - `policy_documents` row (create/update by `source_key` + `external_id`)
  - set `raw_policy_documents.policy_document_id` to the created/found doc id

Schema constraint note: `policy_documents.summary` is currently NOT NULL, so canonicalization must write a non-empty placeholder summary derived from raw (e.g. abstract/excerpts truncated) until enrichment runs.

### 3) Enrichment (`--job enrich`) (planned; not implemented yet)

- Input: `policy_documents`
- Output: AI fields on `policy_documents` (summary, keypoints, impact_score, political_score)
- Selection: documents where AI fields are missing (e.g. `impact_score IS NULL` OR `political_score IS NULL` OR keypoints empty)

### 4) Materialization (`--job materialize`)

- Input: `policy_documents`
- Output: `feed_entries` via upsert keyed by `policy_document_id`
- Idempotency: UPSERT on `policy_document_id`

### Pipeline (`--job pipeline`)

Runs, in order:

1) `sync-agencies`
2) `scrape`
3) `canonicalize`
4) `enrich`
5) `materialize`

## Required Schema / Repo Changes

To support raw ingestion before canonicalization:

- Make `raw_policy_documents.policy_document_id` NULL-able (today it is NOT NULL).

Code changes required:

- Domain: `domain.RawPolicyDocument.PolicyDocumentID` becomes nullable (`*int64`).
- Repository scan/insert/update: handle NULL `policy_document_id` (use `sql.NullInt64`).

## Implementation Steps

1) Introduce `backend/cmd/jobs/main.go`
- Move wiring/config/db init from the legacy scraper entrypoint.
- Parse `--job` and dispatch to a single stage.

2) Split the monolith into stage-callable methods
- Keep existing logic, but separate it so each stage can run independently.
- Likely touch: `backend/internal/services/jobs_service.go`.

3) Update raw ingestion to write unlinked raw rows
- Insert into `raw_policy_documents` without `policy_document_id`.

4) Implement canonicalization linking
- Read unlinked raw rows, create/update `policy_documents`, then link raw rows.

5) Implement enrichment as a standalone pass
- Iterate “needs enrichment” docs, call summarizer, update AI fields.

6) Implement materialization as a standalone pass
- Iterate docs (or docs missing feed entries) and upsert `feed_entries`.

7) Optional wrapper
- `backend/scripts/run-pipeline.sh` calls the jobs binary with stage flags.

## Verification

Run:

1) `./jobs --job migrate`
2) `./jobs --job sync-agencies`
3) `./jobs --job scrape`
4) `./jobs --job canonicalize`
5) `./jobs --job enrich`
6) `./jobs --job materialize`

Validate:

- End-to-end: FR API -> raw -> canonical -> enrich -> feed_entries
- Re-run each job: no duplicates; updates are safe and idempotent
