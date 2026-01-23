package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/alex/opengov-go/internal/middleware"
	"github.com/alex/opengov-go/internal/repository"
)

type LikeHandler struct {
	likeRepo    *repository.LikeRepository
	articleRepo *repository.ArticleRepository
}

func NewLikeHandler(likeRepo *repository.LikeRepository, articleRepo *repository.ArticleRepository) *LikeHandler {
	return &LikeHandler{
		likeRepo:    likeRepo,
		articleRepo: articleRepo,
	}
}

type ToggleLikeRequest struct {
	IsPositive bool `json:"is_positive"`
}

func (h *LikeHandler) Toggle(c *gin.Context) {
	userID, hasAuth := middleware.GetUserID(c)
	if !hasAuth {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	articleID, err := strconv.Atoi(c.Param("article_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
		return
	}

	var req ToggleLikeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	_, err = h.articleRepo.GetByID(c.Request.Context(), articleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check article"})
		return
	}

	like, err := h.likeRepo.SetLike(c.Request.Context(), userID, articleID, req.IsPositive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set like"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"is_liked": like.GetIsLiked(),
	})
}

func (h *LikeHandler) GetCounts(c *gin.Context) {
	articleID, err := strconv.Atoi(c.Param("article_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
		return
	}

	likes, dislikes, err := h.likeRepo.GetArticleCounts(c.Request.Context(), articleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get counts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"likes":    likes,
		"dislikes": dislikes,
	})
}

func (h *LikeHandler) Remove(c *gin.Context) {
	userID, hasAuth := middleware.GetUserID(c)
	if !hasAuth {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	articleID, err := strconv.Atoi(c.Param("article_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
		return
	}

	err = h.likeRepo.Remove(c.Request.Context(), userID, articleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove like"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Like removed",
	})
}

func (h *LikeHandler) GetStatus(c *gin.Context) {
	userID, hasAuth := middleware.GetUserID(c)
	if !hasAuth {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	articleID, err := strconv.Atoi(c.Param("article_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
		return
	}

	status, err := h.likeRepo.GetUserStatus(c.Request.Context(), userID, articleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get like status"})
		return
	}

	if status == nil {
		c.JSON(http.StatusOK, gin.H{"is_positive": nil})
		return
	}

	isPositive := *status == 1
	c.JSON(http.StatusOK, gin.H{"is_positive": isPositive})
}
