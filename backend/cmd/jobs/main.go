package main

import (
	"context"
	"flag"
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

func main() {
	job := flag.String("job", "", "job to run (migrate|sync-agencies|scrape|canonicalize|enrich|materialize|pipeline)")
	flag.Parse()

	if *job == "" {
		fmt.Fprintln(os.Stderr, "missing required flag: --job")
		flag.Usage()
		os.Exit(2)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	database, err := db.New(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	docRepo := repository.NewPolicyDocumentRepository(database)
	feedRepo := repository.NewFeedRepository(database)
	agencyRepo := repository.NewAgencyRepository(database)
	rawRepo := repository.NewRawPolicyDocumentRepository(database)

	frClient := client.NewFederalRegisterClient(cfg)
	jobs := services.NewJobsService(cfg, database, agencyRepo, rawRepo, docRepo, feedRepo, frClient)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Shutting down jobs...")
		cancel()
	}()

	switch *job {
	case "migrate":
		if err := jobs.Migrate(); err != nil {
			log.Fatalf("Failed to run migrations: %v", err)
		}
		log.Println("Migrations completed successfully")
		return
	case "sync-agencies":
		n, err := jobs.SyncAgencies(ctx)
		if err != nil {
			log.Fatalf("sync-agencies failed: %v", err)
		}
		log.Printf("sync-agencies completed: %d agencies synced", n)
	case "scrape":
		processed, skipped, err := jobs.ScrapeRaw(ctx)
		if err != nil {
			log.Fatalf("scrape failed: %v", err)
		}
		log.Printf("scrape completed: inserted=%d skipped=%d", processed, skipped)
	case "canonicalize":
		linked, err := jobs.Canonicalize(ctx, 200)
		if err != nil {
			log.Fatalf("canonicalize failed: %v", err)
		}
		log.Printf("canonicalize completed: linked=%d", linked)
	case "enrich":
		wouldEnrich, err := jobs.Enrich(ctx, 200)
		if err != nil {
			log.Fatalf("enrich failed: %v", err)
		}
		log.Printf("enrich completed (dry-run): would_enrich=%d", wouldEnrich)
	case "materialize":
		upserted, err := jobs.Materialize(ctx, 500)
		if err != nil {
			log.Fatalf("materialize failed: %v", err)
		}
		log.Printf("materialize completed: upserted=%d", upserted)
	case "pipeline":
		if err := jobs.Pipeline(ctx); err != nil {
			log.Fatalf("pipeline failed: %v", err)
		}
		log.Println("pipeline completed")
	default:
		log.Fatalf("unknown job: %q", *job)
	}
}
