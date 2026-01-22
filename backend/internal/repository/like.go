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

func (r *LikeRepository) GetByUserAndArticle(ctx context.Context, userID, articleID int) (*models.Like, error) {
	query := `
		SELECT id, user_id, frarticle_id, is_liked, created_at, updated_at
		FROM likes WHERE user_id = ? AND frarticle_id = ?
	`
	var l models.Like
	var createdAt, updatedAt string
	err := r.db.QueryRowContext(ctx, query, userID, articleID).Scan(
		&l.ID, &l.UserID, &l.FRArticleID, &l.IsLiked, &createdAt, &updatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	l.CreatedAt, _ = time.Parse("2006-01-02 15:04:05Z07:00", createdAt)
	l.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05Z07:00", updatedAt)
	return &l, nil
}

func (r *LikeRepository) Toggle(ctx context.Context, userID, articleID int) (*models.Like, error) {
	now := time.Now().UTC().Format("2006-01-02T15:04:05Z07:00")

	existing, err := r.GetByUserAndArticle(ctx, userID, articleID)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		query := "UPDATE likes SET is_liked = CASE WHEN is_liked = 1 THEN 0 ELSE 1 END, updated_at = ? WHERE id = ?"
		_, err := r.db.ExecContext(ctx, query, now, existing.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to toggle like: %w", err)
		}
		existing.IsLiked = 1 - existing.IsLiked
		return existing, nil
	}

	query := `
		INSERT INTO likes (user_id, frarticle_id, is_liked, created_at, updated_at)
		VALUES (?, ?, 1, ?, ?)
	`
	var l models.Like
	l.UserID = userID
	l.FRArticleID = articleID
	l.IsLiked = 1

	result, err := r.db.ExecContext(ctx, query, userID, articleID, now, now)
	if err != nil {
		return nil, fmt.Errorf("failed to create like: %w", err)
	}
	id, _ := result.LastInsertId()
	l.ID = int(id)
	return &l, nil
}

func (r *LikeRepository) GetArticleCounts(ctx context.Context, articleID int) (likes, dislikes int, err error) {
	query := "SELECT COUNT(*) FROM likes WHERE frarticle_id = ? AND is_liked = 1"
	err = r.db.QueryRowContext(ctx, query, articleID).Scan(&likes)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to count likes: %w", err)
	}

	query = "SELECT COUNT(*) FROM likes WHERE frarticle_id = ? AND is_liked = 0"
	err = r.db.QueryRowContext(ctx, query, articleID).Scan(&dislikes)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to count dislikes: %w", err)
	}

	return likes, dislikes, nil
}

func (r *LikeRepository) GetUserStatus(ctx context.Context, userID, articleID int) (status *int, err error) {
	query := "SELECT is_liked FROM likes WHERE user_id = ? AND frarticle_id = ?"
	var isLiked int
	err = r.db.QueryRowContext(ctx, query, userID, articleID).Scan(&isLiked)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &isLiked, nil
}

func (r *LikeRepository) Remove(ctx context.Context, userID, articleID int) error {
	query := "DELETE FROM likes WHERE user_id = ? AND frarticle_id = ?"
	_, err := r.db.ExecContext(ctx, query, userID, articleID)
	return err
}
