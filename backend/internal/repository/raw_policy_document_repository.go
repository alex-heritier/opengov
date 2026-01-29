package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/alex/opengov-go/internal/db"
	"github.com/alex/opengov-go/internal/domain"
)

type RawPolicyDocumentRepository struct {
	db *db.DB
}

func NewRawPolicyDocumentRepository(db *db.DB) *RawPolicyDocumentRepository {
	return &RawPolicyDocumentRepository{db: db}
}

// Create inserts a raw_policy_documents row.
// If a row already exists for (source_key, external_id), it is treated as already ingested.
func (r *RawPolicyDocumentRepository) Create(ctx context.Context, tx *sql.Tx, sourceKey, externalID string, rawPayload []byte, fetchedAt time.Time, policyDocID *int64) (inserted bool, err error) {
	query := `
		INSERT INTO raw_policy_documents (source_key, external_id, raw_data, fetched_at, policy_document_id)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (source_key, external_id) DO NOTHING
	`

	res, err := tx.ExecContext(ctx, query, sourceKey, externalID, rawPayload, fetchedAt, policyDocID)
	if err != nil {
		return false, fmt.Errorf("failed to insert raw entry: %w", err)
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to read rows affected: %w", err)
	}
	return ra > 0, nil
}

func (r *RawPolicyDocumentRepository) GetByID(ctx context.Context, id int64) (*domain.RawPolicyDocument, error) {
	query := `
		SELECT id, source_key, external_id, raw_data, fetched_at, policy_document_id, created_at
		FROM raw_policy_documents WHERE id = $1
	`
	var entry domain.RawPolicyDocument
	var rawData []byte
	var policyDocID sql.NullInt64
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&entry.ID,
		&entry.SourceKey,
		&entry.ExternalID,
		&rawData,
		&entry.FetchedAt,
		&policyDocID,
		&entry.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	if policyDocID.Valid {
		v := policyDocID.Int64
		entry.PolicyDocumentID = &v
	}
	if err := json.Unmarshal(rawData, &entry.RawData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal raw_data: %w", err)
	}
	return &entry, nil
}

func (r *RawPolicyDocumentRepository) GetByDocumentID(ctx context.Context, policyDocID int64) ([]*domain.RawPolicyDocument, error) {
	query := `
		SELECT id, source_key, external_id, raw_data, fetched_at, policy_document_id, created_at
		FROM raw_policy_documents WHERE policy_document_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query, policyDocID)
	if err != nil {
		return nil, fmt.Errorf("failed to query raw entries: %w", err)
	}
	defer rows.Close()

	var entries []*domain.RawPolicyDocument
	for rows.Next() {
		var entry domain.RawPolicyDocument
		var rawData []byte
		var pdid sql.NullInt64
		err := rows.Scan(
			&entry.ID,
			&entry.SourceKey,
			&entry.ExternalID,
			&rawData,
			&entry.FetchedAt,
			&pdid,
			&entry.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		if pdid.Valid {
			v := pdid.Int64
			entry.PolicyDocumentID = &v
		}
		if err := json.Unmarshal(rawData, &entry.RawData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal raw_data: %w", err)
		}
		entries = append(entries, &entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}
	return entries, nil
}

type UnlinkedRawPolicyDocumentRow struct {
	ID         int64
	SourceKey  string
	ExternalID string
	RawData    []byte
	FetchedAt  time.Time
	CreatedAt  time.Time
}

func (r *RawPolicyDocumentRepository) ListUnlinked(ctx context.Context, limit int) ([]UnlinkedRawPolicyDocumentRow, error) {
	query := `
		SELECT id, source_key, external_id, raw_data, fetched_at, created_at
		FROM raw_policy_documents
		WHERE policy_document_id IS NULL
		ORDER BY created_at ASC
		LIMIT $1
	`
	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query unlinked raw entries: %w", err)
	}
	defer rows.Close()

	var out []UnlinkedRawPolicyDocumentRow
	for rows.Next() {
		var row UnlinkedRawPolicyDocumentRow
		if err := rows.Scan(&row.ID, &row.SourceKey, &row.ExternalID, &row.RawData, &row.FetchedAt, &row.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan unlinked raw entry: %w", err)
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating unlinked raw entries: %w", err)
	}
	return out, nil
}

func (r *RawPolicyDocumentRepository) LinkToPolicyDocument(ctx context.Context, tx *sql.Tx, rawID, policyDocID int64) error {
	query := `
		UPDATE raw_policy_documents
		SET policy_document_id = $1
		WHERE id = $2 AND policy_document_id IS NULL
	`
	res, err := tx.ExecContext(ctx, query, policyDocID, rawID)
	if err != nil {
		return fmt.Errorf("failed to link raw_policy_document %d to policy_document %d: %w", rawID, policyDocID, err)
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to read rows affected when linking raw_policy_document %d: %w", rawID, err)
	}
	if ra == 1 {
		return nil
	}

	// No-op: either the row doesn't exist, or it's already linked (race / retry).
	var existing sql.NullInt64
	check := `SELECT policy_document_id FROM raw_policy_documents WHERE id = $1`
	if err := tx.QueryRowContext(ctx, check, rawID).Scan(&existing); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("raw_policy_document %d not found (cannot link)", rawID)
		}
		return fmt.Errorf("failed to verify link for raw_policy_document %d: %w", rawID, err)
	}
	if existing.Valid {
		// Intentionally treat already-linked as success (idempotent).
		return nil
	}
	return fmt.Errorf("raw_policy_document %d still unlinked after link attempt", rawID)
}
