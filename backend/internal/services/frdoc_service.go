package services

import (
	"context"
	"fmt"

	"github.com/alex/opengov-go/internal/constants"
	"github.com/alex/opengov-go/internal/db"
	"github.com/alex/opengov-go/internal/models"
	"github.com/alex/opengov-go/internal/repository"
)

type FederalRegisterDocumentService struct {
	docRepo  *repository.FederalRegisterDocumentRepository
	feedRepo *repository.FeedRepository
	db       *db.DB
}

func NewFederalRegisterDocumentService(docRepo *repository.FederalRegisterDocumentRepository, feedRepo *repository.FeedRepository, db *db.DB) *FederalRegisterDocumentService {
	return &FederalRegisterDocumentService{
		docRepo:  docRepo,
		feedRepo: feedRepo,
		db:       db,
	}
}

func (s *FederalRegisterDocumentService) CreateFromScrape(ctx context.Context, doc *models.FederalRegisterDocument) (*models.FederalRegisterDocument, error) {
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
		return nil, fmt.Errorf("failed to create document: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return doc, nil
}

func (s *FederalRegisterDocumentService) Update(ctx context.Context, id int, updates *models.FederalRegisterDocument) (*models.FederalRegisterDocument, error) {
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

func (s *FederalRegisterDocumentService) GetByID(ctx context.Context, id int) (*models.FederalRegisterDocument, error) {
	return s.docRepo.GetByID(ctx, id)
}

func (s *FederalRegisterDocumentService) GetByDocumentNumber(ctx context.Context, docNumber string) (*models.FederalRegisterDocument, error) {
	return s.docRepo.GetByDocumentNumber(ctx, docNumber)
}

func (s *FederalRegisterDocumentService) ExistsByUniqueKey(ctx context.Context, uniqueKey string) (bool, error) {
	return s.docRepo.ExistsByUniqueKey(ctx, uniqueKey)
}

func (s *FederalRegisterDocumentService) Count(ctx context.Context) (int, error) {
	return s.docRepo.Count(ctx)
}
