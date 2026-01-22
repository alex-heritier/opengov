package services

import (
	"context"
	"log"

	"github.com/alex/opengov-go/internal/config"
)

type Summarizer interface {
	Summarize(ctx context.Context, text string) (string, error)
}

func NewSummarizer(cfg *config.Config) Summarizer {
	if cfg.UseMockGrok {
		return &MockSummarizer{}
	}
	if cfg.GrokAPIKey == "" {
		log.Fatal("GROK_API_KEY is required when USE_MOCK_GROK=false")
	}
	return NewXAISummarizer(cfg)
}
