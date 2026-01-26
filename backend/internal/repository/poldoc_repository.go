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

type PolicyDocumentRepository struct {
	db *db.DB
}

func NewPolicyDocumentRepository(db *db.DB) *PolicyDocumentRepository {
	return &PolicyDocumentRepository{db: db}
}

func (r *PolicyDocumentRepository) GetByID(ctx context.Context, id int) (*models.PolicyDocument, error) {
	query := `
		SELECT id, source, source_id, unique_key, document_number, fetched_at, title, agency, summary, keypoints, impact_score, political_score, source_url, published_at, document_type, pdf_url, feed_entry_id, created_at, updated_at
		FROM policy_documents WHERE id = $1
	`
	var a models.PolicyDocument
	var agency, impactScore, documentType, pdfURL *string
	var keypointsRaw []byte
	var politicalScore *int
	var feedEntryID sql.NullInt64
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&a.ID, &a.Source, &a.SourceID, &a.UniqueKey, &a.DocumentNumber, &a.FetchedAt,
		&a.Title, &agency, &a.Summary, &keypointsRaw, &impactScore, &politicalScore, &a.SourceURL, &a.PublishedAt,
		&documentType, &pdfURL, &feedEntryID, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	a.Agency = agency
	if len(keypointsRaw) > 0 {
		json.Unmarshal(keypointsRaw, &a.Keypoints)
	}
	a.ImpactScore = impactScore
	a.PoliticalScore = politicalScore
	a.DocumentType = documentType
	a.PDFURL = pdfURL
	if feedEntryID.Valid {
		a.FeedEntryID = int(feedEntryID.Int64)
	}
	return &a, nil
}

