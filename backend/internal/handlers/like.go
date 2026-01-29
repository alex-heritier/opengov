package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/alex/opengov-go/internal/middleware"
	"github.com/alex/opengov-go/internal/repository"
	"github.com/alex/opengov-go/internal/transport"
)

type LikeHandler struct {
	likeRepo *repository.LikeRepository
}

func NewLikeHandler(likeRepo *repository.LikeRepository) *LikeHandler {
	return &LikeHandler{
		likeRepo: likeRepo,
	}
}

func (h *LikeHandler) Toggle(c *gin.Context) {
	userID, hasAuth := middleware.GetUserID(c)
	if !hasAuth {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	feedEntryID, err := strconv.Atoi(c.Param("feed_entry_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feed entry ID"})
		return
	}

	var req transport.ToggleLikeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	if req.Value != 1 && req.Value != -1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "value must be 1 or -1"})
		return
	}

	like, err := h.likeRepo.SetValue(c.Request.Context(), userID, feedEntryID, req.Value)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set like"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"value": like.Value,
	})
}

func (h *LikeHandler) GetCounts(c *gin.Context) {
	feedEntryID, err := strconv.Atoi(c.Param("feed_entry_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feed entry ID"})
		return
	}

	likes, dislikes, err := h.likeRepo.GetFeedEntryCounts(c.Request.Context(), feedEntryID)
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

	feedEntryID, err := strconv.Atoi(c.Param("feed_entry_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feed entry ID"})
		return
	}

	err = h.likeRepo.Remove(c.Request.Context(), userID, feedEntryID)
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

	feedEntryID, err := strconv.Atoi(c.Param("feed_entry_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feed entry ID"})
		return
	}

	status, err := h.likeRepo.GetUserStatus(c.Request.Context(), userID, feedEntryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get like status"})
		return
	}

	if status == nil {
		c.JSON(http.StatusOK, gin.H{"value": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"value": *status})
}
