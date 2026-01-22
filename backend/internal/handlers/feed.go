package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/alex/opengov-go/internal/repository"
	"github.com/alex/opengov-go/internal/services/assembler"
)

type FeedHandler struct {
	articleRepo *repository.ArticleRepository
	assembler   *assembler.ArticleAssembler
}

func NewFeedHandler(articleRepo *repository.ArticleRepository, assembler *assembler.ArticleAssembler) *FeedHandler {
	return &FeedHandler{
		articleRepo: articleRepo,
		assembler:   assembler,
	}
}

type ArticleResponse = assembler.ArticleResponse

type FeedResponse struct {
	Articles []ArticleResponse `json:"articles"`
	Page     int               `json:"page"`
	Limit    int               `json:"limit"`
	Total    int               `json:"total"`
	HasNext  bool              `json:"has_next"`
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

	articles, total, err := h.articleRepo.GetFeed(c.Request.Context(), page, limit, sort)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch feed"})
		return
	}

	c.JSON(http.StatusOK, FeedResponse{
		Articles: h.assembler.EnrichArticles(c, articles),
		Page:     page,
		Limit:    limit,
		Total:    total,
		HasNext:  offset+limit < total,
	})
}

func (h *FeedHandler) GetArticle(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
		return
	}

	article, err := h.articleRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch article"})
		return
	}
	if article == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
		return
	}

	c.JSON(http.StatusOK, h.assembler.EnrichArticle(c, *article))
}

func (h *FeedHandler) GetArticleByDocumentNumber(c *gin.Context) {
	docNumber := c.Param("document_number")

	article, err := h.articleRepo.GetByDocumentNumber(c.Request.Context(), docNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch article"})
		return
	}
	if article == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
		return
	}

	c.JSON(http.StatusOK, h.assembler.EnrichArticle(c, *article))
}
