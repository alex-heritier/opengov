package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/alex/opengov-go/internal/db"
	"github.com/alex/opengov-go/internal/domain"
	"github.com/lib/pq"
)

var ErrDuplicateDocument = errors.New("document already exists")

type PolicyDocumentRepository struct {
	db *db.DB
}

func NewPolicyDocumentRepository(db *db.DB) *PolicyDocumentRepository {
	return &PolicyDocumentRepository{db: db}
}

func (r *PolicyDocumentRepository) GetByID(ctx context.Context, id int) (*domain.PolicyDocument, error) {
	query := `
		SELECT id, source_key, external_id, fetched_at, title, agency, summary, keypoints, impact_score, political_score, source_url, published_at, document_type, pdf_url, created_at, updated_at
		FROM policy_documents WHERE id = $1
	`
	var a domain.PolicyDocument
	var agency, impactScore, documentType, pdfURL *string
	var keypointsRaw []byte
	var politicalScore *int
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&a.ID, &a.SourceKey, &a.ExternalID, &a.FetchedAt,
		&a.Title, &agency, &a.Summary, &keypointsRaw, &impactScore, &politicalScore, &a.SourceURL, &a.PublishedAt,
		&documentType, &pdfURL, &a.CreatedAt, &a.UpdatedAt,
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
	return &a, nil
}

func (r *PolicyDocumentRepository) ExistsBySourceKeyExternalID(ctx context.Context, sourceKey, externalID string) (bool, error) {
	query := "SELECT COUNT(*) FROM policy_documents WHERE source_key = $1 AND external_id = $2"
	var count int
	err := r.db.QueryRowContext(ctx, query, sourceKey, externalID).Scan(&count)
	return count > 0, err
}

func (r *PolicyDocumentRepository) GetBySourceKeyExternalID(ctx context.Context, sourceKey, externalID string) (*domain.PolicyDocument, error) {
	query := `
		SELECT id, source_key, external_id, fetched_at, title, agency, summary, keypoints, impact_score, political_score, source_url, published_at, document_type, pdf_url, created_at, updated_at
		FROM policy_documents WHERE source_key = $1 AND external_id = $2
	`
	var a domain.PolicyDocument
	var agency, impactScore, documentType, pdfURL *string
	var keypointsRaw []byte
	var politicalScore *int
	err := r.db.QueryRowContext(ctx, query, sourceKey, externalID).Scan(
		&a.ID, &a.SourceKey, &a.ExternalID, &a.FetchedAt,
		&a.Title, &agency, &a.Summary, &keypointsRaw, &impactScore, &politicalScore, &a.SourceURL, &a.PublishedAt,
		&documentType, &pdfURL, &a.CreatedAt, &a.UpdatedAt,
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
	return &a, nil
}

func (r *PolicyDocumentRepository) Create(ctx context.Context, tx *sql.Tx, doc *domain.PolicyDocument) error {
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
		INSERT INTO policy_documents (source_key, external_id, fetched_at, title, agency, summary, keypoints, impact_score, political_score, source_url, published_at, document_type, pdf_url)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id
	`
	err = tx.QueryRowContext(ctx, query,
		doc.SourceKey, doc.ExternalID, doc.FetchedAt,
		doc.Title, doc.Agency, doc.Summary, keypointsJSON, doc.ImpactScore, doc.PoliticalScore,
		doc.SourceURL, doc.PublishedAt,
		doc.DocumentType, doc.PDFURL,
	).Scan(&doc.ID)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return ErrDuplicateDocument
		}
		return fmt.Errorf("failed to insert document: %w", err)
	}

	return nil
}

func (r *PolicyDocumentRepository) Update(ctx context.Context, tx *sql.Tx, doc *domain.PolicyDocument) error {
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
		SET source_key = $1, external_id = $2, fetched_at = $3,
			title = $4, agency = $5, summary = $6, keypoints = $7, impact_score = $8, political_score = $9,
			source_url = $10, published_at = $11, document_type = $12, pdf_url = $13,
			updated_at = NOW()
		WHERE id = $14
	`
	_, err = tx.ExecContext(ctx, query,
		doc.SourceKey, doc.ExternalID, doc.FetchedAt,
		doc.Title, doc.Agency, doc.Summary, keypointsJSON, doc.ImpactScore, doc.PoliticalScore,
		doc.SourceURL, doc.PublishedAt,
		doc.DocumentType, doc.PDFURL,
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

func (r *PolicyDocumentRepository) GetLatest(ctx context.Context) (*domain.PolicyDocument, error) {
	query := `
		SELECT id, source_key, external_id, fetched_at, title, agency, summary, keypoints, impact_score, political_score, source_url, published_at, document_type, pdf_url, created_at, updated_at
		FROM policy_documents
		ORDER BY fetched_at DESC
		LIMIT 1
	`
	var a domain.PolicyDocument
	var agency, impactScore, documentType, pdfURL *string
	var keypointsRaw []byte
	var politicalScore *int
	err := r.db.QueryRowContext(ctx, query).Scan(
		&a.ID, &a.SourceKey, &a.ExternalID, &a.FetchedAt,
		&a.Title, &agency, &a.Summary, &keypointsRaw, &impactScore, &politicalScore, &a.SourceURL, &a.PublishedAt,
		&documentType, &pdfURL, &a.CreatedAt, &a.UpdatedAt,
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
	return &a, nil
}
