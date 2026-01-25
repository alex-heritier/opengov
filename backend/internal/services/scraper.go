package services

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/alex/opengov-go/internal/config"
	"github.com/alex/opengov-go/internal/constants"
	"github.com/alex/opengov-go/internal/models"
	"github.com/alex/opengov-go/internal/repository"
)

const batchSize = 50

type ScraperService struct {
	frService   *FederalRegisterService
	summarizer  Summarizer
	docSvc      *FederalRegisterDocumentService
	docRepo     *repository.FederalRegisterDocumentRepository
	agencyRepo  *repository.AgencyRepository
	scraperDays int
}

func NewScraperService(cfg *config.Config, frService *FederalRegisterService, summarizer Summarizer, docSvc *FederalRegisterDocumentService, agencyRepo *repository.AgencyRepository) *ScraperService {
	return &ScraperService{
		frService:   frService,
		summarizer:  summarizer,
		docSvc:      docSvc,
		docRepo:     nil,
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

	var batch []models.FederalRegisterDocument

	for _, doc := range docs {
		select {
		case <-ctx.Done():
			log.Println("Scraper cancelled mid-run, stopping...")
			return
		default:
		}

		exists, _ := s.docSvc.ExistsByUniqueKey(ctx, constants.SourceTypeFederalRegister+":"+doc.DocumentNumber)

		if exists {
			log.Printf("Skipping duplicate: %s", doc.DocumentNumber)
			skippedCount++
			continue
		}

		abstract := ""
		if doc.Abstract != nil {
			abstract = *doc.Abstract
		}
		if abstract == "" && doc.Excerpts != nil {
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

		newDoc := &models.FederalRegisterDocument{
			Source:         constants.SourceTypeFederalRegister,
			SourceID:       doc.DocumentNumber,
			UniqueKey:      constants.SourceTypeFederalRegister + ":" + doc.DocumentNumber,
			DocumentNumber: doc.DocumentNumber,
			Title:          doc.Title,
			Agency:         agencyPtr,
			Summary:        analysis.Summary,
			Keypoints:      analysis.Keypoints,
			ImpactScore:    &analysis.ImpactScore,
			PoliticalScore: &analysis.PoliticalScore,
			SourceURL:      doc.HTMLURL,
			PublishedAt:    pubDate,
			DocumentType:   &doc.Type,
			PDFURL:         &doc.PDFURL,
			RawData: models.JSONMap{
				"abstract":                  doc.Abstract,
				"excerpts":                  doc.Excerpts,
				"pdf_url":                   doc.PDFURL,
				"public_inspection_pdf_url": doc.PublicInspectionPDFURL,
				"type":                      doc.Type,
				"agencies":                  doc.Agencies,
			},
		}

		batch = append(batch, *newDoc)

		if len(batch) >= batchSize {
			for _, a := range batch {
				_, err := s.docSvc.CreateFromScrape(ctx, &a)
				if err != nil {
					log.Printf("Failed to create document %s: %v", a.DocumentNumber, err)
					errorCount++
				} else {
					log.Printf("Created document: %s - %s", a.DocumentNumber, truncate(a.Title, 60))
					processedCount++
				}
			}
			batch = batch[:0]
		}
	}

	if len(batch) > 0 {
		for _, a := range batch {
			_, err := s.docSvc.CreateFromScrape(ctx, &a)
			if err != nil {
				log.Printf("Failed to create document %s: %v", a.DocumentNumber, err)
				errorCount++
			} else {
				log.Printf("Created document: %s - %s", a.DocumentNumber, truncate(a.Title, 60))
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

		now := time.Now().UTC()
		agency := &models.Agency{
			FRAgencyID:  frAgency.ID,
			RawName:     frAgency.RawName,
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
