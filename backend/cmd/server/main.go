package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"

	"github.com/alex/opengov-go/internal/config"
	"github.com/alex/opengov-go/internal/db"
	"github.com/alex/opengov-go/internal/handlers"
	"github.com/alex/opengov-go/internal/middleware"
	"github.com/alex/opengov-go/internal/repository"
	"github.com/alex/opengov-go/internal/services"
	"github.com/alex/opengov-go/internal/services/assembler"
)

const migrationSQL = `
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT NOT NULL UNIQUE,
    hashed_password TEXT NOT NULL,
    is_active INTEGER NOT NULL DEFAULT 1,
    is_superuser INTEGER NOT NULL DEFAULT 0,
    is_verified INTEGER NOT NULL DEFAULT 0,
    google_id TEXT UNIQUE,
    name TEXT,
    picture_url TEXT,
    political_leaning TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    last_login_at TEXT
);

CREATE TABLE IF NOT EXISTS agencies (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    fr_agency_id INTEGER NOT NULL UNIQUE,
    name TEXT NOT NULL,
    short_name TEXT,
    slug TEXT NOT NULL,
    description TEXT,
    url TEXT,
    json_url TEXT,
    parent_id INTEGER,
    raw_data TEXT NOT NULL DEFAULT '{}',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS frarticles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    document_number TEXT NOT NULL UNIQUE,
    raw_data TEXT NOT NULL,
    fetched_at TEXT NOT NULL DEFAULT (datetime('now')),
    title TEXT NOT NULL,
    summary TEXT NOT NULL,
    source_url TEXT NOT NULL UNIQUE,
    published_at TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS bookmarks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    frarticle_id INTEGER NOT NULL REFERENCES frarticles(id) ON DELETE CASCADE,
    is_bookmarked INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE(user_id, frarticle_id)
);

CREATE TABLE IF NOT EXISTS likes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    frarticle_id INTEGER NOT NULL REFERENCES frarticles(id) ON DELETE CASCADE,
    is_liked INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE(user_id, frarticle_id)
);
`

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

func runMigrations(database *db.DB) error {
	statements := strings.Split(migrationSQL, ";")
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := database.Exec(stmt); err != nil {
			return fmt.Errorf("failed to run migration: %w", err)
		}
	}
	log.Println("Database migrations completed")
	return nil
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

	if err := runMigrations(database); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	articleRepo := repository.NewArticleRepository(database)
	userRepo := repository.NewUserRepository(database)
	agencyRepo := repository.NewAgencyRepository(database)
	bookmarkRepo := repository.NewBookmarkRepository(database)
	likeRepo := repository.NewLikeRepository(database)

	articleAssembler := assembler.NewArticleAssembler(bookmarkRepo, likeRepo)

	authService := services.NewAuthService(cfg, userRepo)
	grokService := services.NewGrokService(cfg)
	frService := services.NewFederalRegisterService(cfg)
	scraperService := services.NewScraperService(cfg, frService, grokService, articleRepo, agencyRepo)

	feedHandler := handlers.NewFeedHandler(articleRepo, articleAssembler)
	bookmarkHandler := handlers.NewBookmarkHandler(bookmarkRepo, articleRepo, likeRepo)
	likeHandler := handlers.NewLikeHandler(likeRepo, articleRepo)
	authHandler := handlers.NewAuthHandler(authService, userRepo)
	adminHandler := handlers.NewAdminHandler(articleRepo, agencyRepo, scraperService)
	oauthHandler := handlers.NewOAuthHandler(authService, userRepo, cfg)

	log.Println("Starting OpenGov API")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Println("Checking database schema...")
	var tableCount int
	database.QueryRowContext(ctx, "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='users'").Scan(&tableCount)
	if tableCount == 0 {
		log.Fatal("Database schema is not up to date! Missing 'users' table. Please run migrations.")
	}
	log.Println("Database schema check passed")

	log.Println("Syncing agencies from Federal Register API...")
	agencyCount, err := scraperService.SyncAgencies(ctx)
	if err != nil {
		log.Printf("Error syncing agencies during startup: %v", err)
	} else if agencyCount > 0 {
		log.Printf("Agency sync completed: %d agencies synced", agencyCount)
	} else {
		log.Println("No agencies returned from Federal Register API during startup")
	}

	log.Println("Scheduling initial scraper in 5 minutes...")
	go func() {
		timer := time.NewTimer(5 * time.Minute)
		defer timer.Stop()
		select {
		case <-timer.C:
			runCtx, runCancel := context.WithTimeout(ctx, 10*time.Minute)
			defer runCancel()
			scraperService.Run(runCtx)
		case <-ctx.Done():
			log.Println("Initial scraper cancelled: server shutting down")
		}
	}()

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
			admin.POST("/scrape", adminHandler.TriggerScrape)
			admin.GET("/stats", adminHandler.GetStats)
			admin.POST("/sync-agencies", adminHandler.SyncAgencies)
			admin.GET("/agencies", adminHandler.GetAgencies)
		}
	}

	c := cron.New()
	c.AddFunc(fmt.Sprintf("@every %ds", int(cfg.ScraperInterval().Seconds())), func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		scraperService.Run(ctx)
	})
	c.Start()

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down...")
		cancel()
		c.Stop()
		os.Exit(0)
	}()

	addr := ":8000"
	log.Printf("Starting server on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
