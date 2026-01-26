package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/alex/opengov-go/internal/client"
	"github.com/alex/opengov-go/internal/config"
	"github.com/alex/opengov-go/internal/db"
	"github.com/alex/opengov-go/internal/repository"
	"github.com/alex/opengov-go/internal/services"
)

func runMigrations(database *db.DB) error {
	return database.RunMigrations()
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

	log.Println("Checking database schema...")
	var tableCount int
	database.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'users'").Scan(&tableCount)
	if tableCount == 0 {
		log.Fatal("Database schema is not up to date! Missing 'users' table. Please run migrations.")
	}
	log.Println("Database schema check passed")

	docRepo := repository.NewPolicyDocumentRepository(database)
	feedRepo := repository.NewFeedRepository(database)
	agencyRepo := repository.NewAgencyRepository(database)
	rawEntryRepo := repository.NewRawEntryRepository(database)

	frClient := client.NewFederalRegisterClient(cfg)
	summarizer := services.NewSummarizer(cfg)
	docService := services.NewPolicyDocumentService(docRepo, feedRepo, rawEntryRepo, database)
	scraperService := services.NewScraperService(cfg, frClient, summarizer, docService, agencyRepo)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down scraping service...")
		cancel()
		os.Exit(0)
	}()

	syncAgenciesNow := false
	for _, arg := range os.Args[1:] {
		switch arg {
		case "--sync-agencies":
			syncAgenciesNow = true
		case "--help", "-h":
			fmt.Println("Usage: scraping [options]")
			fmt.Println("Options:")
			fmt.Println("  --sync-agencies Run agency sync once and exit")
			fmt.Println("  --help, -h      Show this help message")
			os.Exit(0)
		}
	}

	if syncAgenciesNow {
		count, err := scraperService.SyncAgencies(ctx)
		if err != nil {
			log.Printf("Error during agency sync: %v", err)
		} else {
			log.Printf("Agency sync completed: %d agencies synced", count)
		}
		return
	}

	log.Println("Starting OpenGov Scraper")

	log.Println("Syncing agencies on startup...")
	agencyCount, err := scraperService.SyncAgencies(ctx)
	if err != nil {
		log.Printf("Error syncing agencies during startup: %v", err)
	} else if agencyCount > 0 {
		log.Printf("Agency sync completed: %d agencies synced", agencyCount)
	} else {
		log.Println("No agencies returned from Federal Register API during startup")
	}

	log.Println("Running scrape...")
	scraperService.Run(ctx)
	log.Println("Scraper completed")

	os.Exit(0)
}
