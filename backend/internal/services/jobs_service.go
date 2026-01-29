package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/alex/opengov-go/internal/client"
	"github.com/alex/opengov-go/internal/config"
	"github.com/alex/opengov-go/internal/constants"
	"github.com/alex/opengov-go/internal/db"
	"github.com/alex/opengov-go/internal/domain"
	"github.com/alex/opengov-go/internal/repository"
	"github.com/alex/opengov-go/internal/scrape"
)

type JobsService struct {
	cfg *config.Config
	db  *db.DB

	agencyRepo *repository.AgencyRepository
	rawRepo    *repository.RawPolicyDocumentRepository
	docRepo    *repository.PolicyDocumentRepository
	feedRepo   *repository.FeedRepository

	fedregClient  *client.FederalRegisterClient
	docScrapers   []scrape.PolicyDocumentScraper
	agencySyncSvc *AgencySyncService
}

func NewJobsService(
	cfg *config.Config,
	database *db.DB,
	agencyRepo *repository.AgencyRepository,
	rawRepo *repository.RawPolicyDocumentRepository,
	docRepo *repository.PolicyDocumentRepository,
	feedRepo *repository.FeedRepository,
	frClient *client.FederalRegisterClient,
) *JobsService {
	agencySyncSvc := NewAgencySyncService(frClient, agencyRepo)

	return &JobsService{
		cfg: cfg,
		db:  database,

		agencyRepo: agencyRepo,
		rawRepo:    rawRepo,
		docRepo:    docRepo,
		feedRepo:   feedRepo,

		fedregClient:  frClient,
		docScrapers:   []scrape.PolicyDocumentScraper{scrape.NewFedregScraper(frClient)},
		agencySyncSvc: agencySyncSvc,
	}
}

func (s *JobsService) Migrate() error {
	return s.db.RunMigrations()
}

func (s *JobsService) SyncAgencies(ctx context.Context) (int, error) {
	return s.agencySyncSvc.SyncAgencies(ctx)
}

