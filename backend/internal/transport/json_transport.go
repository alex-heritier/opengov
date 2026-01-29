package transport

import (
	"time"

	"github.com/alex/opengov-go/internal/domain"
)

// JSON (HTTP) transport DTOs live here. They may contain json/binding tags.

// ScrapedPolicyDocument is an upstream document payload returned by a scraper.
// It is intentionally separate from the DB-backed domain models.
type ScrapedPolicyDocument struct {
	DocumentNumber         string
	Title                  string
	Type                   string
	Abstract               *string
	HTMLURL                string
	PublicationDate        string
	PDFURL                 *string
	PublicInspectionPDFURL *string
	Excerpts               *string
	Agencies               []domain.Agency
}

// Auth
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name,omitempty"`
}

type AuthResponse struct {
	AccessToken string        `json:"access_token"`
	User        *UserResponse `json:"user"`
}

type UserResponse struct {
	ID               int     `json:"id"`
	Email            string  `json:"email"`
	Name             *string `json:"name,omitempty"`
	PictureURL       *string `json:"picture_url,omitempty"`
	GoogleID         *string `json:"google_id,omitempty"`
	PoliticalLeaning *string `json:"political_leaning,omitempty"`
	State            *string `json:"state,omitempty"`
	IsActive         bool    `json:"is_active"`
	IsVerified       bool    `json:"is_verified"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
	LastLoginAt      *string `json:"last_login_at,omitempty"`
}

type UpdateUserRequest struct {
	Name             *string `json:"name,omitempty"`
	PictureURL       *string `json:"picture_url,omitempty"`
	PoliticalLeaning *string `json:"political_leaning,omitempty"`
	State            *string `json:"state,omitempty"`
}

// OAuth (currently only used for docs / historical reference)
type AuthUserResponse struct {
	ID               int     `json:"id"`
	Email            string  `json:"email"`
	Name             *string `json:"name,omitempty"`
	PictureURL       *string `json:"picture_url,omitempty"`
	GoogleID         *string `json:"google_id,omitempty"`
	PoliticalLeaning *string `json:"political_leaning,omitempty"`
	IsActive         bool    `json:"is_active"`
	IsVerified       bool    `json:"is_verified"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
	LastLoginAt      *string `json:"last_login_at,omitempty"`
}

// Likes
type ToggleLikeRequest struct {
	Value int `json:"value"`
}

// Feed
type FeedEntryResponse struct {
	ID             int      `json:"id"`
	Title          string   `json:"title"`
	Summary        string   `json:"summary"`
	Keypoints      []string `json:"keypoints,omitempty"`
	ImpactScore    *string  `json:"impact_score,omitempty"`
	PoliticalScore *int     `json:"political_score,omitempty"`
	SourceURL      string   `json:"source_url"`
	PublishedAt    string   `json:"published_at"`
	IsBookmarked   *bool    `json:"is_bookmarked,omitempty"`
	UserLikeStatus *int     `json:"user_like_status,omitempty"`
	LikesCount     int      `json:"likes_count"`
	DislikesCount  int      `json:"dislikes_count"`
}

type FeedResponse struct {
	Items   []FeedEntryResponse `json:"items"`
	Page    int                 `json:"page"`
	Limit   int                 `json:"limit"`
	Total   int                 `json:"total"`
	HasNext bool                `json:"has_next"`
}

// Admin
type StatsResponse struct {
	TotalArticles  int        `json:"total_articles"`
	LastScrapeTime *time.Time `json:"last_scrape_time,omitempty"`
	LastScrapeAge  string     `json:"last_scrape_human,omitempty"`
}
