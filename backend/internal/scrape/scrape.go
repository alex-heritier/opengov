package scrape

import (
	"context"

	"github.com/alex/opengov-go/internal/models"
)

type ScrapeResult struct {
	PolicyDocument models.PolicyDocument
	RawResult      []byte
}

type PolicyDocumentScraper interface {
	Scrape(ctx context.Context, daysLookback int) ([]ScrapeResult, error)
}
