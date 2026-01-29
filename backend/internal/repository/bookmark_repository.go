package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/alex/opengov-go/internal/db"
	"github.com/alex/opengov-go/internal/domain"
)

type BookmarkRepository struct {
	db *db.DB
}

func NewBookmarkRepository(db *db.DB) *BookmarkRepository {
	return &BookmarkRepository{db: db}
}

func (r *BookmarkRepository) GetByUserAndFeedEntry(ctx context.Context, userID, feedEntryID int64) (*domain.Bookmark, error) {
	query := `
		SELECT id, user_id, feed_entry_id, created_at, updated_at
		FROM bookmarks WHERE user_id = $1 AND feed_entry_id = $2
	`
	var b domain.Bookmark
	err := r.db.QueryRowContext(ctx, query, userID, feedEntryID).Scan(
		&b.ID, &b.UserID, &b.FeedEntryID, &b.CreatedAt, &b.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *BookmarkRepository) Toggle(ctx context.Context, userID, feedEntryID int64) (bool, error) {
	existing, err := r.GetByUserAndFeedEntry(ctx, userID, feedEntryID)
	if err != nil {
		return false, err
	}

	if existing != nil {
		query := "DELETE FROM bookmarks WHERE user_id = $1 AND feed_entry_id = $2"
		_, err := r.db.ExecContext(ctx, query, userID, feedEntryID)
		if err != nil {
			return false, fmt.Errorf("failed to remove bookmark: %w", err)
		}
		return false, nil
	}

	query := `
		INSERT INTO bookmarks (user_id, feed_entry_id)
		VALUES ($1, $2)
	`
	_, err = r.db.ExecContext(ctx, query, userID, feedEntryID)
	if err != nil {
		return false, fmt.Errorf("failed to create bookmark: %w", err)
	}
	return true, nil
}

func (r *BookmarkRepository) Remove(ctx context.Context, userID, feedEntryID int64) error {
	query := "DELETE FROM bookmarks WHERE user_id = $1 AND feed_entry_id = $2"
	_, err := r.db.ExecContext(ctx, query, userID, feedEntryID)
	return err
}

func (r *BookmarkRepository) IsBookmarked(ctx context.Context, userID, feedEntryID int64) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM bookmarks WHERE user_id = $1 AND feed_entry_id = $2)"
	var exists bool
	err := r.db.QueryRowContext(ctx, query, userID, feedEntryID).Scan(&exists)
	return exists, err
}

func (r *BookmarkRepository) GetBookmarkIDsByUser(ctx context.Context, userID int64) ([]int64, error) {
	query := "SELECT feed_entry_id FROM bookmarks WHERE user_id = $1 ORDER BY created_at DESC"
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query bookmarked feed items: %w", err)
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan bookmark id: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}
