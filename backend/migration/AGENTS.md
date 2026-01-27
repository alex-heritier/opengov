# Migration Guidelines

## Auto-updating timestamps

PostgreSQL does **not** support a MySQL-style `ON UPDATE NOW()` column attribute for `updated_at`.

If you need `updated_at` to change on row updates, you must do it in application SQL (preferred in this codebase) or with a trigger (not used here).

```sql
updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
```

**Do not** use triggers for this purpose.

## Timestamp defaults

Always set `DEFAULT NOW()` for both `created_at` and `updated_at`:

```sql
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
```

## Repository code

### Updates
When updating rows with `updated_at` columns:
- Always set `updated_at = NOW()` in the `UPDATE ... SET` clause (and in `ON CONFLICT ... DO UPDATE SET` upserts)
- Prefer database time (`NOW()`) over application time to avoid clock skew

### Inserts
**Do not** include `created_at` or `updated_at` in INSERT columns - both have defaults:

```sql
INSERT INTO table (col1, col2) VALUES ($1, $2)
```

This avoids clock skew and simplifies application code.
