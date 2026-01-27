| Stage | Table | Purpose |
|-------|-------|---------|
| Raw ingestion | `raw_policy_documents` | Staging area for scraped data |
| Canonicalization | `policy_documents` | Authoritative source with document identity |
| Enrichment | `feed_entries` + `enrichments` | AI-processed results |

---

# Federal Register Data Processing Pipeline  


Federal Register API

        ↓

Raw ingestion (scraper)
→ `raw_policy_documents` table

        ↓

canonicalization service
→ `policy_documents` table

        ↓

enrichment service (AI)
→ `feed_entries` + `enrichments` tables

