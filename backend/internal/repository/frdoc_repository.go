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

type FederalRegisterDocumentRepository struct {
	db *db.DB
}

func NewFederalRegisterDocumentRepository(db *db.DB) *FederalRegisterDocumentRepository {
	return &FederalRegisterDocumentRepository{db: db}
}

func (r *FederalRegisterDocumentRepository) GetByID(ctx context.Context, id int) (*models.FederalRegisterDocument, error) {
	query := `
		SELECT id, source, source_id, unique_key, document_number, raw_data, fetched_at, title, agency, summary, keypoints, impact_score, political_score, source_url, published_at, document_type, pdf_url, feed_entry_id, created_at, updated_at
		FROM federal_register_documents WHERE id = $1
	`
	var a models.FederalRegisterDocument
	var rawData []byte
	var agency, impactScore, documentType, pdfURL *string
	var keypointsRaw []byte
	var politicalScore *int
	var feedEntryID sql.NullInt64
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&a.ID, &a.Source, &a.SourceID, &a.UniqueKey, &a.DocumentNumber, &rawData, &a.FetchedAt,
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
	json.Unmarshal(rawData, &a.RawData)
	return &a, nil
}

func (r *FederalRegisterDocumentRepository) GetByDocumentNumber(ctx context.Context, docNumber string) (*models.FederalRegisterDocument, error) {
	query := `
		SELECT id, source, source_id, unique_key, document_number, raw_data, fetched_at, title, agency, summary, keypoints, impact_score, political_score, source_url, published_at, document_type, pdf_url, feed_entry_id, created_at, updated_at
		FROM federal_register_documents WHERE document_number = $1
	`
	var a models.FederalRegisterDocument
	var rawData []byte
	var agency, impactScore, documentType, pdfURL *string
	var keypointsRaw []byte
	var politicalScore *int
	var feedEntryID sql.NullInt64
	err := r.db.QueryRowContext(ctx, query, docNumber).Scan(
		&a.ID, &a.Source, &a.SourceID, &a.UniqueKey, &a.DocumentNumber, &rawData, &a.FetchedAt,
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
	json.Unmarshal(rawData, &a.RawData)
	return &a, nil
}

func (r *FederalRegisterDocumentRepository) ExistsByUniqueKey(ctx context.Context, uniqueKey string) (bool, error) {
	query := "SELECT COUNT(*) FROM federal_register_documents WHERE unique_key = $1"
	var count int
	err := r.db.QueryRowContext(ctx, query, uniqueKey).Scan(&count)
	return count > 0, err
}

func (r *FederalRegisterDocumentRepository) GetByUniqueKey(ctx context.Context, uniqueKey string) (*models.FederalRegisterDocument, error) {
	query := `
		SELECT id, source, source_id, unique_key, document_number, raw_data, fetched_at, title, agency, summary, keypoints, impact_score, political_score, source_url, published_at, document_type, pdf_url, feed_entry_id, created_at, updated_at
		FROM federal_register_documents WHERE unique_key = $1
	`
	var a models.FederalRegisterDocument
	var rawData []byte
	var agency, impactScore, documentType, pdfURL *string
	var keypointsRaw []byte
	var politicalScore *int
	var feedEntryID sql.NullInt64
	err := r.db.QueryRowContext(ctx, query, uniqueKey).Scan(
		&a.ID, &a.Source, &a.SourceID, &a.UniqueKey, &a.DocumentNumber, &rawData, &a.FetchedAt,
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
	json.Unmarshal(rawData, &a.RawData)
	return &a, nil
}

func (r *FederalRegisterDocumentRepository) Create(ctx context.Context, tx *sql.Tx, doc *models.FederalRegisterDocument, feedEntryID int) error {
	rawData, err := json.Marshal(doc.RawData)
	if err != nil {
		return fmt.Errorf("failed to marshal raw_data: %w", err)
	}

	now := time.Now().UTC()
	doc.CreatedAt = now
	doc.UpdatedAt = now
	doc.FetchedAt = now

	var keypointsJSON []byte
	if len(doc.Keypoints) > 0 {
		keypointsJSON, err = json.Marshal(doc.Keypoints)
		if err != nil {
			return fmt.Errorf("failed to marshal keypoints: %w", err)
		}
	}

	query := `
		INSERT INTO federal_register_documents (source, source_id, unique_key, document_number, raw_data, fetched_at, title, agency, summary, keypoints, impact_score, political_score, source_url, published_at, document_type, pdf_url, feed_entry_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
		RETURNING id
	`
	err = tx.QueryRowContext(ctx, query,
		doc.Source, doc.SourceID, doc.UniqueKey, doc.DocumentNumber, rawData, doc.FetchedAt,
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

func (r *FederalRegisterDocumentRepository) Update(ctx context.Context, tx *sql.Tx, doc *models.FederalRegisterDocument) error {
	rawData, err := json.Marshal(doc.RawData)
	if err != nil {
		return fmt.Errorf("failed to marshal raw_data: %w", err)
	}

	doc.UpdatedAt = time.Now().UTC()

	var keypointsJSON []byte
	if len(doc.Keypoints) > 0 {
		keypointsJSON, err = json.Marshal(doc.Keypoints)
		if err != nil {
			return fmt.Errorf("failed to marshal keypoints: %w", err)
		}
	}

	query := `
		UPDATE federal_register_documents
		SET source = $1, source_id = $2, unique_key = $3, document_number = $4, raw_data = $5, fetched_at = $6,
			title = $7, agency = $8, summary = $9, keypoints = $10, impact_score = $11, political_score = $12,
			source_url = $13, published_at = $14, document_type = $15, pdf_url = $16, updated_at = $17
		WHERE id = $18
	`
	_, err = tx.ExecContext(ctx, query,
		doc.Source, doc.SourceID, doc.UniqueKey, doc.DocumentNumber, rawData, doc.FetchedAt,
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

func (r *FederalRegisterDocumentRepository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM federal_register_documents").Scan(&count)
	return count, err
}

func (r *FederalRegisterDocumentRepository) GetLatest(ctx context.Context) (*models.FederalRegisterDocument, error) {
	query := `
		SELECT id, source, source_id, unique_key, document_number, raw_data, fetched_at, title, agency, summary, keypoints, impact_score, political_score, source_url, published_at, document_type, pdf_url, feed_entry_id, created_at, updated_at
		FROM federal_register_documents
		ORDER BY fetched_at DESC
		LIMIT 1
	`
	var a models.FederalRegisterDocument
	var rawData []byte
	var agency, impactScore, documentType, pdfURL *string
	var keypointsRaw []byte
	var politicalScore *int
	var feedEntryID sql.NullInt64
	err := r.db.QueryRowContext(ctx, query).Scan(
		&a.ID, &a.Source, &a.SourceID, &a.UniqueKey, &a.DocumentNumber, &rawData, &a.FetchedAt,
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
	json.Unmarshal(rawData, &a.RawData)
	return &a, nil
}
