package scrape

import (
	"context"

	"github.com/alex/opengov-go/internal/models"
)

type ScrapeResult struct {
	PolicyDocument ScrapedPolicyDocument
	RawResult      []byte
}

type PolicyDocumentScraper interface {
	Scrape(ctx context.Context, daysLookback int) ([]ScrapeResult, error)
}

// ScrapedPolicyDocument is an upstream document payload returned by a scraper.
// It is intentionally separate from the DB-backed models.
type ScrapedPolicyDocument struct {
	DocumentNumber         string
	Title                  string
	Type                   string
	Abstract               *string
	HTMLURL                string
	PublicationDate        string
	PDFURL                 *string
	PublicInspectionPDFURL *string
	Excerpts               *string
	Agencies               []models.Agency
}
