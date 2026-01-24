package assembler

/*
	ASSEMBLER PERFORMANCE NOTES:

	This assembler currently suffers from N+1 query problems on the feed page.

	For each article, the following queries are executed:
	- bookmarkRepo.IsBookmarked(userID, articleID)
	- likeRepo.GetUserStatus(userID, articleID)
	- likeRepo.GetArticleCounts(articleID) // two count queries

	A 20-article page results in 40-60+ queries.

	TODO: Refactor to use bulk repository methods:
	- GetBookmarksForArticles(userID, articleIDs[]) -> map[articleID]bool
	- GetUserStatusesForArticles(userID, articleIDs[]) -> map[articleID]int
	- GetCountsForArticles(articleIDs[]) -> map[articleID]struct{likes, dislikes int}

	This will reduce a 20-article page from 60+ queries to 3-4 queries.
*/

import (
	"github.com/gin-gonic/gin"

	"github.com/alex/opengov-go/internal/middleware"
	"github.com/alex/opengov-go/internal/models"
	"github.com/alex/opengov-go/internal/repository"
	"github.com/alex/opengov-go/internal/timeformat"
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
	UniqueKey      string `json:"unique_key"`
	PublishedAt    string `json:"published_at"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
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
		UniqueKey:      article.UniqueKey,
		PublishedAt:    article.PublishedAt.Format(timeformat.DBTime),
		CreatedAt:      article.CreatedAt.Format(timeformat.DBTime),
		UpdatedAt:      article.UpdatedAt.Format(timeformat.DBTime),
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