func (r *PolicyDocumentRepository) GetByDocumentNumber(ctx context.Context, docNumber string) (*models.PolicyDocument, error) {
	query := `
		SELECT id, source, source_id, unique_key, document_number, fetched_at, title, agency, summary, keypoints, impact_score, political_score, source_url, published_at, document_type, pdf_url, feed_entry_id, created_at, updated_at
		FROM policy_documents WHERE document_number = $1
	`
	var a models.PolicyDocument
	var agency, impactScore, documentType, pdfURL *string
	var keypointsRaw []byte
	var politicalScore *int
	var feedEntryID sql.NullInt64
	err := r.db.QueryRowContext(ctx, query, docNumber).Scan(
		&a.ID, &a.Source, &a.SourceID, &a.UniqueKey, &a.DocumentNumber, &a.FetchedAt,
		&a.Title, &agency, &a.Summary, &keypointsRaw, &impactScore, &politicalScore, &a.SourceURL, &a.PublishedAt,
		&documentType, &pdfURL, &feedEntryID, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	a.Agency = agency
	if len(keypointsRaw) > 0 {
		json.Unmarshal(keypointsRaw, &a.Keypoints)
	}
	a.ImpactScore = impactScore
	a.PoliticalScore = politicalScore
	a.DocumentType = documentType
	a.PDFURL = pdfURL
	if feedEntryID.Valid {
		a.FeedEntryID = int(feedEntryID.Int64)
	}
	return &a, nil
}

func (r *PolicyDocumentRepository) ExistsByUniqueKey(ctx context.Context, uniqueKey string) (bool, error) {
	query := "SELECT COUNT(*) FROM policy_documents WHERE unique_key = $1"
	var count int
	err := r.db.QueryRowContext(ctx, query, uniqueKey).Scan(&count)
	return count > 0, err
}

func (r *PolicyDocumentRepository) GetByUniqueKey(ctx context.Context, uniqueKey string) (*models.PolicyDocument, error) {
	query := `
		SELECT id, source, source_id, unique_key, document_number, fetched_at, title, agency, summary, keypoints, impact_score, political_score, source_url, published_at, document_type, pdf_url, feed_entry_id, created_at, updated_at
		FROM policy_documents WHERE unique_key = $1
	`
	var a models.PolicyDocument
	var agency, impactScore, documentType, pdfURL *string
	var keypointsRaw []byte
	var politicalScore *int
	var feedEntryID sql.NullInt64
	err := r.db.QueryRowContext(ctx, query, uniqueKey).Scan(
		&a.ID, &a.Source, &a.SourceID, &a.UniqueKey, &a.DocumentNumber, &a.FetchedAt,
		&a.Title, &agency, &a.Summary, &keypointsRaw, &impactScore, &politicalScore, &a.SourceURL, &a.PublishedAt,
		&documentType, &pdfURL, &feedEntryID, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	a.Agency = agency
	if len(keypointsRaw) > 0 {
		json.Unmarshal(keypointsRaw, &a.Keypoints)
	}
	a.ImpactScore = impactScore
	a.PoliticalScore = politicalScore
	a.DocumentType = documentType
	a.PDFURL = pdfURL
	if feedEntryID.Valid {
		a.FeedEntryID = int(feedEntryID.Int64)
	}
	return &a, nil
}

func (r *PolicyDocumentRepository) Create(ctx context.Context, tx *sql.Tx, doc *models.PolicyDocument, feedEntryID int) error {
	now := time.Now().UTC()
	doc.CreatedAt = now
	doc.UpdatedAt = now
	doc.FetchedAt = now

	var err error
	var keypointsJSON []byte
	if len(doc.Keypoints) > 0 {
		keypointsJSON, err = json.Marshal(doc.Keypoints)
		if err != nil {
			return fmt.Errorf("failed to marshal keypoints: %w", err)
		}
	}

	query := `
		INSERT INTO policy_documents (source, source_id, unique_key, document_number, fetched_at, title, agency, summary, keypoints, impact_score, political_score, source_url, published_at, document_type, pdf_url, feed_entry_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		RETURNING id
	`
	err = tx.QueryRowContext(ctx, query,
		doc.Source, doc.SourceID, doc.UniqueKey, doc.DocumentNumber, doc.FetchedAt,
		doc.Title, doc.Agency, doc.Summary, keypointsJSON, doc.ImpactScore, doc.PoliticalScore,
		doc.SourceURL, doc.PublishedAt,
		doc.DocumentType, doc.PDFURL,
		feedEntryID,
		doc.CreatedAt, doc.UpdatedAt,
	).Scan(&doc.ID)
	if err != nil {
		return fmt.Errorf("failed to insert document: %w", err)
	}

	return nil
}

func (r *PolicyDocumentRepository) Update(ctx context.Context, tx *sql.Tx, doc *models.PolicyDocument) error {
	doc.UpdatedAt = time.Now().UTC()

	var err error
	var keypointsJSON []byte
	if len(doc.Keypoints) > 0 {
		keypointsJSON, err = json.Marshal(doc.Keypoints)
		if err != nil {
			return fmt.Errorf("failed to marshal keypoints: %w", err)
		}
	}

	query := `
		UPDATE policy_documents
		SET source = $1, source_id = $2, unique_key = $3, document_number = $4, fetched_at = $5,
			title = $6, agency = $7, summary = $8, keypoints = $9, impact_score = $10, political_score = $11,
			source_url = $12, published_at = $13, document_type = $14, pdf_url = $15, updated_at = $16
		WHERE id = $17
	`
	_, err = tx.ExecContext(ctx, query,
		doc.Source, doc.SourceID, doc.UniqueKey, doc.DocumentNumber, doc.FetchedAt,
		doc.Title, doc.Agency, doc.Summary, keypointsJSON, doc.ImpactScore, doc.PoliticalScore,
		doc.SourceURL, doc.PublishedAt,
		doc.DocumentType, doc.PDFURL,
		doc.UpdatedAt,
		doc.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update document: %w", err)
	}

	return nil
}

func (r *PolicyDocumentRepository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM policy_documents").Scan(&count)
	return count, err
}

func (r *PolicyDocumentRepository) GetLatest(ctx context.Context) (*models.PolicyDocument, error) {
	query := `
		SELECT id, source, source_id, unique_key, document_number, fetched_at, title, agency, summary, keypoints, impact_score, political_score, source_url, published_at, document_type, pdf_url, feed_entry_id, created_at, updated_at
		FROM policy_documents
		ORDER BY fetched_at DESC
		LIMIT 1
	`
	var a models.PolicyDocument
	var agency, impactScore, documentType, pdfURL *string
	var keypointsRaw []byte
	var politicalScore *int
	var feedEntryID sql.NullInt64
	err := r.db.QueryRowContext(ctx, query).Scan(
		&a.ID, &a.Source, &a.SourceID, &a.UniqueKey, &a.DocumentNumber, &a.FetchedAt,
		&a.Title, &agency, &a.Summary, &keypointsRaw, &impactScore, &politicalScore, &a.SourceURL, &a.PublishedAt,
		&documentType, &pdfURL, &feedEntryID, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	a.Agency = agency
	if len(keypointsRaw) > 0 {
		json.Unmarshal(keypointsRaw, &a.Keypoints)
	}
	a.ImpactScore = impactScore
	a.PoliticalScore = politicalScore
	a.DocumentType = documentType
	a.PDFURL = pdfURL
	if feedEntryID.Valid {
		a.FeedEntryID = int(feedEntryID.Int64)
	}
	return &a, nil
}
