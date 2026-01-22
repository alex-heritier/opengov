package services

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/alex/opengov-go/internal/config"
	"github.com/alex/opengov-go/internal/models"
	"github.com/alex/opengov-go/internal/repository"
)

const batchSize = 50

type ScraperService struct {
	frService   *FederalRegisterService
	grokService *GrokService
	articleRepo *repository.ArticleRepository
	agencyRepo  *repository.AgencyRepository
	scraperDays int
}

func NewScraperService(cfg *config.Config, frService *FederalRegisterService, grokService *GrokService, articleRepo *repository.ArticleRepository, agencyRepo *repository.AgencyRepository) *ScraperService {
	return &ScraperService{
		frService:   frService,
		grokService: grokService,
		articleRepo: articleRepo,
		agencyRepo:  agencyRepo,
		scraperDays: cfg.ScraperDaysLookback,
	}
}

func (s *ScraperService) Run(ctx context.Context) {
	log.Println("Starting scraper...")

	docs, err := s.frService.FetchRecentDocuments(ctx, s.scraperDays)
	if err != nil {
		log.Printf("Failed to fetch documents: %v", err)
		return
	}

	log.Printf("Fetched %d documents", len(docs))

	processedCount := 0
	skippedCount := 0
	errorCount := 0

	var batch []models.FRArticle

	for _, doc := range docs {
		select {
		case <-ctx.Done():
			log.Println("Scraper cancelled mid-run, stopping...")
			return
		default:
		}

		existsByDoc, _ := s.articleRepo.ExistsByDocumentNumber(ctx, doc.DocumentNumber)
		existsByURL, _ := s.articleRepo.ExistsBySourceURL(ctx, doc.HTMLURL)

		if existsByDoc || existsByURL {
			log.Printf("Skipping duplicate: %s", doc.DocumentNumber)
			skippedCount++
			continue
		}

		abstract := doc.Abstract
		if abstract == "" {
			abstract = doc.FullText
		}
		if len(abstract) > 1000 {
			abstract = abstract[:1000]
		}

		summary, err := s.grokService.Summarize(ctx, abstract)
		if err != nil {
			log.Printf("Failed to summarize %s: %v", doc.DocumentNumber, err)
			summary = abstract
		}

		pubDate, _ := time.Parse("2006-01-02", doc.PublicationDate)

		article := &models.FRArticle{
			DocumentNumber: doc.DocumentNumber,
			Title:          doc.Title,
			Summary:        summary,
			SourceURL:      doc.HTMLURL,
			PublishedAt:    pubDate,
			RawData: models.JSONMap{
				"abstract":  doc.Abstract,
				"full_text": doc.FullText,
			},
		}

		batch = append(batch, *article)

		if len(batch) >= batchSize {
			for _, a := range batch {
				if err := s.articleRepo.Create(ctx, &a); err != nil {
					log.Printf("Failed to create article %s: %v", a.DocumentNumber, err)
					errorCount++
				} else {
					log.Printf("Created article: %s - %s", a.DocumentNumber, truncate(a.Title, 60))
					processedCount++
				}
			}
			batch = batch[:0]
		}
	}

	if len(batch) > 0 {
		for _, a := range batch {
			if err := s.articleRepo.Create(ctx, &a); err != nil {
				log.Printf("Failed to create article %s: %v", a.DocumentNumber, err)
				errorCount++
			} else {
				log.Printf("Created article: %s - %s", a.DocumentNumber, truncate(a.Title, 60))
				processedCount++
			}
		}
	}

	log.Printf("Scraper completed. Processed: %d, Skipped: %d, Errors: %d", processedCount, skippedCount, errorCount)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

func (s *ScraperService) SyncAgencies(ctx context.Context) (int, error) {
	log.Println("Syncing agencies...")

	agencies, err := s.frService.FetchAgencies(ctx)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, frAgency := range agencies {
		rawData, _ := json.Marshal(frAgency)

		now := time.Now().UTC().Format("2006-01-02T15:04:05Z07:00")
		agency := &models.Agency{
			FRAgencyID:  frAgency.ID,
			Name:        frAgency.Name,
			ShortName:   frAgency.ShortName,
			Slug:        frAgency.Slug,
			Description: frAgency.Description,
			URL:         frAgency.URL,
			JSONURL:     frAgency.JSONURL,
			ParentID:    frAgency.ParentID,
			RawData:     models.JSONMap{},
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
