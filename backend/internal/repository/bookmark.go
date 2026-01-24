package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/alex/opengov-go/internal/db"
	"github.com/alex/opengov-go/internal/models"
)

type BookmarkRepository struct {
	db *db.DB
}

func NewBookmarkRepository(db *db.DB) *BookmarkRepository {
	return &BookmarkRepository{db: db}
}

func (r *BookmarkRepository) GetByUserAndArticle(ctx context.Context, userID, articleID int) (*models.Bookmark, error) {
	query := `
		SELECT id, user_id, frarticle_id, is_bookmarked, created_at, updated_at
		FROM bookmarks WHERE user_id = $1 AND frarticle_id = $2
	`
	var b models.Bookmark
	err := r.db.QueryRowContext(ctx, query, userID, articleID).Scan(
		&b.ID, &b.UserID, &b.FRArticleID, &b.IsBookmarked, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *BookmarkRepository) GetBookmarkedArticles(ctx context.Context, userID int) ([]models.FRArticle, error) {
	query := `
		SELECT a.id, a.source, a.source_id, a.unique_key, a.document_number, a.raw_data, a.fetched_at, a.title, a.summary, a.source_url, a.published_at, a.document_type, a.pdf_url, a.created_at, a.updated_at
		FROM frarticles a
		JOIN bookmarks b ON a.id = b.frarticle_id
		WHERE b.user_id = $1 AND b.is_bookmarked = 1
		ORDER BY b.created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query bookmarked articles: %w", err)
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
			return nil, fmt.Errorf("failed to scan article: %w", err)
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
	return articles, nil
}

func (r *BookmarkRepository) Toggle(ctx context.Context, userID, articleID int) (*models.Bookmark, error) {
	now := time.Now().UTC()

	existing, err := r.GetByUserAndArticle(ctx, userID, articleID)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		query := "UPDATE bookmarks SET is_bookmarked = CASE WHEN is_bookmarked = 1 THEN 0 ELSE 1 END, updated_at = $1 WHERE id = $2"
		_, err := r.db.ExecContext(ctx, query, now, existing.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to toggle bookmark: %w", err)
		}
		existing.IsBookmarked = 1 - existing.IsBookmarked
		return existing, nil
	}

	query := `
		INSERT INTO bookmarks (user_id, frarticle_id, is_bookmarked, created_at, updated_at)
		VALUES ($1, $2, 1, $3, $4)
		RETURNING id
	`
	var b models.Bookmark
	b.UserID = userID
	b.FRArticleID = articleID
	b.IsBookmarked = 1
	b.CreatedAt = now
	b.UpdatedAt = now

	err = r.db.QueryRowContext(ctx, query, userID, articleID, now, now).Scan(&b.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create bookmark: %w", err)
	}
	return &b, nil
}

func (r *BookmarkRepository) Remove(ctx context.Context, userID, articleID int) error {
	query := "UPDATE bookmarks SET is_bookmarked = 0, updated_at = $1 WHERE user_id = $2 AND frarticle_id = $3"
	_, err := r.db.ExecContext(ctx, query, time.Now().UTC(), userID, articleID)
	return err
}

func (r *BookmarkRepository) IsBookmarked(ctx context.Context, userID, articleID int) (bool, error) {
	query := "SELECT COUNT(*) FROM bookmarks WHERE user_id = $1 AND frarticle_id = $2 AND is_bookmarked = 1"
	var count int
	err := r.db.QueryRowContext(ctx, query, userID, articleID).Scan(&count)
	return count > 0, err
}
