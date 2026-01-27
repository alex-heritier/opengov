package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"
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
	// TODO: This map is NOT concurrency-safe. Gin handlers run concurrently.
	// Multiple goroutines can read/write map simultaneously causing data races.
	// Fix options:
	//   1. Add sync.Mutex to protect map access
	//   2. Use sync.Map (but needs separate expiration handling)
	//   3. Move to signed cookie (stateless) or Redis (persistent, distributed)
	oauthStatesMu sync.Mutex
	oauthStates   map[string]time.Time
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

func (h *OAuthHandler) cleanupExpiredStatesLocked(now time.Time) {
	for state, timestamp := range h.oauthStates {
		if now.Sub(timestamp) > oauthStateTTL {
			delete(h.oauthStates, state)
		}
	}
}

func (h *OAuthHandler) GoogleLogin(c *gin.Context) {
	now := time.Now()
	state := generateState()
	h.oauthStatesMu.Lock()
	h.cleanupExpiredStatesLocked(now)
	h.oauthStates[state] = now
	h.oauthStatesMu.Unlock()

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
	h.oauthStatesMu.Lock()
	h.cleanupExpiredStatesLocked(time.Now())
	_, ok := h.oauthStates[state]
	if ok {
		delete(h.oauthStates, state)
	}
	h.oauthStatesMu.Unlock()
	if !ok {
		log.Printf("Invalid or expired OAuth state: %s", state)
		c.Redirect(http.StatusTemporaryRedirect, h.cfg.FrontendURL+"/login?error=invalid_state")
		return
	}

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
				CreatedAt:   time.Now().UTC(),
				UpdatedAt:   time.Now().UTC(),
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

	// Redirect to frontend callback with token in URL fragment
	// The callback page will extract the token and store it in the auth store
	c.Redirect(http.StatusTemporaryRedirect, h.cfg.FrontendURL+"/auth/callback#access_token="+jwtToken)
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
		return "", fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
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
		return nil, fmt.Errorf("userinfo request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("userinfo request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to decode userinfo response: %w", err)
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

// TestLogin handles test authentication for development environments only.
// It creates or retrieves a test user and logs them in, mimicking the Google OAuth flow
// to avoid special cases in the frontend.
func (h *OAuthHandler) TestLogin(c *gin.Context) {
	// Only allow test login in development environment
	if h.cfg.Environment != "development" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Test login is only available in development environment"})
		return
	}

	ctx := c.Request.Context()

	// Test user credentials
	testEmail := "testuser@opengov.test"
	testGoogleID := "test-google-id-12345"
	testName := "Test User"
	testPicture := "https://api.dicebear.com/7.x/avataaars/svg?seed=testuser"

	// Try to find existing test user by Google ID
	user, err := h.userRepo.GetByGoogleID(ctx, testGoogleID)
	if err != nil {
		log.Printf("Database error getting test user: %v", err)
		c.Redirect(http.StatusTemporaryRedirect, h.cfg.FrontendURL+"/login?error=test_login_error")
		return
	}

	if user == nil {
		// Check if email exists (might have been created differently)
		user, err = h.userRepo.GetByEmail(ctx, testEmail)
		if err != nil {
			log.Printf("Database error getting user by email: %v", err)
			c.Redirect(http.StatusTemporaryRedirect, h.cfg.FrontendURL+"/login?error=test_login_error")
			return
		}

		if user != nil {
			// Link Google ID to existing user
			user.GoogleID = &testGoogleID
			user.Name = &testName
			user.PictureURL = &testPicture
		} else {
			// Create new test user
			user = &models.User{
				Email:       testEmail,
				GoogleID:    &testGoogleID,
				Name:        &testName,
				PictureURL:  &testPicture,
				IsActive:    1,
				IsSuperuser: 0,
				IsVerified:  1,
				CreatedAt:   time.Now().UTC(),
				UpdatedAt:   time.Now().UTC(),
			}
			if err := h.userRepo.CreateFromGoogle(ctx, user); err != nil {
				log.Printf("Failed to create test user: %v", err)
				c.Redirect(http.StatusTemporaryRedirect, h.cfg.FrontendURL+"/login?error=test_login_error")
				return
			}
			log.Printf("Created test user with email: %s", testEmail)
		}
	}

	// Generate JWT token (same as Google OAuth flow)
	jwtToken, err := h.authService.GenerateToken(user)
	if err != nil {
		log.Printf("Failed to generate JWT token for test user: %v", err)
		c.Redirect(http.StatusTemporaryRedirect, h.cfg.FrontendURL+"/login?error=test_login_error")
		return
	}

	// Update last login time
	h.userRepo.UpdateLoginTime(ctx, user.ID)

	// Redirect to frontend callback with token in URL fragment (same as Google OAuth)
	log.Printf("Test user logged in: %s", testEmail)
	c.Redirect(http.StatusTemporaryRedirect, h.cfg.FrontendURL+"/auth/callback#access_token="+jwtToken)
}
