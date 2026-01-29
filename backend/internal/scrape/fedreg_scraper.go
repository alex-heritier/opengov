package scrape

import (
	"context"

	"github.com/alex/opengov-go/internal/client"
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

func transformAgencies(frAgencies []client.FRAgency) []transport.ScrapedAgency {
	agencies := make([]transport.ScrapedAgency, len(frAgencies))
	for i, frAgency := range frAgencies {
		var parentID *int64
		if frAgency.ParentID != nil {
			p := int64(*frAgency.ParentID)
			parentID = &p
		}

		var shortName *string
		if frAgency.ShortName != "" {
			s := frAgency.ShortName
			shortName = &s
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

		agencies[i] = transport.ScrapedAgency{
			FRAgencyID:  int64(frAgency.ID),
			Name:        frAgency.Name,
			ShortName:   shortName,
			Slug:        frAgency.Slug,
			URL:         url,
			ParentID:    parentID,
			Description: frAgency.Description,
			RawName:     frAgency.RawName,
			JSONURL:     jsonURL,
		}
	}
	return agencies
}

func (s *FedregScraper) Client() *client.FederalRegisterClient {
	return s.client
}
