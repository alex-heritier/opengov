package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/alex/opengov-go/internal/db"
	"github.com/alex/opengov-go/internal/models"
)

type RawPolicyDocumentRepository struct {
	db *db.DB
}

func NewRawPolicyDocumentRepository(db *db.DB) *RawPolicyDocumentRepository {
	return &RawPolicyDocumentRepository{db: db}
}

func (r *RawPolicyDocumentRepository) Create(ctx context.Context, tx *sql.Tx, sourceKey, externalID string, rawPayload []byte, fetchedAt time.Time, policyDocID int) error {
	query := `
		INSERT INTO raw_policy_documents (source_key, external_id, raw_data, fetched_at, policy_document_id, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
	`

	_, err := tx.ExecContext(ctx, query, sourceKey, externalID, rawPayload, fetchedAt, policyDocID)
	if err != nil {
		return fmt.Errorf("failed to insert raw entry: %w", err)
	}
	return nil
}

func (r *RawPolicyDocumentRepository) GetByID(ctx context.Context, id int) (*models.RawPolicyDocument, error) {
	query := `
		SELECT id, source_key, external_id, raw_data, fetched_at, policy_document_id, created_at
		FROM raw_policy_documents WHERE id = $1
	`
	var entry models.RawPolicyDocument
	var rawData []byte
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&entry.ID,
		&entry.SourceKey,
		&entry.ExternalID,
		&rawData,
		&entry.FetchedAt,
		&entry.PolicyDocumentID,
		&entry.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(rawData, &entry.RawData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal raw_data: %w", err)
	}
	return &entry, nil
}

func (r *RawPolicyDocumentRepository) GetByDocumentID(ctx context.Context, policyDocID int) ([]*models.RawPolicyDocument, error) {
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

	var entries []*models.RawPolicyDocument
	for rows.Next() {
		var entry models.RawPolicyDocument
		var rawData []byte
		err := rows.Scan(
			&entry.ID,
			&entry.SourceKey,
			&entry.ExternalID,
			&rawData,
			&entry.FetchedAt,
			&entry.PolicyDocumentID,
			&entry.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
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
