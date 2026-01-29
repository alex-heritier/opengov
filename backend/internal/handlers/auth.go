package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/alex/opengov-go/internal/domain"
	"github.com/alex/opengov-go/internal/middleware"
	"github.com/alex/opengov-go/internal/repository"
	"github.com/alex/opengov-go/internal/services"
	"github.com/alex/opengov-go/internal/timeformat"
	"github.com/alex/opengov-go/internal/transport"
)

type AuthHandler struct {
	authService *services.AuthService
	userRepo    *repository.UserRepository
}

func NewAuthHandler(authService *services.AuthService, userRepo *repository.UserRepository) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		userRepo:    userRepo,
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req transport.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	user, err := h.authService.Authenticate(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, err := h.authService.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, transport.AuthResponse{
		AccessToken: token,
		User:        userToResponse(user),
	})
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req transport.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	existing, _ := h.userRepo.GetByEmail(c.Request.Context(), req.Email)
	if existing != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}

	user := &domain.User{
		Email: req.Email,
		Name:  strPtr(req.Name),
	}
	if err := h.userRepo.Create(c.Request.Context(), user, req.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	token, err := h.authService.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, transport.AuthResponse{
		AccessToken: token,
		User:        userToResponse(user),
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID, hasAuth := middleware.GetUserID(c)
	if !hasAuth {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	user, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, userToResponse(user))
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	userID, hasAuth := middleware.GetUserID(c)
	if !hasAuth {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	user, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	token, err := h.authService.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"access_token": token})
}

func (h *AuthHandler) UpdateUser(c *gin.Context) {
	userID, hasAuth := middleware.GetUserID(c)
	if !hasAuth {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	user, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var req transport.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if req.Name != nil {
		user.Name = req.Name
	}
	if req.PictureURL != nil {
		user.PictureURL = req.PictureURL
	}
	if req.PoliticalLeaning != nil {
		user.PoliticalLeaning = req.PoliticalLeaning
	}
	if req.State != nil {
		user.State = req.State
	}

	if err := h.userRepo.Update(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, userToResponse(user))
}

func userToResponse(u *domain.User) *transport.UserResponse {
	var lastLoginAt *string
	if u.LastLoginAt != nil {
		s := u.LastLoginAt.Format(timeformat.RFC3339)
		lastLoginAt = &s
	}
	return &transport.UserResponse{
		ID:               u.ID,
		Email:            u.Email,
		Name:             u.Name,
		PictureURL:       u.PictureURL,
		GoogleID:         u.GoogleID,
		PoliticalLeaning: u.PoliticalLeaning,
		State:            u.State,
		IsActive:         u.GetIsActive(),
		IsVerified:       u.GetIsVerified(),
		CreatedAt:        u.CreatedAt.Format(timeformat.RFC3339),
		UpdatedAt:        u.UpdatedAt.Format(timeformat.RFC3339),
		LastLoginAt:      lastLoginAt,
	}
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
