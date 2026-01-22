package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/alex/opengov-go/internal/repository"
)

type AdminAPIHandler struct {
	articleRepo *repository.ArticleRepository
	agencyRepo  *repository.AgencyRepository
}

func NewAdminAPIHandler(articleRepo *repository.ArticleRepository, agencyRepo *repository.AgencyRepository) *AdminAPIHandler {
	return &AdminAPIHandler{
		articleRepo: articleRepo,
		agencyRepo:  agencyRepo,
	}
}

func (h *AdminAPIHandler) GetStats(c *gin.Context) {
	total, err := h.articleRepo.Count(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get stats"})
		return
	}

	lastArticle, _ := h.articleRepo.GetLatest(c.Request.Context())

	resp := StatsResponse{
		TotalArticles: total,
	}

	if lastArticle != nil {
		resp.LastScrapeTime = &lastArticle.FetchedAt
		age := time.Since(lastArticle.FetchedAt)
		resp.LastScrapeAge = fmt.Sprintf("%d seconds ago", int(age.Seconds()))
	}

	c.JSON(http.StatusOK, resp)
}

func (h *AdminAPIHandler) GetAgencies(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 500 {
		limit = 500
	}

	agencies, total, err := h.agencyRepo.GetAll(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get agencies"})
		return
	}

	var results []gin.H
	for _, a := range agencies {
		results = append(results, gin.H{
			"id":           a.ID,
			"fr_agency_id": a.FRAgencyID,
			"name":         a.Name,
			"short_name":   a.ShortName,
			"slug":         a.Slug,
			"description":  a.Description,
			"url":          a.URL,
			"parent_id":    a.ParentID,
			"created_at":   a.CreatedAt,
			"updated_at":   a.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"agencies": results,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
	})
}
