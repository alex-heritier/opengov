package services

import (
	"context"
)

type MockSummarizer struct{}

func (s *MockSummarizer) Analyze(ctx context.Context, title, abstract, agency string) (*AIAnalysis, error) {
	summary := "This document relates to government activity."
	if abstract != "" {
		if len(abstract) <= 100 {
			summary = "This document relates to government activity. " + abstract + "..."
		} else {
			summary = "This document relates to government activity. " + abstract[:100] + "..."
		}
	}

	return &AIAnalysis{
		Summary: summary,
		Keypoints: []string{
			"Key regulatory update from " + agency,
			"May affect compliance requirements",
			"Public comment period may apply",
		},
		ImpactScore:    "medium",
		PoliticalScore: 0,
	}, nil
}
