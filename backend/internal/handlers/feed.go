package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/alex/opengov-go/internal/middleware"
	"github.com/alex/opengov-go/internal/services"
)

type FeedHandler struct {
	feedService *services.FeedService
}

func NewFeedHandler(feedService *services.FeedService) *FeedHandler {
	return &FeedHandler{
		feedService: feedService,
	}
}

func (h *FeedHandler) GetFeed(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	sort := c.DefaultQuery("sort", "newest")

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit
	if offset > 10000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Page number too high"})
		return
	}

	userID, hasAuth := middleware.GetUserID(c)
	var resp services.FeedResponse
	var err error

	if hasAuth {
		resp, err = h.feedService.GetFeed(c.Request.Context(), &userID, page, limit, sort)
	} else {
		resp, err = h.feedService.GetFeed(c.Request.Context(), nil, page, limit, sort)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch feed"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *FeedHandler) GetItem(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feed entry ID"})
		return
	}

	userID, hasAuth := middleware.GetUserID(c)
	var item *services.FeedEntryResponse
	var svcErr error

	if hasAuth {
		item, svcErr = h.feedService.GetItem(c.Request.Context(), &userID, id)
	} else {
		item, svcErr = h.feedService.GetItem(c.Request.Context(), nil, id)
	}

	if svcErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch feed entry"})
		return
	}
	if item == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Feed entry not found"})
		return
	}

	c.JSON(http.StatusOK, item)
}
