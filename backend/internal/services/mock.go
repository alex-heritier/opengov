package services

import (
	"context"
)

type MockSummarizer struct{}

func (s *MockSummarizer) Summarize(ctx context.Context, text string) (string, error) {
	if text == "" {
		return "No summary available.", nil
	}
	if len(text) <= 100 {
		return "This document relates to government activity. " + text + "...", nil
	}
	return "This document relates to government activity. " + text[:100] + "...", nil
}
