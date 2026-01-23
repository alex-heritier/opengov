package main

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/alex/opengov-go/internal/config"
	"github.com/alex/opengov-go/internal/db"
	"github.com/alex/opengov-go/internal/handlers"
	"github.com/alex/opengov-go/internal/middleware"
	"github.com/alex/opengov-go/internal/services"
)

type RouteDeps struct {
	DB              *db.DB
	AuthService     *services.AuthService
	FeedHandler     *handlers.FeedHandler
	BookmarkHandler *handlers.BookmarkHandler
	LikeHandler     *handlers.LikeHandler
	AuthHandler     *handlers.AuthHandler
	AdminHandler    *handlers.AdminAPIHandler
	OAuthHandler    *handlers.OAuthHandler
}

func setupRoutes(router *gin.Engine, _ *config.Config, deps RouteDeps) {
	router.GET("/health", func(c *gin.Context) {
		if err := deps.DB.HealthCheck(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "error", "database": "disconnected"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.GET("/health/db", func(c *gin.Context) {
		if err := deps.DB.HealthCheck(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":   "error",
				"database": "disconnected",
				"error":    err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":   "ok",
			"database": "connected",
		})
	})

	api := router.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/login", deps.AuthHandler.Login)
			auth.POST("/register", deps.AuthHandler.Register)
			auth.POST("/logout", deps.AuthHandler.Logout)
			auth.GET("/me", middleware.AuthMiddleware(deps.AuthService), deps.AuthHandler.Me)
			auth.POST("/refresh", middleware.AuthMiddleware(deps.AuthService), deps.AuthHandler.Refresh)
		}

		googleAuth := api.Group("/auth/google")
		{
			googleAuth.GET("/login", deps.OAuthHandler.GoogleLogin)
			googleAuth.GET("/callback", deps.OAuthHandler.GoogleCallback)
		}

		testAuth := api.Group("/auth/test")
		{
			testAuth.GET("/login", deps.OAuthHandler.TestLogin)
		}

		feed := api.Group("/feed")
		feed.Use(middleware.OptionalAuthMiddleware(deps.AuthService))
		{
			feed.GET("", deps.FeedHandler.GetFeed)
			feed.GET("/:id", deps.FeedHandler.GetArticle)
			feed.GET("/document/:document_number", deps.FeedHandler.GetArticleByDocumentNumber)
			feed.GET("/slug/:unique_key", deps.FeedHandler.GetArticleByUniqueKey)
		}

		bookmarks := api.Group("/bookmarks")
		bookmarks.Use(middleware.AuthMiddleware(deps.AuthService))
		{
			bookmarks.POST("/:article_id", deps.BookmarkHandler.Toggle)
			bookmarks.GET("", deps.BookmarkHandler.GetBookmarks)
			bookmarks.DELETE("/:article_id", deps.BookmarkHandler.Remove)
			bookmarks.GET("/status/:article_id", deps.BookmarkHandler.GetStatus)
		}

		likes := api.Group("/likes")
		likes.Use(middleware.AuthMiddleware(deps.AuthService))
		{
			likes.POST("/:article_id", deps.LikeHandler.Toggle)
			likes.GET("/counts/:article_id", deps.LikeHandler.GetCounts)
			likes.DELETE("/:article_id", deps.LikeHandler.Remove)
			likes.GET("/status/:article_id", deps.LikeHandler.GetStatus)
		}

		admin := api.Group("/admin")
		{
			admin.GET("/stats", deps.AdminHandler.GetStats)
			admin.GET("/agencies", deps.AdminHandler.GetAgencies)
		}
	}
}
