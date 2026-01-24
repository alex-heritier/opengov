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
	model   string
	timeout time.Duration
	client  *http.Client
}

func NewXAISummarizer(cfg *config.Config) *XAISummarizer {
	return &XAISummarizer{
		baseURL: cfg.GrokAPIURL,
		apiKey:  cfg.GrokAPIKey,
		model:   cfg.GrokModel,
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

const analysisPrompt = `You are an expert at analyzing government documents and Federal Register entries. Analyze the following document and provide a structured analysis.

Document Title: %s
Agency: %s
Abstract: %s

Provide your analysis as a JSON object with exactly these fields:
{
  "summary": "A short, punchy summary (1-2 sentences max, under 280 chars) that captures the essence and why it matters to everyday Americans. Be clear, accessible, avoid jargon.",
  "keypoints": ["Key point 1", "Key point 2", "Key point 3"],
  "impact_score": "low|medium|high",
  "political_score": <number from -100 to 100>
}

Guidelines:
- summary: Focus on human impact, make it engaging and viral-worthy
- keypoints: 3-5 bullet points of the most important takeaways
- impact_score: "low" = routine bureaucratic update, "medium" = noteworthy policy change, "high" = major news that affects many Americans
- political_score: -100 = strongly left/progressive, 0 = neutral/bipartisan, 100 = strongly right/conservative

Return ONLY the JSON object, no other text.`

type analysisResponse struct {
	Summary        string   `json:"summary"`
	Keypoints      []string `json:"keypoints"`
	ImpactScore    string   `json:"impact_score"`
	PoliticalScore int      `json:"political_score"`
}

func extractJSON(content string) (string, error) {
	trimmed := strings.TrimSpace(content)
	if strings.HasPrefix(trimmed, "```") {
		trimmed = strings.TrimPrefix(trimmed, "```")
		trimmed = strings.TrimSpace(trimmed)
		lowered := strings.ToLower(trimmed)
		if strings.HasPrefix(lowered, "json") {
			trimmed = strings.TrimSpace(trimmed[len("json"):])
		}
		if strings.HasSuffix(trimmed, "```") {
			trimmed = strings.TrimSuffix(trimmed, "```")
			trimmed = strings.TrimSpace(trimmed)
		}
	}

	start := strings.Index(trimmed, "{")
	end := strings.LastIndex(trimmed, "}")
	if start == -1 || end == -1 || end <= start {
		return "", fmt.Errorf("no JSON object found in response")
	}

	return trimmed[start : end+1], nil
}

func (s *XAISummarizer) Analyze(ctx context.Context, title, abstract, agency string) (*AIAnalysis, error) {
	if abstract == "" && title == "" {
		return nil, fmt.Errorf("title and abstract cannot both be empty")
	}

	prompt := fmt.Sprintf(analysisPrompt, title, agency, abstract)

	reqBody := grokRequest{
		Model:       s.model,
		Messages:    []grokMessage{{Role: "user", Content: prompt}},
		Temperature: 0.7,
		MaxTokens:   800,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+"/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var result grokResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned from API")
	}

	content := strings.TrimSpace(result.Choices[0].Message.Content)
	if content == "" {
		return nil, fmt.Errorf("empty response from API")
	}

	// Parse JSON response
	var analysis analysisResponse
	jsonPayload, err := extractJSON(content)
	if err != nil {
		return nil, fmt.Errorf("failed to extract JSON from AI response: %w", err)
	}
	if err := json.Unmarshal([]byte(jsonPayload), &analysis); err != nil {
		return nil, fmt.Errorf("failed to parse AI response as JSON: %w", err)
	}

	// Validate and clamp political score
	if analysis.PoliticalScore < -100 {
		analysis.PoliticalScore = -100
	}
	if analysis.PoliticalScore > 100 {
		analysis.PoliticalScore = 100
	}

	// Validate impact score
	switch analysis.ImpactScore {
	case "low", "medium", "high":
		// valid
	default:
		analysis.ImpactScore = "medium"
	}

	return &AIAnalysis{
		Summary:        analysis.Summary,
		Keypoints:      analysis.Keypoints,
		ImpactScore:    analysis.ImpactScore,
		PoliticalScore: analysis.PoliticalScore,
	}, nil
}
