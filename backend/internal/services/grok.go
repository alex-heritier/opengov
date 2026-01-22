package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/alex/opengov-go/internal/config"
)

type GrokService struct {
	baseURL   string
	apiKey    string
	timeout   time.Duration
	maxTokens int
	client    *http.Client
	useMock   bool
}

type GrokRequest struct {
	Model       string        `json:"model"`
	Messages    []GrokMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
}

type GrokMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type GrokResponse struct {
	Choices []GrokChoice `json:"choices"`
}

type GrokChoice struct {
	Message GrokMessage `json:"message"`
}

const (
	viralSummaryPrompt = `You are an expert at writing engaging, viral-worthy summaries of government documents and Federal Register entries.

Your task is to create a short, punchy summary (1-2 sentences max) that captures the essence of what the government is doing and why it matters to everyday Americans.

Guidelines:
- Be clear and accessible (avoid jargon)
- Focus on human impact
- Make it engaging and interesting
- Keep it under 280 characters when possible
- Start with the most important information

Document to summarize:
%s

Generate only the summary, nothing else.`
)

func NewGrokService(cfg *config.Config) *GrokService {
	return &GrokService{
		baseURL:   cfg.GrokAPIURL,
		apiKey:    cfg.GrokAPIKey,
		timeout:   time.Duration(cfg.GrokTimeout) * time.Second,
		maxTokens: 300,
		client: &http.Client{
			Timeout: time.Duration(cfg.GrokTimeout) * time.Second,
		},
		useMock: cfg.UseMockGrok,
	}
}

func (s *GrokService) Summarize(ctx context.Context, text string) (string, error) {
	if s.useMock {
		return s.mockSummarize(text)
	}

	if s.apiKey == "" {
		return s.fallbackSummarize(text), nil
	}

	if text == "" {
		return "No summary available.", nil
	}

	prompt := fmt.Sprintf(viralSummaryPrompt, text)

	reqBody := GrokRequest{
		Model:       "grok-4-fast",
		Messages:    []GrokMessage{{Role: "user", Content: prompt}},
		Temperature: 0.7,
		MaxTokens:   s.maxTokens,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return s.fallbackSummarize(text), fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+"/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return s.fallbackSummarize(text), fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return s.fallbackSummarize(text), fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return s.fallbackSummarize(text), fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var result GrokResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return s.fallbackSummarize(text), fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Choices) == 0 {
		return s.fallbackSummarize(text), nil
	}

	summary := result.Choices[0].Message.Content
	if summary == "" {
		return s.fallbackSummarize(text), nil
	}

	return strings.TrimSpace(summary), nil
}

func (s *GrokService) mockSummarize(text string) (string, error) {
	if text == "" {
		return "No summary available.", nil
	}
	return "This document relates to government activity. " + text[:min(100, len(text))] + "...", nil
}

func (s *GrokService) fallbackSummarize(text string) string {
	if len(text) > 200 {
		return text[:200] + "..."
	}
	return text
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
