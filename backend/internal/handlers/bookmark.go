package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/alex/opengov-go/internal/middleware"
	"github.com/alex/opengov-go/internal/repository"
	"github.com/alex/opengov-go/internal/services"
)

type BookmarkHandler struct {
	bookmarkRepo *repository.BookmarkRepository
	feedService  *services.FeedService
}

func NewBookmarkHandler(bookmarkRepo *repository.BookmarkRepository, feedService *services.FeedService) *BookmarkHandler {
	return &BookmarkHandler{
		bookmarkRepo: bookmarkRepo,
		feedService:  feedService,
	}
}

func (h *BookmarkHandler) Toggle(c *gin.Context) {
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

	isBookmarked, err := h.bookmarkRepo.Toggle(c.Request.Context(), userID, feedEntryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to toggle bookmark"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"is_bookmarked": isBookmarked,
	})
}

func (h *BookmarkHandler) GetBookmarks(c *gin.Context) {
	userID, hasAuth := middleware.GetUserID(c)
	if !hasAuth {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	items, err := h.feedService.GetBookmarkedFeed(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch bookmarks"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": items,
		"total": len(items),
	})
}

func (h *BookmarkHandler) Remove(c *gin.Context) {
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

	err = h.bookmarkRepo.Remove(c.Request.Context(), userID, feedEntryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove bookmark"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Bookmark removed",
	})
}

func (h *BookmarkHandler) GetStatus(c *gin.Context) {
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

	isBookmarked, err := h.bookmarkRepo.IsBookmarked(c.Request.Context(), userID, feedEntryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get bookmark status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"is_bookmarked": isBookmarked})
}
