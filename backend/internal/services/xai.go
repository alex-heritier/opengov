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

type XAISummarizer struct {
	baseURL string
	apiKey  string
	timeout time.Duration
	client  *http.Client
}

func NewXAISummarizer(cfg *config.Config) *XAISummarizer {
	return &XAISummarizer{
		baseURL: cfg.GrokAPIURL,
		apiKey:  cfg.GrokAPIKey,
		timeout: time.Duration(cfg.GrokTimeout) * time.Second,
		client: &http.Client{
			Timeout: time.Duration(cfg.GrokTimeout) * time.Second,
		},
	}
}

type grokRequest struct {
	Model       string        `json:"model"`
	Messages    []grokMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
}

type grokMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type grokResponse struct {
	Choices []grokChoice `json:"choices"`
}

type grokChoice struct {
	Message grokMessage `json:"message"`
}

const viralSummaryPrompt = `You are an expert at writing engaging, viral-worthy summaries of government documents and Federal Register entries.

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

func (s *XAISummarizer) Summarize(ctx context.Context, text string) (string, error) {
	if text == "" {
		return "", fmt.Errorf("text cannot be empty")
	}

	prompt := fmt.Sprintf(viralSummaryPrompt, text)

	reqBody := grokRequest{
		Model:       "grok-4-fast",
		Messages:    []grokMessage{{Role: "user", Content: prompt}},
		Temperature: 0.7,
		MaxTokens:   300,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+"/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var result grokResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no choices returned from API")
	}

	summary := result.Choices[0].Message.Content
	if summary == "" {
		return "", fmt.Errorf("empty summary returned from API")
	}

	return strings.TrimSpace(summary), nil
}
