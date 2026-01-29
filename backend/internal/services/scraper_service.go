package services

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/alex/opengov-go/internal/client"
	"github.com/alex/opengov-go/internal/config"
	"github.com/alex/opengov-go/internal/constants"
	"github.com/alex/opengov-go/internal/domain"
	"github.com/alex/opengov-go/internal/repository"
	"github.com/alex/opengov-go/internal/scrape"
)

type ScraperService struct {
	summarizer          Summarizer
	docSvc              *PolicyDocumentService
	agencyRepo          *repository.AgencyRepository
	scraperDaysLookback int
	fedregScraper       *scrape.FedregScraper

	// Use a slice to support multiple scrapers in the future
	retrievers []scrape.PolicyDocumentScraper
}

func NewScraperService(cfg *config.Config, frClient *client.FederalRegisterClient, summarizer Summarizer, docSvc *PolicyDocumentService, agencyRepo *repository.AgencyRepository) *ScraperService {
	svc := ScraperService{
		summarizer:          summarizer,
		docSvc:              docSvc,
		agencyRepo:          agencyRepo,
		scraperDaysLookback: cfg.ScraperDaysLookback,
	}

	svc.fedregScraper = scrape.NewFedregScraper(frClient)
	svc.retrievers = []scrape.PolicyDocumentScraper{
		svc.fedregScraper,
	}

	return &svc
}

func (s *ScraperService) Run(ctx context.Context) {
	log.Println("Starting scrape...")

	totalProcessed := 0
	totalSkipped := 0
	totalErrors := 0

	for _, retriever := range s.retrievers {
		if ctx.Err() != nil {
			log.Println("Scraper cancelled before retriever start, stopping...")
			return
		}
		results, err := retriever.Scrape(ctx, s.scraperDaysLookback)
		if err != nil {
			if ctx.Err() != nil {
				log.Println("Scraper cancelled during retriever scrape, stopping...")
				return
			}
			log.Printf("Failed to fetch documents from retriever: %v", err)
			continue
		}

		log.Printf("Fetched %d documents", len(results))

		processed, skipped, errors := s.processScrapeResults(ctx, results)
		totalProcessed += processed
		totalSkipped += skipped
		totalErrors += errors
	}

	log.Printf("Scraper completed. Total Processed: %d, Skipped: %d, Errors: %d", totalProcessed, totalSkipped, totalErrors)
}

func (s *ScraperService) processScrapeResults(ctx context.Context, results []scrape.ScrapeResult) (int, int, int) {
	processedCount := 0
	skippedCount := 0
	errorCount := 0

	for _, result := range results {
		select {
		case <-ctx.Done():
			log.Println("Scraper cancelled mid-run, stopping...")
			return processedCount, skippedCount, errorCount
		default:
		}

		status := s.processSingleScrapeResult(ctx, result)
		switch status {
		case "processed":
			processedCount++
		case "skipped":
			skippedCount++
		case "error":
			errorCount++
		}
	}

	log.Printf("Retriever completed. Processed: %d, Skipped: %d, Errors: %d", processedCount, skippedCount, errorCount)
	return processedCount, skippedCount, errorCount
}

func (s *ScraperService) processSingleScrapeResult(ctx context.Context, result scrape.ScrapeResult) string {
	doc := result.PolicyDocument

	abstract := ""
	if doc.Abstract != nil {
		abstract = *doc.Abstract
	}
	if doc.Excerpts != nil {
		abstract = *doc.Excerpts
	}
	if len(abstract) > 1000 {
		abstract = abstract[:1000]
	}

	var agency string
	if len(doc.Agencies) > 0 {
		agency = doc.Agencies[0].Name
	}

	analysis, err := s.summarizer.Analyze(ctx, doc.Title, abstract, agency)
	if err != nil {
		log.Printf("Failed to analyze %s: %v", doc.DocumentNumber, err)
		analysis = &AIAnalysis{
			Summary:        abstract,
			Keypoints:      []string{},
			ImpactScore:    "medium",
			PoliticalScore: 0,
		}
	}

	pubDate, _ := time.Parse("2006-01-02", doc.PublicationDate)

	var agencyPtr *string
	if agency != "" {
		agencyPtr = &agency
	}

	newDoc := &domain.PolicyDocument{
		SourceKey:      constants.SourceTypeFederalRegister,
		ExternalID:     doc.DocumentNumber,
		Title:          doc.Title,
		Agency:         agencyPtr,
		Summary:        analysis.Summary,
		Keypoints:      analysis.Keypoints,
		ImpactScore:    &analysis.ImpactScore,
		PoliticalScore: &analysis.PoliticalScore,
		SourceURL:      doc.HTMLURL,
		PublishedAt:    pubDate,
		DocumentType:   &doc.Type,
		PDFURL:         doc.PDFURL,
	}

	rawPayload := result.RawResult
	fetchedAt := time.Now().UTC()

	resultDoc, err := s.docSvc.CreateFromScrape(ctx, newDoc, rawPayload, fetchedAt)
	if err != nil {
		log.Printf("Failed to create document %s: %v", doc.DocumentNumber, err)
		return "error"
	}

	if time.Since(resultDoc.CreatedAt) < 5*time.Second {
		log.Printf("Created document: %s - %s", doc.DocumentNumber, truncate(newDoc.Title, 60))
		return "processed"
	}

	log.Printf("Skipped duplicate: %s", doc.DocumentNumber)
	return "skipped"
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

func (s *ScraperService) SyncAgencies(ctx context.Context) (int, error) {
	log.Println("Syncing agencies...")

	frAgencies, err := s.fedregScraper.Client().FetchAgencies(ctx)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, frAgency := range frAgencies {
		rawData, _ := json.Marshal(frAgency)

		now := time.Now().UTC()
		agency := &domain.Agency{
			FRAgencyID:  frAgency.ID,
			RawName:     frAgency.RawName,
			Name:        frAgency.Name,
			ShortName:   &frAgency.ShortName,
			Slug:        frAgency.Slug,
			Description: frAgency.Description,
			URL:         &frAgency.URL,
			JSONURL:     &frAgency.JSONURL,
			ParentID:    frAgency.ParentID,
			RawData:     domain.JSONMap{},
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		json.Unmarshal(rawData, &agency.RawData)

		if err := s.agencyRepo.Upsert(ctx, agency); err != nil {
			log.Printf("Failed to upsert agency %s: %v", frAgency.Name, err)
			continue
		}
		count++
	}

	log.Printf("Synced %d agencies", count)
	return count, nil
}
