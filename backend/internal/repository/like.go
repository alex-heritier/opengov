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
		FROM likes WHERE user_id = $1 AND frarticle_id = $2
	`
	var l models.Like
	err := r.db.QueryRowContext(ctx, query, userID, articleID).Scan(
		&l.ID, &l.UserID, &l.FRArticleID, &l.IsLiked, &l.CreatedAt, &l.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *LikeRepository) Toggle(ctx context.Context, userID, articleID int) (*models.Like, error) {
	now := time.Now().UTC()

	existing, err := r.GetByUserAndArticle(ctx, userID, articleID)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		query := "UPDATE likes SET is_liked = CASE WHEN is_liked = 1 THEN 0 ELSE 1 END, updated_at = $1 WHERE id = $2"
		_, err := r.db.ExecContext(ctx, query, now, existing.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to toggle like: %w", err)
		}
		existing.IsLiked = 1 - existing.IsLiked
		return existing, nil
	}

	query := `
		INSERT INTO likes (user_id, frarticle_id, is_liked, created_at, updated_at)
		VALUES ($1, $2, 1, $3, $4)
		RETURNING id
	`
	var l models.Like
	l.UserID = userID
	l.FRArticleID = articleID
	l.IsLiked = 1
	l.CreatedAt = now
	l.UpdatedAt = now

	err = r.db.QueryRowContext(ctx, query, userID, articleID, now, now).Scan(&l.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create like: %w", err)
	}
	return &l, nil
}

func (r *LikeRepository) SetLike(ctx context.Context, userID, articleID int, isPositive bool) (*models.Like, error) {
	now := time.Now().UTC()
	isLikedValue := 0
	if isPositive {
		isLikedValue = 1
	}

	existing, err := r.GetByUserAndArticle(ctx, userID, articleID)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		query := "UPDATE likes SET is_liked = $1, updated_at = $2 WHERE id = $3"
		_, err := r.db.ExecContext(ctx, query, isLikedValue, now, existing.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to update like: %w", err)
		}
		existing.IsLiked = isLikedValue
		return existing, nil
	}

	query := `
		INSERT INTO likes (user_id, frarticle_id, is_liked, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	var l models.Like
	l.UserID = userID
	l.FRArticleID = articleID
	l.IsLiked = isLikedValue
	l.CreatedAt = now
	l.UpdatedAt = now

	err = r.db.QueryRowContext(ctx, query, userID, articleID, isLikedValue, now, now).Scan(&l.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create like: %w", err)
	}
	return &l, nil
}

func (r *LikeRepository) GetArticleCounts(ctx context.Context, articleID int) (likes, dislikes int, err error) {
	query := "SELECT COUNT(*) FROM likes WHERE frarticle_id = $1 AND is_liked = 1"
	err = r.db.QueryRowContext(ctx, query, articleID).Scan(&likes)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to count likes: %w", err)
	}

	query = "SELECT COUNT(*) FROM likes WHERE frarticle_id = $1 AND is_liked = 0"
	err = r.db.QueryRowContext(ctx, query, articleID).Scan(&dislikes)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to count dislikes: %w", err)
	}

	return likes, dislikes, nil
}

func (r *LikeRepository) GetUserStatus(ctx context.Context, userID, articleID int) (status *int, err error) {
	query := "SELECT is_liked FROM likes WHERE user_id = $1 AND frarticle_id = $2"
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
	query := "DELETE FROM likes WHERE user_id = $1 AND frarticle_id = $2"
	_, err := r.db.ExecContext(ctx, query, userID, articleID)
	return err
}
