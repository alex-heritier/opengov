package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/alex/opengov-go/internal/middleware"
	"github.com/alex/opengov-go/internal/models"
	"github.com/alex/opengov-go/internal/repository"
)

type BookmarkHandler struct {
	bookmarkRepo *repository.BookmarkRepository
	articleRepo  *repository.ArticleRepository
	likeRepo     *repository.LikeRepository
}

func NewBookmarkHandler(bookmarkRepo *repository.BookmarkRepository, articleRepo *repository.ArticleRepository, likeRepo *repository.LikeRepository) *BookmarkHandler {
	return &BookmarkHandler{
		bookmarkRepo: bookmarkRepo,
		articleRepo:  articleRepo,
		likeRepo:     likeRepo,
	}
}

// enrichBookmarkedArticle creates a fully enriched ArticleResponse for bookmarked articles.
// Since these are by definition bookmarked, we skip the IsBookmarked check and always set it to true.
// UserLikeStatus and counts are still enriched based on auth context.
func (h *BookmarkHandler) enrichBookmarkedArticle(c *gin.Context, article models.FRArticle) ArticleResponse {
	resp := ArticleResponse{
		ID:           article.ID,
		Title:        article.Title,
		Summary:      article.Summary,
		PublishedAt:  article.PublishedAt.Format("2006-01-02T15:04:05Z07:00"),
		IsBookmarked: true, // Always true for bookmarked articles
	}

	userID, hasAuth := middleware.GetUserID(c)
	if hasAuth {
		status, _ := h.likeRepo.GetUserStatus(c.Request.Context(), userID, article.ID)
		if status != nil {
			liked := *status == 1
			resp.UserLikeStatus = &liked
		}
	}

	likes, dislikes, _ := h.likeRepo.GetArticleCounts(c.Request.Context(), article.ID)
	resp.LikesCount = likes
	resp.DislikesCount = dislikes

	return resp
}

func (h *BookmarkHandler) Toggle(c *gin.Context) {
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

	_, err = h.articleRepo.GetByID(c.Request.Context(), articleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check article"})
		return
	}

	bookmark, err := h.bookmarkRepo.Toggle(c.Request.Context(), userID, articleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to toggle bookmark"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"is_bookmarked": bookmark.GetIsBookmarked(),
	})
}

func (h *BookmarkHandler) GetBookmarks(c *gin.Context) {
	userID, hasAuth := middleware.GetUserID(c)
	if !hasAuth {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	articles, err := h.bookmarkRepo.GetBookmarkedArticles(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get bookmarks"})
		return
	}

	// Use the same ArticleResponse enrichment logic for consistency
	articleResponses := make([]ArticleResponse, len(articles))
	for i, article := range articles {
		articleResponses[i] = h.enrichBookmarkedArticle(c, article)
	}

	c.JSON(http.StatusOK, gin.H{"articles": articleResponses})
}

func (h *BookmarkHandler) Remove(c *gin.Context) {
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

	err = h.bookmarkRepo.Remove(c.Request.Context(), userID, articleID)
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

	articleID, err := strconv.Atoi(c.Param("article_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
		return
	}

	isBookmarked, err := h.bookmarkRepo.IsBookmarked(c.Request.Context(), userID, articleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get bookmark status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"is_bookmarked": isBookmarked})
}
