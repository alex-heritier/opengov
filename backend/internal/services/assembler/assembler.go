package assembler

import (
	"github.com/gin-gonic/gin"

	"github.com/alex/opengov-go/internal/middleware"
	"github.com/alex/opengov-go/internal/models"
	"github.com/alex/opengov-go/internal/repository"
)

type ArticleAssembler struct {
	bookmarkRepo *repository.BookmarkRepository
	likeRepo     *repository.LikeRepository
}

func NewArticleAssembler(bookmarkRepo *repository.BookmarkRepository, likeRepo *repository.LikeRepository) *ArticleAssembler {
	return &ArticleAssembler{
		bookmarkRepo: bookmarkRepo,
		likeRepo:     likeRepo,
	}
}

type ArticleResponse struct {
	ID             int    `json:"id"`
	Title          string `json:"title"`
	Summary        string `json:"summary"`
	SourceURL      string `json:"source_url"`
	DocumentNumber string `json:"document_number"`
	PublishedAt    string `json:"published_at"`
	IsBookmarked   bool   `json:"is_bookmarked,omitempty"`
	UserLikeStatus *bool  `json:"user_like_status,omitempty"`
	LikesCount     int    `json:"likes_count"`
	DislikesCount  int    `json:"dislikes_count"`
}

func (a *ArticleAssembler) EnrichArticle(c *gin.Context, article models.FRArticle) ArticleResponse {
	resp := ArticleResponse{
		ID:             article.ID,
		Title:          article.Title,
		Summary:        article.Summary,
		SourceURL:      article.SourceURL,
		DocumentNumber: article.DocumentNumber,
		PublishedAt:    article.PublishedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	userID, hasAuth := middleware.GetUserID(c)
	if hasAuth {
		isBookmarked, _ := a.bookmarkRepo.IsBookmarked(c.Request.Context(), userID, article.ID)
		resp.IsBookmarked = isBookmarked

		status, _ := a.likeRepo.GetUserStatus(c.Request.Context(), userID, article.ID)
		if status != nil {
			liked := *status == 1
			resp.UserLikeStatus = &liked
		}
	}

	likes, dislikes, _ := a.likeRepo.GetArticleCounts(c.Request.Context(), article.ID)
	resp.LikesCount = likes
	resp.DislikesCount = dislikes

	return resp
}

func (a *ArticleAssembler) EnrichArticles(c *gin.Context, articles []models.FRArticle) []ArticleResponse {
	responses := make([]ArticleResponse, len(articles))
	for i, article := range articles {
		responses[i] = a.EnrichArticle(c, article)
	}
	return responses
}
