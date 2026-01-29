package services

import (
	"context"
	"encoding/json"
	"log"

	"github.com/alex/opengov-go/internal/client"
	"github.com/alex/opengov-go/internal/db/dbtypes"
	"github.com/alex/opengov-go/internal/domain"
	"github.com/alex/opengov-go/internal/repository"
)

// AgencySyncService syncs Federal Register agencies into the local agencies table.
type AgencySyncService struct {
	frClient   *client.FederalRegisterClient
	agencyRepo *repository.AgencyRepository
}

func NewAgencySyncService(frClient *client.FederalRegisterClient, agencyRepo *repository.AgencyRepository) *AgencySyncService {
	return &AgencySyncService{
		frClient:   frClient,
		agencyRepo: agencyRepo,
	}
}

func (s *AgencySyncService) SyncAgencies(ctx context.Context) (int, error) {
	log.Println("Syncing agencies...")

	frAgencies, err := s.frClient.FetchAgencies(ctx)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, frAgency := range frAgencies {
		rawData, _ := json.Marshal(frAgency)

		frAgencyID := int64(frAgency.ID)

		var parentID *int64
		if frAgency.ParentID != nil {
			p := int64(*frAgency.ParentID)
			parentID = &p
		}

		var shortName *string
		if frAgency.ShortName != "" {
			sn := frAgency.ShortName
			shortName = &sn
		}

		var url *string
		if frAgency.URL != "" {
			u := frAgency.URL
			url = &u
		}

		var jsonURL *string
		if frAgency.JSONURL != "" {
			j := frAgency.JSONURL
			jsonURL = &j
		}

		agency := &domain.Agency{
			FRAgencyID:  frAgencyID,
			RawName:     frAgency.RawName,
			Name:        frAgency.Name,
			ShortName:   shortName,
			Slug:        frAgency.Slug,
			Description: frAgency.Description,
			URL:         url,
			JSONURL:     jsonURL,
			ParentID:    parentID,
			RawData:     dbtypes.JSONMap{},
		}
		_ = json.Unmarshal(rawData, &agency.RawData)

		if err := s.agencyRepo.Upsert(ctx, agency); err != nil {
			log.Printf("Failed to upsert agency %s: %v", frAgency.Name, err)
			continue
		}
		count++
	}

	log.Printf("Synced %d agencies", count)
	return count, nil
}
