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
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/alex/opengov-go/internal/config"
	"github.com/alex/opengov-go/internal/db"
	"github.com/alex/opengov-go/internal/handlers"
	"github.com/alex/opengov-go/internal/repository"
	"github.com/alex/opengov-go/internal/services"
)

func corsMiddleware(cfg *config.Config) gin.HandlerFunc {
	corsConfig := cors.DefaultConfig()
	if cfg.CORSEnabled {
		corsConfig.AllowOrigins = cfg.AllowedOrigins
	} else {
		corsConfig.AllowAllOrigins = true
	}
	corsConfig.AllowCredentials = true
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	return cors.New(corsConfig)
}

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

	feedRepo := repository.NewFeedRepository(database)
	docRepo := repository.NewFederalRegisterDocumentRepository(database)
	userRepo := repository.NewUserRepository(database)
	agencyRepo := repository.NewAgencyRepository(database)
	bookmarkRepo := repository.NewBookmarkRepository(database)
	likeRepo := repository.NewLikeRepository(database)

	feedService := services.NewFeedService(feedRepo)
	docService := services.NewFederalRegisterDocumentService(docRepo, feedRepo, database)

	authService := services.NewAuthService(cfg, userRepo)

	feedHandler := handlers.NewFeedHandler(feedService)
	bookmarkHandler := handlers.NewBookmarkHandler(bookmarkRepo, feedService)
	likeHandler := handlers.NewLikeHandler(likeRepo)
	authHandler := handlers.NewAuthHandler(authService, userRepo)
	frService := services.NewFederalRegisterService(cfg)
	summarizer := services.NewSummarizer(cfg)
	scraperService := services.NewScraperService(cfg, frService, summarizer, docService, agencyRepo)
	adminHandler := handlers.NewAdminHandler(docRepo, agencyRepo, scraperService)
	oauthHandler := handlers.NewOAuthHandler(authService, userRepo, cfg)

	deps := RouteDeps{
		DB:              database,
		AuthService:     authService,
		FeedHandler:     feedHandler,
		BookmarkHandler: bookmarkHandler,
		LikeHandler:     likeHandler,
		AuthHandler:     authHandler,
		AdminHandler:    adminHandler,
		OAuthHandler:    oauthHandler,
	}

	log.Println("Starting OpenGov API")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Println("Running database migrations...")
	if err := database.RunMigrations(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	log.Println("Checking database schema...")
	var tableCount int
	database.QueryRowContext(ctx, "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'users'").Scan(&tableCount)
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

	router.Use(corsMiddleware(cfg))

	router.Use(func(c *gin.Context) {
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate, private")
		c.Next()
	})

	router.Use(requestSizeLimitMiddleware(cfg))

	setupRoutes(router, cfg, deps)

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down API server...")
		cancel()
	}()

	addr := ":" + cfg.Port
	log.Printf("Starting API server on %s", addr)

	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}

	log.Println("API server stopped")
}
