package scrape

import (
	"context"

	"github.com/alex/opengov-go/internal/client"
	"github.com/alex/opengov-go/internal/domain"
	"github.com/alex/opengov-go/internal/transport"
)

type FedregScraper struct {
	client *client.FederalRegisterClient
}

func NewFedregScraper(client *client.FederalRegisterClient) *FedregScraper {
	return &FedregScraper{
		client: client,
	}
}

func (s *FedregScraper) Scrape(ctx context.Context, daysLookback int) ([]ScrapeResult, error) {
	docs, err := s.client.Scrape(ctx, daysLookback)
	if err != nil {
		return nil, err
	}

	results := make([]ScrapeResult, len(docs))
	for i, frDoc := range docs {
		doc := transport.ScrapedPolicyDocument{
			DocumentNumber:         frDoc.Document.DocumentNumber,
			Title:                  frDoc.Document.Title,
			Type:                   frDoc.Document.Type,
			Abstract:               frDoc.Document.Abstract,
			HTMLURL:                frDoc.Document.HTMLURL,
			PublicationDate:        frDoc.Document.PublicationDate,
			PDFURL:                 frDoc.Document.PDFURL,
			PublicInspectionPDFURL: frDoc.Document.PublicInspectionPDFURL,
			Excerpts:               frDoc.Document.Excerpts,
			Agencies:               transformAgencies(frDoc.Document.Agencies),
		}
		results[i] = ScrapeResult{
			PolicyDocument: doc,
			RawResult:      frDoc.RawJSON,
		}
	}
	return results, nil
}

func transformAgencies(frAgencies []client.FRAgency) []domain.Agency {
	agencies := make([]domain.Agency, len(frAgencies))
	for i, frAgency := range frAgencies {
		agencies[i] = domain.Agency{
			FRAgencyID:  frAgency.ID,
			Name:        frAgency.Name,
			ShortName:   &frAgency.ShortName,
			Slug:        frAgency.Slug,
			URL:         &frAgency.URL,
			ParentID:    frAgency.ParentID,
			Description: frAgency.Description,
			RawName:     frAgency.RawName,
			JSONURL:     &frAgency.JSONURL,
		}
	}
	return agencies
}

func (s *FedregScraper) Client() *client.FederalRegisterClient {
	return s.client
}
