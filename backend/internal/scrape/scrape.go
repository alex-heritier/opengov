package scrape

import (
	"context"

	"github.com/alex/opengov-go/internal/transport"
)

// ScrapeResult holds scraped document data along with raw payload.
type ScrapeResult struct {
	PolicyDocument transport.ScrapedPolicyDocument
	RawResult      []byte
}

// PolicyDocumentScraper defines the interface for document scrapers.
type PolicyDocumentScraper interface {
	Scrape(ctx context.Context, daysLookback int) ([]ScrapeResult, error)
}
