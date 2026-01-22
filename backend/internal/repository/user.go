package repository

import (
	"context"
	"database/sql"
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
		       google_id, name, picture_url, political_leaning, created_at, updated_at, last_login_at
		FROM users WHERE id = ?
	`
	var u models.User
	var lastLoginAt sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&u.ID, &u.Email, &u.HashedPassword, &u.IsActive, &u.IsSuperuser, &u.IsVerified,
		&u.GoogleID, &u.Name, &u.PictureURL, &u.PoliticalLeaning,
		&u.CreatedAt, &u.UpdatedAt, &lastLoginAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if lastLoginAt.Valid {
		t, _ := time.Parse("2006-01-02 15:04:05Z07:00", lastLoginAt.String)
		u.LastLoginAt = &t
	}
	return &u, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, email, hashed_password, is_active, is_superuser, is_verified,
		       google_id, name, picture_url, political_leaning, created_at, updated_at, last_login_at
		FROM users WHERE email = ?
	`
	var u models.User
	var lastLoginAt sql.NullString
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&u.ID, &u.Email, &u.HashedPassword, &u.IsActive, &u.IsSuperuser, &u.IsVerified,
		&u.GoogleID, &u.Name, &u.PictureURL, &u.PoliticalLeaning,
		&u.CreatedAt, &u.UpdatedAt, &lastLoginAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	if lastLoginAt.Valid {
		t, _ := time.Parse("2006-01-02T15:04:05Z07:00", lastLoginAt.String)
		u.LastLoginAt = &t
	}
	return &u, nil
}

func (r *UserRepository) GetByGoogleID(ctx context.Context, googleID string) (*models.User, error) {
	query := `
		SELECT id, email, hashed_password, is_active, is_superuser, is_verified,
		       google_id, name, picture_url, political_leaning, created_at, updated_at, last_login_at
		FROM users WHERE google_id = ?
	`
	var u models.User
	var lastLoginAt sql.NullString
	err := r.db.QueryRowContext(ctx, query, googleID).Scan(
		&u.ID, &u.Email, &u.HashedPassword, &u.IsActive, &u.IsSuperuser, &u.IsVerified,
		&u.GoogleID, &u.Name, &u.PictureURL, &u.PoliticalLeaning,
		&u.CreatedAt, &u.UpdatedAt, &lastLoginAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by google id: %w", err)
	}
	if lastLoginAt.Valid {
		t, _ := time.Parse("2006-01-02 15:04:05Z07:00", lastLoginAt.String)
		u.LastLoginAt = &t
	}
	return &u, nil
}

func (r *UserRepository) Create(ctx context.Context, user *models.User, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	now := time.Now().UTC()
	nowStr := now.Format("2006-01-02T15:04:05Z07:00")
	user.CreatedAt = nowStr
	user.UpdatedAt = nowStr
	user.IsActive = 1
	user.IsSuperuser = 0
	user.IsVerified = 0

	query := `
		INSERT INTO users (email, hashed_password, is_active, is_superuser, is_verified, google_id, name, picture_url, political_leaning, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.ExecContext(ctx, query,
		user.Email, string(hashedPassword), user.IsActive, user.IsSuperuser, user.IsVerified,
		user.GoogleID, user.Name, user.PictureURL, user.PoliticalLeaning,
		user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}
	id, _ := result.LastInsertId()
	user.ID = int(id)
	return nil
}

func (r *UserRepository) CreateFromGoogle(ctx context.Context, user *models.User) error {
	now := time.Now().UTC()
	nowStr := now.Format("2006-01-02T15:04:05Z07:00")
	user.CreatedAt = nowStr
	user.UpdatedAt = nowStr
	user.IsActive = 1
	user.IsSuperuser = 0
	user.IsVerified = 1

	query := `
		INSERT INTO users (email, hashed_password, is_active, is_superuser, is_verified, google_id, name, picture_url, political_leaning, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.ExecContext(ctx, query,
		user.Email, "", user.IsActive, user.IsSuperuser, user.IsVerified,
		user.GoogleID, user.Name, user.PictureURL, user.PoliticalLeaning,
		user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}
	id, _ := result.LastInsertId()
	user.ID = int(id)
	return nil
}

func (r *UserRepository) UpdateLoginTime(ctx context.Context, id int) error {
	query := "UPDATE users SET last_login_at = ? WHERE id = ?"
	_, err := r.db.ExecContext(ctx, query, time.Now().UTC().Format("2006-01-02T15:04:05Z07:00"), id)
	return err
}

func (r *UserRepository) VerifyPassword(user *models.User, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password))
	return err == nil
}
