package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/alex/opengov-go/internal/db"
	"github.com/alex/opengov-go/internal/models"
	"github.com/alex/opengov-go/internal/timeformat"
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
		FROM bookmarks WHERE user_id = ? AND frarticle_id = ?
	`
	var b models.Bookmark
	var createdAt, updatedAt string
	err := r.db.QueryRowContext(ctx, query, userID, articleID).Scan(
		&b.ID, &b.UserID, &b.FRArticleID, &b.IsBookmarked, &createdAt, &updatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	b.CreatedAt, _ = time.Parse(timeformat.DBTime, createdAt)
	b.UpdatedAt, _ = time.Parse(timeformat.DBTime, updatedAt)
	return &b, nil
}

func (r *BookmarkRepository) GetBookmarkedArticles(ctx context.Context, userID int) ([]models.FRArticle, error) {
	query := `
		SELECT a.id, a.document_number, a.raw_data, a.fetched_at, a.title, a.summary, a.source_url, a.published_at, a.created_at, a.updated_at
		FROM frarticles a
		JOIN bookmarks b ON a.id = b.frarticle_id
		WHERE b.user_id = ? AND b.is_bookmarked = 1
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
		var fetchedAt, publishedAt, createdAt, updatedAt string
		err := rows.Scan(
			&a.ID, &a.DocumentNumber, &rawData, &fetchedAt,
			&a.Title, &a.Summary, &a.SourceURL, &publishedAt,
			&createdAt, &updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan article: %w", err)
		}
		a.FetchedAt, _ = time.Parse(timeformat.DBTime, fetchedAt)
		a.PublishedAt, _ = time.Parse(timeformat.DBTime, publishedAt)
		a.CreatedAt, _ = time.Parse(timeformat.DBTime, createdAt)
		a.UpdatedAt, _ = time.Parse(timeformat.DBTime, updatedAt)
		json.Unmarshal(rawData, &a.RawData)
		articles = append(articles, a)
	}
	return articles, nil
}

func (r *BookmarkRepository) Toggle(ctx context.Context, userID, articleID int) (*models.Bookmark, error) {
	now := time.Now().UTC().Format(timeformat.DBTime)

	existing, err := r.GetByUserAndArticle(ctx, userID, articleID)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		query := "UPDATE bookmarks SET is_bookmarked = CASE WHEN is_bookmarked = 1 THEN 0 ELSE 1 END, updated_at = ? WHERE id = ?"
		_, err := r.db.ExecContext(ctx, query, now, existing.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to toggle bookmark: %w", err)
		}
		existing.IsBookmarked = 1 - existing.IsBookmarked
		return existing, nil
	}

	query := `
		INSERT INTO bookmarks (user_id, frarticle_id, is_bookmarked, created_at, updated_at)
		VALUES (?, ?, 1, ?, ?)
	`
	var b models.Bookmark
	b.UserID = userID
	b.FRArticleID = articleID
	b.IsBookmarked = 1

	result, err := r.db.ExecContext(ctx, query, userID, articleID, now, now)
	if err != nil {
		return nil, fmt.Errorf("failed to create bookmark: %w", err)
	}
	id, _ := result.LastInsertId()
	b.ID = int(id)
	return &b, nil
}

func (r *BookmarkRepository) Remove(ctx context.Context, userID, articleID int) error {
	query := "UPDATE bookmarks SET is_bookmarked = 0, updated_at = ? WHERE user_id = ? AND frarticle_id = ?"
	_, err := r.db.ExecContext(ctx, query, time.Now().UTC().Format(timeformat.DBTime), userID, articleID)
	return err
}

func (r *BookmarkRepository) IsBookmarked(ctx context.Context, userID, articleID int) (bool, error) {
	query := "SELECT COUNT(*) FROM bookmarks WHERE user_id = ? AND frarticle_id = ? AND is_bookmarked = 1"
	var count int
	err := r.db.QueryRowContext(ctx, query, userID, articleID).Scan(&count)
	return count > 0, err
}
