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
		LIMIT ? OFFSET ?
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
		var fetchedAt, publishedAt, createdAt, updatedAt string
		var documentType, pdfURL sql.NullString
		err := rows.Scan(
			&a.ID, &a.Source, &a.SourceID, &a.UniqueKey, &a.DocumentNumber, &rawData, &fetchedAt,
			&a.Title, &a.Summary, &a.SourceURL, &publishedAt,
			&documentType, &pdfURL, &createdAt, &updatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan article: %w", err)
		}
		a.FetchedAt, _ = time.Parse("2006-01-02 15:04:05Z07:00", fetchedAt)
		a.PublishedAt, _ = time.Parse("2006-01-02 15:04:05Z07:00", publishedAt)
		a.CreatedAt, _ = time.Parse("2006-01-02 15:04:05Z07:00", createdAt)
		a.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05Z07:00", updatedAt)
		if documentType.Valid {
			a.DocumentType = &documentType.String
		}
		if pdfURL.Valid {
			a.PDFURL = &pdfURL.String
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
		FROM frarticles WHERE id = ?
	`
	var a models.FRArticle
	var rawData []byte
	var fetchedAt, publishedAt, createdAt, updatedAt string
	var documentType, pdfURL sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&a.ID, &a.Source, &a.SourceID, &a.UniqueKey, &a.DocumentNumber, &rawData, &fetchedAt,
		&a.Title, &a.Summary, &a.SourceURL, &publishedAt,
		&documentType, &pdfURL, &createdAt, &updatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get article: %w", err)
	}
	a.FetchedAt, _ = time.Parse("2006-01-02 15:04:05Z07:00", fetchedAt)
	a.PublishedAt, _ = time.Parse("2006-01-02 15:04:05Z07:00", publishedAt)
	a.CreatedAt, _ = time.Parse("2006-01-02 15:04:05Z07:00", createdAt)
	a.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05Z07:00", updatedAt)
	if documentType.Valid {
		a.DocumentType = &documentType.String
	}
	if pdfURL.Valid {
		a.PDFURL = &pdfURL.String
	}
	json.Unmarshal(rawData, &a.RawData)
	return &a, nil
}

func (r *ArticleRepository) GetByDocumentNumber(ctx context.Context, docNumber string) (*models.FRArticle, error) {
	query := `
		SELECT id, source, source_id, unique_key, document_number, raw_data, fetched_at, title, summary, source_url, published_at, document_type, pdf_url, created_at, updated_at
		FROM frarticles WHERE document_number = ?
	`
	var a models.FRArticle
	var rawData []byte
	var fetchedAt, publishedAt, createdAt, updatedAt string
	var documentType, pdfURL sql.NullString
	err := r.db.QueryRowContext(ctx, query, docNumber).Scan(
		&a.ID, &a.Source, &a.SourceID, &a.UniqueKey, &a.DocumentNumber, &rawData, &fetchedAt,
		&a.Title, &a.Summary, &a.SourceURL, &publishedAt,
		&documentType, &pdfURL, &createdAt, &updatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get article: %w", err)
	}
	a.FetchedAt, _ = time.Parse("2006-01-02 15:04:05Z07:00", fetchedAt)
	a.PublishedAt, _ = time.Parse("2006-01-02 15:04:05Z07:00", publishedAt)
	a.CreatedAt, _ = time.Parse("2006-01-02 15:04:05Z07:00", createdAt)
	a.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05Z07:00", updatedAt)
	if documentType.Valid {
		a.DocumentType = &documentType.String
	}
	if pdfURL.Valid {
		a.PDFURL = &pdfURL.String
	}
	json.Unmarshal(rawData, &a.RawData)
	return &a, nil
}

func (r *ArticleRepository) ExistsByUniqueKey(ctx context.Context, uniqueKey string) (bool, error) {
	query := "SELECT COUNT(*) FROM frarticles WHERE unique_key = ?"
	var count int
	err := r.db.QueryRowContext(ctx, query, uniqueKey).Scan(&count)
	return count > 0, err
}

func (r *ArticleRepository) GetByUniqueKey(ctx context.Context, uniqueKey string) (*models.FRArticle, error) {
	query := `
		SELECT id, source, source_id, unique_key, document_number, raw_data, fetched_at, title, summary, source_url, published_at, document_type, pdf_url, created_at, updated_at
		FROM frarticles WHERE unique_key = ?
	`
	var a models.FRArticle
	var rawData []byte
	var fetchedAt, publishedAt, createdAt, updatedAt string
	var documentType, pdfURL sql.NullString
	err := r.db.QueryRowContext(ctx, query, uniqueKey).Scan(
		&a.ID, &a.Source, &a.SourceID, &a.UniqueKey, &a.DocumentNumber, &rawData, &fetchedAt,
		&a.Title, &a.Summary, &a.SourceURL, &publishedAt,
		&documentType, &pdfURL, &createdAt, &updatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get article by unique_key: %w", err)
	}
	a.FetchedAt, _ = time.Parse("2006-01-02 15:04:05Z07:00", fetchedAt)
	a.PublishedAt, _ = time.Parse("2006-01-02 15:04:05Z07:00", publishedAt)
	a.CreatedAt, _ = time.Parse("2006-01-02 15:04:05Z07:00", createdAt)
	a.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05Z07:00", updatedAt)
	if documentType.Valid {
		a.DocumentType = &documentType.String
	}
	if pdfURL.Valid {
		a.PDFURL = &pdfURL.String
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
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.ExecContext(ctx, query,
		article.Source, article.SourceID, article.UniqueKey, article.DocumentNumber, rawData, article.FetchedAt.Format("2006-01-02T15:04:05Z07:00"),
		article.Title, article.Summary, article.SourceURL, article.PublishedAt.Format("2006-01-02T15:04:05Z07:00"),
		article.DocumentType, article.PDFURL,
		article.CreatedAt.Format("2006-01-02T15:04:05Z07:00"), article.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	)
	if err != nil {
		return fmt.Errorf("failed to insert article: %w", err)
	}

	id, _ := result.LastInsertId()
	article.ID = int(id)
	return nil
}

func (r *ArticleRepository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM frarticles").Scan(&count)
	return count, err
}

func (r *ArticleRepository) GetLatest(ctx context.Context) (*models.FRArticle, error) {
	query := `
		SELECT id, document_number, raw_data, fetched_at, title, summary, source_url, published_at, document_type, pdf_url, created_at, updated_at
		FROM frarticles
		ORDER BY fetched_at DESC
		LIMIT 1
	`
	var a models.FRArticle
	var rawData []byte
	var fetchedAt, publishedAt, createdAt, updatedAt string
	var documentType, pdfURL sql.NullString
	err := r.db.QueryRowContext(ctx, query).Scan(
		&a.ID, &a.DocumentNumber, &rawData, &fetchedAt,
		&a.Title, &a.Summary, &a.SourceURL, &publishedAt,
		&documentType, &pdfURL, &createdAt, &updatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest article: %w", err)
	}
	a.FetchedAt, _ = time.Parse("2006-01-02 15:04:05Z07:00", fetchedAt)
	a.PublishedAt, _ = time.Parse("2006-01-02 15:04:05Z07:00", publishedAt)
	a.CreatedAt, _ = time.Parse("2006-01-02 15:04:05Z07:00", createdAt)
	a.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05Z07:00", updatedAt)
	if documentType.Valid {
		a.DocumentType = &documentType.String
	}
	if pdfURL.Valid {
		a.PDFURL = &pdfURL.String
	}
	json.Unmarshal(rawData, &a.RawData)
	return &a, nil
}
