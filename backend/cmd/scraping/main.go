package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/alex/opengov-go/internal/config"
	"github.com/alex/opengov-go/internal/db"
	"github.com/alex/opengov-go/internal/repository"
	"github.com/alex/opengov-go/internal/services"
)

func runMigrations(database *db.DB) error {
	return database.RunMigrations()
}

func runOnce(scraperService *services.ScraperService, ctx context.Context) {
	log.Println("Running immediate scrape...")
	scraperService.Run(ctx)
	log.Println("Immediate scrape completed")
}

func syncAgenciesOnce(scraperService *services.ScraperService, ctx context.Context) {
	log.Println("Running immediate agency sync...")
	count, err := scraperService.SyncAgencies(ctx)
	if err != nil {
		log.Printf("Error during agency sync: %v", err)
	} else {
		log.Printf("Agency sync completed: %d agencies synced", count)
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

	if err := runMigrations(database); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	log.Println("Checking database schema...")
	var tableCount int
	database.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='users'").Scan(&tableCount)
	if tableCount == 0 {
		log.Fatal("Database schema is not up to date! Missing 'users' table. Please run migrations.")
	}
	log.Println("Database schema check passed")

	articleRepo := repository.NewArticleRepository(database)
	agencyRepo := repository.NewAgencyRepository(database)

	frService := services.NewFederalRegisterService(cfg)
	summarizer := services.NewSummarizer(cfg)
	scraperService := services.NewScraperService(cfg, frService, summarizer, articleRepo, agencyRepo)

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

	runOnceNow := false
	syncAgenciesNow := false
	for _, arg := range os.Args[1:] {
		switch arg {
		case "--once":
			runOnceNow = true
		case "--sync-agencies":
			syncAgenciesNow = true
		case "--help", "-h":
			fmt.Println("Usage: scraping [options]")
			fmt.Println("Options:")
			fmt.Println("  --once          Run scrape once and exit")
			fmt.Println("  --sync-agencies Run agency sync once and exit")
			fmt.Println("  --help, -h      Show this help message")
			os.Exit(0)
		}
	}

	if runOnceNow {
		runCtx, runCancel := context.WithTimeout(ctx, 10*time.Minute)
		defer runCancel()
		runOnce(scraperService, runCtx)
		return
	}

	if syncAgenciesNow {
		syncCtx, syncCancel := context.WithTimeout(ctx, 5*time.Minute)
		defer syncCancel()
		syncAgenciesOnce(scraperService, syncCtx)
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

	log.Println("Scheduling initial scraper in 5 minutes...")
	go func() {
		timer := time.NewTimer(5 * time.Minute)
		defer timer.Stop()
		select {
		case <-timer.C:
			runCtx, runCancel := context.WithTimeout(ctx, 10*time.Minute)
			defer runCancel()
			runOnce(scraperService, runCtx)
		case <-ctx.Done():
			log.Println("Initial scraper cancelled: service shutting down")
		}
	}()

	c := cron.New()
	intervalSeconds := int(cfg.ScraperInterval().Seconds())
	c.AddFunc(fmt.Sprintf("@every %ds", intervalSeconds), func() {
		runCtx, runCancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer runCancel()
		scraperService.Run(runCtx)
	})
	c.Start()

	log.Printf("Scraper service started. Interval: %d seconds", intervalSeconds)
	log.Println("Press Ctrl+C to stop")

	select {}
}
