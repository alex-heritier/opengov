package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/alex/opengov-go/internal/config"
	"github.com/alex/opengov-go/internal/domain"
	"github.com/alex/opengov-go/internal/repository"
)

type AuthService struct {
	jwtSecret string
	jwtExpiry time.Duration
	userRepo  *repository.UserRepository
}

type Claims struct {
	UserID      int64  `json:"user_id"`
	Email       string `json:"email"`
	IsSuperuser bool   `json:"is_superuser"`
	jwt.RegisteredClaims
}

func NewAuthService(cfg *config.Config, userRepo *repository.UserRepository) *AuthService {
	return &AuthService{
		jwtSecret: cfg.JWTSecretKey,
		jwtExpiry: time.Duration(cfg.JWTAccessTokenExpireMin) * time.Minute,
		userRepo:  userRepo,
	}
}

func (s *AuthService) GenerateToken(user *domain.User) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:      user.ID,
		Email:       user.Email,
		IsSuperuser: user.GetIsSuperuser(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.jwtExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Hardcoded: we only accept HS256 tokens.
		if token.Method == nil || token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func (s *AuthService) Authenticate(ctx context.Context, email, password string) (*domain.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	if !user.GetIsActive() {
		return nil, errors.New("user is inactive")
	}

	if !s.userRepo.VerifyPassword(user, password) {
		return nil, errors.New("invalid password")
	}

	if err := s.userRepo.UpdateLoginTime(ctx, user.ID); err != nil {
		return nil, fmt.Errorf("failed to update login time: %w", err)
	}

	return user, nil
}

func (s *AuthService) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, id)
}
