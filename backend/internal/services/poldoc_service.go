package services

import (
	"context"
	"fmt"
	"time"

	"github.com/alex/opengov-go/internal/constants"
	"github.com/alex/opengov-go/internal/db"
	"github.com/alex/opengov-go/internal/models"
	"github.com/alex/opengov-go/internal/repository"
	"github.com/lib/pq"
)

type PolicyDocumentService struct {
	docRepo  *repository.PolicyDocumentRepository
	feedRepo *repository.FeedRepository
	rawRepo  *repository.RawEntryRepository
	db       *db.DB
}

func NewPolicyDocumentService(docRepo *repository.PolicyDocumentRepository, feedRepo *repository.FeedRepository, rawRepo *repository.RawEntryRepository, db *db.DB) *PolicyDocumentService {
	return &PolicyDocumentService{
		docRepo:  docRepo,
		feedRepo: feedRepo,
		rawRepo:  rawRepo,
		db:       db,
	}
}

func (s *PolicyDocumentService) CreateFromScrape(ctx context.Context, doc *models.PolicyDocument, rawPayload []byte, fetchedAt time.Time) (*models.PolicyDocument, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	feedEntryID, err := s.feedRepo.CreateFeedEntry(ctx, tx, constants.SourceTypeFederalRegister, doc.Title, doc.Summary, doc.Keypoints, doc.PoliticalScore, "", doc.SourceURL, doc.PublishedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create feed entry: %w", err)
	}

	err = s.docRepo.Create(ctx, tx, doc, feedEntryID)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			tx.Rollback()
			existing, fetchErr := s.docRepo.GetByUniqueKey(ctx, doc.UniqueKey)
			if fetchErr != nil {
				return nil, fmt.Errorf("failed to fetch existing document: %w", fetchErr)
			}
			return existing, nil
		}
		return nil, fmt.Errorf("failed to create document: %w", err)
	}

	err = s.rawRepo.Create(ctx, tx, doc.Source, doc.DocumentNumber, rawPayload, fetchedAt, doc.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create raw entry: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return doc, nil
}

func (s *PolicyDocumentService) Update(ctx context.Context, id int, updates *models.PolicyDocument) (*models.PolicyDocument, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	existing, err := s.docRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}
	if existing == nil {
		return nil, fmt.Errorf("document not found")
	}

	if updates.Title != "" {
		existing.Title = updates.Title
	}
	if updates.Summary != "" {
		existing.Summary = updates.Summary
	}
	if updates.Keypoints != nil {
		existing.Keypoints = updates.Keypoints
	}
	if updates.PoliticalScore != nil {
		existing.PoliticalScore = updates.PoliticalScore
	}
	if updates.ImpactScore != nil {
		existing.ImpactScore = updates.ImpactScore
	}
	if updates.Agency != nil {
		existing.Agency = updates.Agency
	}

	err = s.docRepo.Update(ctx, tx, existing)
	if err != nil {
		return nil, fmt.Errorf("failed to update document: %w", err)
	}

	err = s.feedRepo.UpdateFeedEntry(ctx, tx, existing.FeedEntryID, existing.Title, existing.Summary, existing.Keypoints, existing.PoliticalScore, "", existing.SourceURL, existing.PublishedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to update feed entry: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return existing, nil
}

func (s *PolicyDocumentService) GetByID(ctx context.Context, id int) (*models.PolicyDocument, error) {
	return s.docRepo.GetByID(ctx, id)
}

func (s *PolicyDocumentService) GetByDocumentNumber(ctx context.Context, docNumber string) (*models.PolicyDocument, error) {
	return s.docRepo.GetByDocumentNumber(ctx, docNumber)
}

func (s *PolicyDocumentService) ExistsByUniqueKey(ctx context.Context, uniqueKey string) (bool, error) {
	return s.docRepo.ExistsByUniqueKey(ctx, uniqueKey)
}

func (s *PolicyDocumentService) GetByUniqueKey(ctx context.Context, uniqueKey string) (*models.PolicyDocument, error) {
	return s.docRepo.GetByUniqueKey(ctx, uniqueKey)
}

func (s *PolicyDocumentService) Count(ctx context.Context) (int, error) {
	return s.docRepo.Count(ctx)
}