// ScrapeRaw ingests raw upstream JSON into raw_policy_documents with no policy_document_id.
func (s *JobsService) ScrapeRaw(ctx context.Context) (processed int, skipped int, err error) {
	log.Println("Starting raw ingestion scrape...")

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	fetchedAt := time.Now().UTC()

	for _, retriever := range s.docScrapers {
		results, err := retriever.Scrape(ctx, s.cfg.ScraperDaysLookback)
		if err != nil {
			return processed, skipped, fmt.Errorf("failed to scrape documents: %w", err)
		}

		for _, r := range results {
			ins, err := s.rawRepo.Create(ctx, tx, constants.SourceTypeFederalRegister, r.PolicyDocument.DocumentNumber, r.RawResult, fetchedAt, nil)
			if err != nil {
				return processed, skipped, err
			}
			if ins {
				processed++
			} else {
				skipped++
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return processed, skipped, fmt.Errorf("failed to commit raw ingestion: %w", err)
	}

	log.Printf("Raw ingestion completed. Inserted: %d, Skipped: %d", processed, skipped)
	return processed, skipped, nil
}

func (s *JobsService) Canonicalize(ctx context.Context, batchSize int) (linked int, err error) {
	if batchSize <= 0 {
		batchSize = 200
	}

	log.Println("Starting canonicalization...")
	for {
		rows, err := s.rawRepo.ListUnlinked(ctx, batchSize)
		if err != nil {
			return linked, err
		}
		if len(rows) == 0 {
			break
		}

		for _, raw := range rows {
			select {
			case <-ctx.Done():
				return linked, ctx.Err()
			default:
			}

			if _, err := s.canonicalizeOne(ctx, raw); err != nil {
				return linked, err
			}
			linked++
		}
	}

	log.Printf("Canonicalization completed. Linked: %d", linked)
	return linked, nil
}

func (s *JobsService) canonicalizeOne(ctx context.Context, raw repository.UnlinkedRawPolicyDocumentRow) (policyDocID int64, err error) {
	var frDoc client.FederalRegisterDocument
	if err := json.Unmarshal(raw.RawData, &frDoc); err != nil {
		return 0, fmt.Errorf("failed to unmarshal raw_policy_documents(%d) into federal register document: %w", raw.ID, err)
	}

	publishedAt, err := time.Parse("2006-01-02", frDoc.PublicationDate)
	if err != nil {
		return 0, fmt.Errorf("invalid publication_date for raw_policy_documents(%d): %w", raw.ID, err)
	}

	summary := derivePlaceholderSummary(frDoc)
	if summary == "" {
		summary = "Pending summary."
	}

	var agencyPtr *string
	if len(frDoc.Agencies) > 0 && frDoc.Agencies[0].Name != "" {
		a := frDoc.Agencies[0].Name
		agencyPtr = &a
	}

	doc := &domain.PolicyDocument{
		SourceKey:      raw.SourceKey,
		ExternalID:     raw.ExternalID,
		FetchedAt:      raw.FetchedAt,
		Title:          frDoc.Title,
		Agency:         agencyPtr,
		Summary:        summary,
		Keypoints:      nil,
		ImpactScore:    nil,
		PoliticalScore: nil,
		SourceURL:      frDoc.HTMLURL,
		PublishedAt:    publishedAt,
		DocumentType:   &frDoc.Type,
		PDFURL:         frDoc.PDFURL,
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to begin canonicalization tx: %w", err)
	}
	defer tx.Rollback()

	id, err := s.docRepo.UpsertCanonical(ctx, tx, doc)
	if err != nil {
		return 0, err
	}

	if err := s.rawRepo.LinkToPolicyDocument(ctx, tx, raw.ID, id); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit canonicalization tx: %w", err)
	}

	return id, nil
}

func derivePlaceholderSummary(frDoc client.FederalRegisterDocument) string {
	// Mirror legacy behavior: prefer excerpts over abstract, truncate to ~1000 chars.
	s := ""
	if frDoc.Abstract != nil {
		s = *frDoc.Abstract
	}
	if frDoc.Excerpts != nil && *frDoc.Excerpts != "" {
		s = *frDoc.Excerpts
	}
	if len(s) > 1000 {
		s = s[:1000]
	}
	return s
}

func needsEnrichment(d *domain.PolicyDocument) bool {
	if d.ImpactScore == nil {
		return true
	}
	if d.PoliticalScore == nil {
		return true
	}
	return len(d.Keypoints) == 0
}

// Enrich is the enrichment stage. For now, it is implemented as a dry-run and does not
// call any external AI APIs or write any changes. It reports how many documents would
// be enriched based on missing AI fields.
func (s *JobsService) Enrich(ctx context.Context, batchSize int) (wouldEnrich int, err error) {
	if batchSize <= 0 {
		batchSize = 200
	}

	log.Println("Starting enrichment (dry-run; no writes)...")
	for {
		docs, err := s.docRepo.ListNeedingEnrichment(ctx, batchSize)
		if err != nil {
			return wouldEnrich, err
		}
		if len(docs) == 0 {
			break
		}

		for _, d := range docs {
			select {
			case <-ctx.Done():
				return wouldEnrich, ctx.Err()
			default:
			}

			// Guardrail: ensure the in-memory predicate matches expectations too.
			if needsEnrichment(d) {
				wouldEnrich++
			}
		}

		// Since we are not writing anything yet, stop after one batch to avoid
		// repeatedly returning the same set of documents.
		break
	}

	log.Printf("Enrichment dry-run completed. Would enrich: %d", wouldEnrich)
	return wouldEnrich, nil
}

func (s *JobsService) Materialize(ctx context.Context, batchSize int) (upserted int, err error) {
	if batchSize <= 0 {
		batchSize = 500
	}

	log.Println("Starting materialization...")
	for {
		docs, err := s.docRepo.ListNeedingMaterialization(ctx, batchSize)
		if err != nil {
			return upserted, err
		}
		if len(docs) == 0 {
			break
		}

		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return upserted, fmt.Errorf("failed to begin materialization tx: %w", err)
		}

		for _, d := range docs {
			select {
			case <-ctx.Done():
				_ = tx.Rollback()
				return upserted, ctx.Err()
			default:
			}

			impactScore := ""
			if d.ImpactScore != nil {
				impactScore = *d.ImpactScore
			}

			if err := s.feedRepo.UpsertFeedEntryByPolicyDocID(
				ctx, tx, d.ID,
				d.Title, d.Summary, d.Keypoints,
				d.PoliticalScore, impactScore,
				d.SourceURL, d.PublishedAt,
			); err != nil {
				_ = tx.Rollback()
				return upserted, err
			}
			upserted++
		}

		if err := tx.Commit(); err != nil {
			_ = tx.Rollback()
			return upserted, fmt.Errorf("failed to commit materialization tx: %w", err)
		}
	}

	log.Printf("Materialization completed. Upserted: %d", upserted)
	return upserted, nil
}

func (s *JobsService) Pipeline(ctx context.Context) error {
	if _, err := s.SyncAgencies(ctx); err != nil {
		return err
	}
	if _, _, err := s.ScrapeRaw(ctx); err != nil {
		return err
	}
	if _, err := s.Canonicalize(ctx, 200); err != nil {
		return err
	}
	if _, err := s.Enrich(ctx, 200); err != nil {
		return err
	}
	if _, err := s.Materialize(ctx, 500); err != nil {
		return err
	}
	return nil
}
