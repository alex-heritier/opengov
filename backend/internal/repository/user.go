package repository

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/alex/opengov-go/internal/db"
	"github.com/alex/opengov-go/internal/models"
)

type UserRepository struct {
	db *db.DB
}

func NewUserRepository(db *db.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetByID(ctx context.Context, id int) (*models.User, error) {
	query := `
		SELECT id, email, hashed_password, is_active, is_superuser, is_verified,
		       google_id, name, picture_url, political_leaning, state, created_at, updated_at, last_login_at
		FROM users WHERE id = $1
	`
	var u models.User
	var lastLoginAt *time.Time
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&u.ID, &u.Email, &u.HashedPassword, &u.IsActive, &u.IsSuperuser, &u.IsVerified,
		&u.GoogleID, &u.Name, &u.PictureURL, &u.PoliticalLeaning, &u.State,
		&u.CreatedAt, &u.UpdatedAt, &lastLoginAt,
	)
	if err != nil {
		return nil, err
	}
	u.LastLoginAt = lastLoginAt
	return &u, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, email, hashed_password, is_active, is_superuser, is_verified,
		       google_id, name, picture_url, political_leaning, state, created_at, updated_at, last_login_at
		FROM users WHERE email = $1
	`
	var u models.User
	var lastLoginAt *time.Time
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&u.ID, &u.Email, &u.HashedPassword, &u.IsActive, &u.IsSuperuser, &u.IsVerified,
		&u.GoogleID, &u.Name, &u.PictureURL, &u.PoliticalLeaning, &u.State,
		&u.CreatedAt, &u.UpdatedAt, &lastLoginAt,
	)
	if err != nil {
		return nil, err
	}
	u.LastLoginAt = lastLoginAt
	return &u, nil
}

func (r *UserRepository) GetByGoogleID(ctx context.Context, googleID string) (*models.User, error) {
	query := `
		SELECT id, email, hashed_password, is_active, is_superuser, is_verified,
		       google_id, name, picture_url, political_leaning, state, created_at, updated_at, last_login_at
		FROM users WHERE google_id = $1
	`
	var u models.User
	var lastLoginAt *time.Time
	err := r.db.QueryRowContext(ctx, query, googleID).Scan(
		&u.ID, &u.Email, &u.HashedPassword, &u.IsActive, &u.IsSuperuser, &u.IsVerified,
		&u.GoogleID, &u.Name, &u.PictureURL, &u.PoliticalLeaning, &u.State,
		&u.CreatedAt, &u.UpdatedAt, &lastLoginAt,
	)
	if err != nil {
		return nil, err
	}
	u.LastLoginAt = lastLoginAt
	return &u, nil
}

func (r *UserRepository) Create(ctx context.Context, user *models.User, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	now := time.Now().UTC()
	user.CreatedAt = now
	user.UpdatedAt = now
	user.IsActive = 1
	user.IsSuperuser = 0
	user.IsVerified = 0

	query := `
		INSERT INTO users (email, hashed_password, is_active, is_superuser, is_verified, google_id, name, picture_url, political_leaning, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`
	err = r.db.QueryRowContext(ctx, query,
		user.Email, string(hashedPassword), user.IsActive, user.IsSuperuser, user.IsVerified,
		user.GoogleID, user.Name, user.PictureURL, user.PoliticalLeaning,
		user.CreatedAt, user.UpdatedAt,
	).Scan(&user.ID)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}
	return nil
}

func (r *UserRepository) CreateFromGoogle(ctx context.Context, user *models.User) error {
	now := time.Now().UTC()
	user.CreatedAt = now
	user.UpdatedAt = now
	user.IsActive = 1
	user.IsSuperuser = 0
	user.IsVerified = 1

	query := `
		INSERT INTO users (email, hashed_password, is_active, is_superuser, is_verified, google_id, name, picture_url, political_leaning, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`
	err := r.db.QueryRowContext(ctx, query,
		user.Email, "", user.IsActive, user.IsSuperuser, user.IsVerified,
		user.GoogleID, user.Name, user.PictureURL, user.PoliticalLeaning,
		user.CreatedAt, user.UpdatedAt,
	).Scan(&user.ID)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}
	return nil
}

func (r *UserRepository) UpdateLoginTime(ctx context.Context, id int) error {
	query := "UPDATE users SET last_login_at = $1 WHERE id = $2"
	_, err := r.db.ExecContext(ctx, query, time.Now().UTC(), id)
	return err
}

func (r *UserRepository) VerifyPassword(user *models.User, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password))
	return err == nil
}

func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users SET
			name = $1, picture_url = $2, political_leaning = $3, state = $4, updated_at = $5
		WHERE id = $6
	`
	_, err := r.db.ExecContext(ctx, query,
		user.Name, user.PictureURL, user.PoliticalLeaning, user.State,
		time.Now().UTC(), user.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}
