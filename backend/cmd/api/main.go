package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/alex/opengov-go/internal/config"
	"github.com/alex/opengov-go/internal/db"
	"github.com/alex/opengov-go/internal/handlers"
	"github.com/alex/opengov-go/internal/middleware"
	"github.com/alex/opengov-go/internal/repository"
	"github.com/alex/opengov-go/internal/services"
	"github.com/alex/opengov-go/internal/services/assembler"
)

func requestSizeLimitMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		contentLength := c.GetHeader("Content-Length")
		if contentLength != "" {
			length, err := strconv.ParseInt(contentLength, 10, 64)
			if err == nil && length > int64(cfg.MaxRequestSizeBytes) {
				maxMB := cfg.MaxRequestSizeBytes / (1024 * 1024)
				c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{
					"detail": fmt.Sprintf("Request body too large (max %d MB)", maxMB),
				})
				return
			}
		}
		c.Next()
	}
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	database, err := db.New(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	articleRepo := repository.NewArticleRepository(database)
	userRepo := repository.NewUserRepository(database)
	agencyRepo := repository.NewAgencyRepository(database)
	bookmarkRepo := repository.NewBookmarkRepository(database)
	likeRepo := repository.NewLikeRepository(database)

	articleAssembler := assembler.NewArticleAssembler(bookmarkRepo, likeRepo)

	authService := services.NewAuthService(cfg, userRepo)

	feedHandler := handlers.NewFeedHandler(articleRepo, articleAssembler)
	bookmarkHandler := handlers.NewBookmarkHandler(bookmarkRepo, articleRepo, likeRepo)
	likeHandler := handlers.NewLikeHandler(likeRepo, articleRepo)
	authHandler := handlers.NewAuthHandler(authService, userRepo)
	adminHandler := handlers.NewAdminAPIHandler(articleRepo, agencyRepo)
	oauthHandler := handlers.NewOAuthHandler(authService, userRepo, cfg)

	log.Println("Starting OpenGov API")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Println("Running database migrations...")
	if err := database.RunMigrations(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	log.Println("Checking database schema...")
	var tableCount int
	database.QueryRowContext(ctx, "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='users'").Scan(&tableCount)
	if tableCount == 0 {
		log.Fatal("Database schema is not up to date! Missing 'users' table. Please run migrations.")
	}
	log.Println("Database schema check passed")

	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = cfg.AllowedOrigins
	corsConfig.AllowCredentials = true
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	router.Use(cors.New(corsConfig))

	router.Use(func(c *gin.Context) {
		c.Header("Cache-Control", "public, max-age=300")
		c.Next()
	})

	router.Use(requestSizeLimitMiddleware(cfg))

	router.GET("/health", func(c *gin.Context) {
		if err := database.HealthCheck(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "error", "database": "disconnected"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.GET("/health/db", func(c *gin.Context) {
		if err := database.HealthCheck(); err != nil {
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
			auth.POST("/login", authHandler.Login)
			auth.POST("/register", authHandler.Register)
			auth.POST("/logout", authHandler.Logout)
			auth.GET("/me", middleware.AuthMiddleware(authService), authHandler.Me)
			auth.POST("/refresh", middleware.AuthMiddleware(authService), authHandler.Refresh)
		}

		googleAuth := api.Group("/auth/google")
		{
			googleAuth.GET("/login", oauthHandler.GoogleLogin)
			googleAuth.GET("/callback", oauthHandler.GoogleCallback)
		}

		feed := api.Group("/feed")
		{
			feed.GET("", feedHandler.GetFeed)
			feed.GET("/:id", feedHandler.GetArticle)
			feed.GET("/document/:document_number", feedHandler.GetArticleByDocumentNumber)
			feed.GET("/slug/:unique_key", feedHandler.GetArticleByUniqueKey)
		}

		bookmarks := api.Group("/bookmarks")
		bookmarks.Use(middleware.AuthMiddleware(authService))
		{
			bookmarks.POST("/:article_id", bookmarkHandler.Toggle)
			bookmarks.GET("", bookmarkHandler.GetBookmarks)
			bookmarks.DELETE("/:article_id", bookmarkHandler.Remove)
			bookmarks.GET("/status/:article_id", bookmarkHandler.GetStatus)
		}

		likes := api.Group("/likes")
		likes.Use(middleware.AuthMiddleware(authService))
		{
			likes.POST("/:article_id", likeHandler.Toggle)
			likes.GET("/counts/:article_id", likeHandler.GetCounts)
			likes.DELETE("/:article_id", likeHandler.Remove)
			likes.GET("/status/:article_id", likeHandler.GetStatus)
		}

		admin := api.Group("/admin")
		{
			admin.GET("/stats", adminHandler.GetStats)
			admin.GET("/agencies", adminHandler.GetAgencies)
		}
	}

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down API server...")
		cancel()
		os.Exit(0)
	}()

	addr := ":8000"
	log.Printf("Starting API server on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
