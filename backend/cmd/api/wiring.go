package main

import (
	"github.com/alex/opengov-go/internal/client"
	"github.com/alex/opengov-go/internal/config"
	"github.com/alex/opengov-go/internal/db"
	"github.com/alex/opengov-go/internal/handlers"
	"github.com/alex/opengov-go/internal/repository"
	"github.com/alex/opengov-go/internal/services"
)

func wireDependencies(cfg *config.Config, database *db.DB) (RouteDeps, error) {
	feedRepo := repository.NewFeedRepository(database)
	docRepo := repository.NewPolicyDocumentRepository(database)
	userRepo := repository.NewUserRepository(database)
	agencyRepo := repository.NewAgencyRepository(database)
	bookmarkRepo := repository.NewBookmarkRepository(database)
	likeRepo := repository.NewLikeRepository(database)

	feedService := services.NewFeedService(feedRepo)
	authService := services.NewAuthService(cfg, userRepo)

	feedHandler := handlers.NewFeedHandler(feedService)
	bookmarkHandler := handlers.NewBookmarkHandler(bookmarkRepo, feedService)
	likeHandler := handlers.NewLikeHandler(likeRepo)
	authHandler := handlers.NewAuthHandler(authService, userRepo)

	frClient := client.NewFederalRegisterClient(cfg)
	agencySync := services.NewAgencySyncService(frClient, agencyRepo)

	adminHandler := handlers.NewAdminHandler(docRepo, agencyRepo, agencySync)
	oauthHandler := handlers.NewOAuthHandler(authService, userRepo, cfg)

	return RouteDeps{
		DB:              database,
		AuthService:     authService,
		FeedHandler:     feedHandler,
		BookmarkHandler: bookmarkHandler,
		LikeHandler:     likeHandler,
		AuthHandler:     authHandler,
		AdminHandler:    adminHandler,
		OAuthHandler:    oauthHandler,
	}, nil
}
