package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/alex/opengov-go/internal/db"
	"github.com/alex/opengov-go/internal/models"
)

type ArticleRepository struct {
	db *db.DB
}

func NewArticleRepository(db *db.DB) *ArticleRepository {
	return &ArticleRepository{db: db}
}

func (r *ArticleRepository) GetFeed(ctx context.Context, page, limit int, sort string) ([]models.FRArticle, int, error) {
	offset := (page - 1) * limit
	var orderDir string
	if sort == "newest" {
		orderDir = "DESC"
	} else {
		orderDir = "ASC"
	}

	query := fmt.Sprintf(`
		SELECT id, source, source_id, unique_key, document_number, raw_data, fetched_at, title, summary, source_url, published_at, document_type, pdf_url, created_at, updated_at
		FROM frarticles
		ORDER BY published_at %s
		LIMIT $1 OFFSET $2
	`, orderDir)

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query articles: %w", err)
	}
	defer rows.Close()

	var articles []models.FRArticle
	for rows.Next() {
		var a models.FRArticle
		var rawData []byte
		var documentType, pdfURL *string
		err := rows.Scan(
			&a.ID, &a.Source, &a.SourceID, &a.UniqueKey, &a.DocumentNumber, &rawData, &a.FetchedAt,
			&a.Title, &a.Summary, &a.SourceURL, &a.PublishedAt,
			&documentType, &pdfURL, &a.CreatedAt, &a.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan article: %w", err)
		}
		if documentType != nil {
			a.DocumentType = documentType
		}
		if pdfURL != nil {
			a.PDFURL = pdfURL
		}
		json.Unmarshal(rawData, &a.RawData)
		articles = append(articles, a)
	}

	var total int
	countQuery := "SELECT COUNT(*) FROM frarticles"
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count articles: %w", err)
	}

	return articles, total, nil
}

func (r *ArticleRepository) GetByID(ctx context.Context, id int) (*models.FRArticle, error) {
	query := `
		SELECT id, source, source_id, unique_key, document_number, raw_data, fetched_at, title, summary, source_url, published_at, document_type, pdf_url, created_at, updated_at
		FROM frarticles WHERE id = $1
	`
	var a models.FRArticle
	var rawData []byte
	var documentType, pdfURL *string
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&a.ID, &a.Source, &a.SourceID, &a.UniqueKey, &a.DocumentNumber, &rawData, &a.FetchedAt,
		&a.Title, &a.Summary, &a.SourceURL, &a.PublishedAt,
		&documentType, &pdfURL, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if documentType != nil {
		a.DocumentType = documentType
	}
	if pdfURL != nil {
		a.PDFURL = pdfURL
	}
	json.Unmarshal(rawData, &a.RawData)
	return &a, nil
}

func (r *ArticleRepository) GetByDocumentNumber(ctx context.Context, docNumber string) (*models.FRArticle, error) {
	query := `
		SELECT id, source, source_id, unique_key, document_number, raw_data, fetched_at, title, summary, source_url, published_at, document_type, pdf_url, created_at, updated_at
		FROM frarticles WHERE document_number = $1
	`
	var a models.FRArticle
	var rawData []byte
	var documentType, pdfURL *string
	err := r.db.QueryRowContext(ctx, query, docNumber).Scan(
		&a.ID, &a.Source, &a.SourceID, &a.UniqueKey, &a.DocumentNumber, &rawData, &a.FetchedAt,
		&a.Title, &a.Summary, &a.SourceURL, &a.PublishedAt,
		&documentType, &pdfURL, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if documentType != nil {
		a.DocumentType = documentType
	}
	if pdfURL != nil {
		a.PDFURL = pdfURL
	}
	json.Unmarshal(rawData, &a.RawData)
	return &a, nil
}

func (r *ArticleRepository) ExistsByUniqueKey(ctx context.Context, uniqueKey string) (bool, error) {
	query := "SELECT COUNT(*) FROM frarticles WHERE unique_key = $1"
	var count int
	err := r.db.QueryRowContext(ctx, query, uniqueKey).Scan(&count)
	return count > 0, err
}

func (r *ArticleRepository) GetByUniqueKey(ctx context.Context, uniqueKey string) (*models.FRArticle, error) {
	query := `
		SELECT id, source, source_id, unique_key, document_number, raw_data, fetched_at, title, summary, source_url, published_at, document_type, pdf_url, created_at, updated_at
		FROM frarticles WHERE unique_key = $1
	`
	var a models.FRArticle
	var rawData []byte
	var documentType, pdfURL *string
	err := r.db.QueryRowContext(ctx, query, uniqueKey).Scan(
		&a.ID, &a.Source, &a.SourceID, &a.UniqueKey, &a.DocumentNumber, &rawData, &a.FetchedAt,
		&a.Title, &a.Summary, &a.SourceURL, &a.PublishedAt,
		&documentType, &pdfURL, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if documentType != nil {
		a.DocumentType = documentType
	}
	if pdfURL != nil {
		a.PDFURL = pdfURL
	}
	json.Unmarshal(rawData, &a.RawData)
	return &a, nil
}

func (r *ArticleRepository) Create(ctx context.Context, article *models.FRArticle) error {
	rawData, err := json.Marshal(article.RawData)
	if err != nil {
		return fmt.Errorf("failed to marshal raw_data: %w", err)
	}

	now := time.Now().UTC()
	article.CreatedAt = now
	article.UpdatedAt = now
	article.FetchedAt = now

	query := `
		INSERT INTO frarticles (source, source_id, unique_key, document_number, raw_data, fetched_at, title, summary, source_url, published_at, document_type, pdf_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id
	`
	err = r.db.QueryRowContext(ctx, query,
		article.Source, article.SourceID, article.UniqueKey, article.DocumentNumber, rawData, article.FetchedAt,
		article.Title, article.Summary, article.SourceURL, article.PublishedAt,
		article.DocumentType, article.PDFURL,
		article.CreatedAt, article.UpdatedAt,
	).Scan(&article.ID)
	if err != nil {
		return fmt.Errorf("failed to insert article: %w", err)
	}

	return nil
}

func (r *ArticleRepository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM frarticles").Scan(&count)
	return count, err
}

func (r *ArticleRepository) GetLatest(ctx context.Context) (*models.FRArticle, error) {
	query := `
		SELECT id, source, source_id, unique_key, document_number, raw_data, fetched_at, title, summary, source_url, published_at, document_type, pdf_url, created_at, updated_at
		FROM frarticles
		ORDER BY fetched_at DESC
		LIMIT 1
	`
	var a models.FRArticle
	var rawData []byte
	var documentType, pdfURL *string
	err := r.db.QueryRowContext(ctx, query).Scan(
		&a.ID, &a.Source, &a.SourceID, &a.UniqueKey, &a.DocumentNumber, &rawData, &a.FetchedAt,
		&a.Title, &a.Summary, &a.SourceURL, &a.PublishedAt,
		&documentType, &pdfURL, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if documentType != nil {
		a.DocumentType = documentType
	}
	if pdfURL != nil {
		a.PDFURL = pdfURL
	}
	json.Unmarshal(rawData, &a.RawData)
	return &a, nil
}
