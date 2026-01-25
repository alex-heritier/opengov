package services

import (
	"context"
	"log"

	"github.com/alex/opengov-go/internal/config"
)

// AIAnalysis contains all AI-generated fields for an article
type AIAnalysis struct {
	Summary        string   // 1-2 sentence viral summary
	Keypoints      []string // Key takeaways from the document
	ImpactScore    string   // low, medium, high
	PoliticalScore int      // -100 (left) to 100 (right)
}

type Summarizer interface {
	Analyze(ctx context.Context, title, abstract, agency string) (*AIAnalysis, error)
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
