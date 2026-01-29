package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/alex/opengov-go/internal/db"
	"github.com/alex/opengov-go/internal/domain"
	"github.com/alex/opengov-go/internal/repository"
)

type PolicyDocumentService struct {
	docRepo    *repository.PolicyDocumentRepository
	feedRepo   *repository.FeedRepository
	sourceRepo *repository.RawPolicyDocumentRepository
	db         *db.DB
}

func NewPolicyDocumentService(docRepo *repository.PolicyDocumentRepository, feedRepo *repository.FeedRepository, sourceRepo *repository.RawPolicyDocumentRepository, db *db.DB) *PolicyDocumentService {
	return &PolicyDocumentService{
		docRepo:    docRepo,
		feedRepo:   feedRepo,
		sourceRepo: sourceRepo,
		db:         db,
	}
}

func (s *PolicyDocumentService) CreateFromScrape(ctx context.Context, doc *domain.PolicyDocument, rawPayload []byte, fetchedAt time.Time) (*domain.PolicyDocument, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	err = s.docRepo.Create(ctx, tx, doc)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateDocument) {
			existing, fetchErr := s.docRepo.GetBySourceKeyExternalID(ctx, doc.SourceKey, doc.ExternalID)
			if fetchErr != nil {
				return nil, fmt.Errorf("failed to fetch existing document: %w", fetchErr)
			}

			impactScore := ""
			if existing.ImpactScore != nil {
				impactScore = *existing.ImpactScore
			}
			if upsertErr := s.feedRepo.UpsertFeedEntryByPolicyDocID(
				ctx, tx, existing.ID,
				existing.Title, existing.Summary, existing.Keypoints,
				existing.PoliticalScore, impactScore,
				existing.SourceURL, existing.PublishedAt,
			); upsertErr != nil {
				return nil, fmt.Errorf("failed to upsert feed entry for existing doc: %w", upsertErr)
			}

			if commitErr := tx.Commit(); commitErr != nil {
				return nil, fmt.Errorf("failed to commit transaction: %w", commitErr)
			}
			return existing, nil
		}
		return nil, fmt.Errorf("failed to create document: %w", err)
	}

	impactScore := ""
	if doc.ImpactScore != nil {
		impactScore = *doc.ImpactScore
	}
	err = s.feedRepo.UpsertFeedEntryByPolicyDocID(ctx, tx, doc.ID, doc.Title, doc.Summary, doc.Keypoints, doc.PoliticalScore, impactScore, doc.SourceURL, doc.PublishedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert feed entry: %w", err)
	}

	err = s.sourceRepo.Create(ctx, tx, doc.SourceKey, doc.ExternalID, rawPayload, fetchedAt, doc.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create raw entry: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return doc, nil
}

func (s *PolicyDocumentService) Update(ctx context.Context, id int64, updates *domain.PolicyDocument) (*domain.PolicyDocument, error) {
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

	impactScore := ""
	if existing.ImpactScore != nil {
		impactScore = *existing.ImpactScore
	}

	err = s.feedRepo.UpsertFeedEntryByPolicyDocID(ctx, tx, existing.ID, existing.Title, existing.Summary, existing.Keypoints, existing.PoliticalScore, impactScore, existing.SourceURL, existing.PublishedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert feed entry: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return existing, nil
}

func (s *PolicyDocumentService) GetByID(ctx context.Context, id int64) (*domain.PolicyDocument, error) {
	return s.docRepo.GetByID(ctx, id)
}

func (s *PolicyDocumentService) ExistsBySourceKeyExternalID(ctx context.Context, sourceKey, externalID string) (bool, error) {
	return s.docRepo.ExistsBySourceKeyExternalID(ctx, sourceKey, externalID)
}

func (s *PolicyDocumentService) GetBySourceKeyExternalID(ctx context.Context, sourceKey, externalID string) (*domain.PolicyDocument, error) {
	return s.docRepo.GetBySourceKeyExternalID(ctx, sourceKey, externalID)
}

func (s *PolicyDocumentService) Count(ctx context.Context) (int, error) {
	return s.docRepo.Count(ctx)
}
