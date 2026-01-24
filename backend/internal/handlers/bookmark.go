package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/alex/opengov-go/internal/middleware"
	"github.com/alex/opengov-go/internal/repository"
	"github.com/alex/opengov-go/internal/services/assembler"
)

type BookmarkHandler struct {
	bookmarkRepo *repository.BookmarkRepository
	articleRepo  *repository.ArticleRepository
	assembler    *assembler.ArticleAssembler
}

func NewBookmarkHandler(bookmarkRepo *repository.BookmarkRepository, articleRepo *repository.ArticleRepository, assembler *assembler.ArticleAssembler) *BookmarkHandler {
	return &BookmarkHandler{
		bookmarkRepo: bookmarkRepo,
		articleRepo:  articleRepo,
		assembler:    assembler,
	}
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
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
		return
	}
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

	articleResponses := h.assembler.EnrichArticles(c, articles)
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
