package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/alex/opengov-go/internal/config"
	"github.com/alex/opengov-go/internal/models"
	"github.com/alex/opengov-go/internal/repository"
	"github.com/alex/opengov-go/internal/services"
)

type OAuthHandler struct {
	authService *services.AuthService
	userRepo    *repository.UserRepository
	cfg         *config.Config
	// In-memory state store (use Redis in production)
	oauthStates map[string]time.Time
}

const oauthStateTTL = 10 * time.Minute

func NewOAuthHandler(authService *services.AuthService, userRepo *repository.UserRepository, cfg *config.Config) *OAuthHandler {
	return &OAuthHandler{
		authService: authService,
		userRepo:    userRepo,
		cfg:         cfg,
		oauthStates: make(map[string]time.Time),
	}
}

func (h *OAuthHandler) cleanupExpiredStates() {
	now := time.Now()
	for state, timestamp := range h.oauthStates {
		if now.Sub(timestamp) > oauthStateTTL {
			delete(h.oauthStates, state)
		}
	}
}

func (h *OAuthHandler) GoogleLogin(c *gin.Context) {
	h.cleanupExpiredStates()
	state := generateState()
	h.oauthStates[state] = time.Now()

	authURL := fmt.Sprintf(
		"https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=email%%20profile&state=%s",
		h.cfg.GoogleClientID,
		url.QueryEscape(h.cfg.GoogleRedirectURI),
		state,
	)

	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

func (h *OAuthHandler) GoogleCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	// Validate state
	if _, ok := h.oauthStates[state]; !ok {
		log.Printf("Invalid or expired OAuth state: %s", state)
		c.Redirect(http.StatusTemporaryRedirect, h.cfg.FrontendURL+"/login?error=invalid_state")
		return
	}
	delete(h.oauthStates, state)

	// Exchange code for token
	token, err := exchangeGoogleToken(code, h.cfg)
	if err != nil {
		log.Printf("Google OAuth token exchange failed: %v", err)
		c.Redirect(http.StatusTemporaryRedirect, h.cfg.FrontendURL+"/login?error=token_exchange_failed")
		return
	}

	// Get user info
	userInfo, err := getGoogleUserInfo(token, h.cfg)
	if err != nil {
		log.Printf("Failed to get Google user info: %v", err)
		c.Redirect(http.StatusTemporaryRedirect, h.cfg.FrontendURL+"/login?error=invalid_response")
		return
	}

	googleID, ok := userInfo["sub"].(string)
	if !ok {
		log.Printf("No Google ID in user info")
		c.Redirect(http.StatusTemporaryRedirect, h.cfg.FrontendURL+"/login?error=invalid_user_info")
		return
	}

	email, _ := userInfo["email"].(string)

	// Find or create user
	ctx := c.Request.Context()
	user, err := h.userRepo.GetByGoogleID(ctx, googleID)
	if err != nil {
		log.Printf("Database error getting user by Google ID: %v", err)
		c.Redirect(http.StatusTemporaryRedirect, h.cfg.FrontendURL+"/login?error=oauth_error")
		return
	}

	if user == nil {
		// Check if email exists
		user, err = h.userRepo.GetByEmail(ctx, email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		if user != nil {
			// Link Google account
			user.GoogleID = &googleID
			name, _ := userInfo["name"].(string)
			user.Name = &name
			picture, _ := userInfo["picture"].(string)
			user.PictureURL = &picture
		} else {
			// Create new user
			name, _ := userInfo["name"].(string)
			picture, _ := userInfo["picture"].(string)
			verified, _ := userInfo["email_verified"].(bool)

			user = &models.User{
				Email:       email,
				GoogleID:    &googleID,
				Name:        &name,
				PictureURL:  &picture,
				IsActive:    1,
				IsSuperuser: 0,
				IsVerified:  map[bool]int{true: 1, false: 0}[verified],
				CreatedAt:   time.Now().UTC().Format("2006-01-02T15:04:05Z07:00"),
				UpdatedAt:   time.Now().UTC().Format("2006-01-02T15:04:05Z07:00"),
			}
			if err := h.userRepo.CreateFromGoogle(ctx, user); err != nil {
				log.Printf("Failed to create user from Google OAuth: %v", err)
				c.Redirect(http.StatusTemporaryRedirect, h.cfg.FrontendURL+"/login?error=oauth_error")
				return
			}
		}
	} else {
		// Update profile
		name, _ := userInfo["name"].(string)
		user.Name = &name
		picture, _ := userInfo["picture"].(string)
		user.PictureURL = &picture
	}

	// Generate JWT token
	jwtToken, err := h.authService.GenerateToken(user)
	if err != nil {
		log.Printf("Failed to generate JWT token: %v", err)
		c.Redirect(http.StatusTemporaryRedirect, h.cfg.FrontendURL+"/login?error=oauth_error")
		return
	}

	// Set auth cookie and redirect to frontend (matching Python behavior)
	c.SetCookie("opengov_auth", jwtToken, h.cfg.JWTAccessTokenExpireMin*60, "/", "", h.cfg.CookieSecure, true)
	c.Redirect(307, h.cfg.FrontendURL+"/feed")
}

func generateState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func exchangeGoogleToken(code string, cfg *config.Config) (string, error) {
	data := url.Values{
		"code":          {code},
		"client_id":     {cfg.GoogleClientID},
		"client_secret": {cfg.GoogleClientSecret},
		"redirect_uri":  {cfg.GoogleRedirectURI},
		"grant_type":    {"authorization_code"},
	}

	resp, err := http.PostForm("https://oauth2.googleapis.com/token", data)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if errStr, ok := result["error"].(string); ok {
		return "", fmt.Errorf("google error: %s", errStr)
	}

	accessToken, ok := result["access_token"].(string)
	if !ok {
		return "", fmt.Errorf("no access_token in response")
	}

	return accessToken, nil
}

func getGoogleUserInfo(accessToken string, cfg *config.Config) (map[string]interface{}, error) {
	req, _ := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

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

func userToAuthResponse(u *models.User) *AuthUserResponse {
	var lastLoginAt *string
	if u.LastLoginAt != nil {
		s := u.LastLoginAt.Format("2006-01-02T15:04:05Z07:00")
		lastLoginAt = &s
	}
	return &AuthUserResponse{
		ID:               u.ID,
		Email:            u.Email,
		Name:             u.Name,
		PictureURL:       u.PictureURL,
		GoogleID:         u.GoogleID,
		PoliticalLeaning: u.PoliticalLeaning,
		IsActive:         u.GetIsActive(),
		IsVerified:       u.GetIsVerified(),
		CreatedAt:        u.CreatedAt,
		UpdatedAt:        u.UpdatedAt,
		LastLoginAt:      lastLoginAt,
	}
}
