package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/alex/opengov-go/internal/db"
	"github.com/alex/opengov-go/internal/models"
)

type LikeRepository struct {
	db *db.DB
}

func NewLikeRepository(db *db.DB) *LikeRepository {
	return &LikeRepository{db: db}
}

func (r *LikeRepository) GetByUserAndFeedEntry(ctx context.Context, userID, feedEntryID int) (*models.Like, error) {
	query := `
		SELECT id, user_id, feed_entry_id, value, created_at, updated_at
		FROM likes WHERE user_id = $1 AND feed_entry_id = $2
	`
	var l models.Like
	err := r.db.QueryRowContext(ctx, query, userID, feedEntryID).Scan(
		&l.ID, &l.UserID, &l.FeedEntryID, &l.Value, &l.CreatedAt, &l.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *LikeRepository) SetValue(ctx context.Context, userID, feedEntryID int, value int) (*models.Like, error) {
	existing, err := r.GetByUserAndFeedEntry(ctx, userID, feedEntryID)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		query := "UPDATE likes SET value = $1, updated_at = NOW() WHERE id = $2"
		_, err := r.db.ExecContext(ctx, query, value, existing.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to update like: %w", err)
		}
		existing.Value = value
		return existing, nil
	}

	query := `
		INSERT INTO likes (user_id, feed_entry_id, value)
		VALUES ($1, $2, $3)
		RETURNING id
	`
	var l models.Like
	l.UserID = userID
	l.FeedEntryID = feedEntryID
	l.Value = value
	l.CreatedAt = time.Now().UTC()
	l.UpdatedAt = l.CreatedAt

	err = r.db.QueryRowContext(ctx, query, userID, feedEntryID, value).Scan(&l.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create like: %w", err)
	}
	return &l, nil
}

func (r *LikeRepository) GetFeedEntryCounts(ctx context.Context, feedEntryID int) (likes, dislikes int, err error) {
	query := "SELECT COUNT(*) FROM likes WHERE feed_entry_id = $1 AND value = 1"
	err = r.db.QueryRowContext(ctx, query, feedEntryID).Scan(&likes)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to count likes: %w", err)
	}

	query = "SELECT COUNT(*) FROM likes WHERE feed_entry_id = $1 AND value = -1"
	err = r.db.QueryRowContext(ctx, query, feedEntryID).Scan(&dislikes)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to count dislikes: %w", err)
	}

	return likes, dislikes, nil
}

func (r *LikeRepository) GetUserStatus(ctx context.Context, userID, feedEntryID int) (status *int, err error) {
	query := "SELECT value FROM likes WHERE user_id = $1 AND feed_entry_id = $2"
	var value int
	err = r.db.QueryRowContext(ctx, query, userID, feedEntryID).Scan(&value)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func (r *LikeRepository) Remove(ctx context.Context, userID, feedEntryID int) error {
	query := "DELETE FROM likes WHERE user_id = $1 AND feed_entry_id = $2"
	_, err := r.db.ExecContext(ctx, query, userID, feedEntryID)
	return err
}
