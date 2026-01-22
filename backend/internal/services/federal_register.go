package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/alex/opengov-go/internal/config"
)

type FederalRegisterService struct {
	baseURL  string
	timeout  time.Duration
	perPage  int
	maxPages int
	client   *http.Client
}

type FRDocument struct {
	DocumentNumber  string `json:"document_number"`
	Title           string `json:"title"`
	Abstract        string `json:"abstract,omitempty"`
	FullText        string `json:"full_text,omitempty"`
	HTMLURL         string `json:"html_url"`
	PublicationDate string `json:"publication_date"`
}

type FRAgency struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	ShortName   *string `json:"short_name"`
	Slug        string  `json:"slug"`
	Description *string `json:"description,omitempty"`
	URL         *string `json:"url,omitempty"`
	JSONURL     *string `json:"json_url,omitempty"`
	ParentID    *int    `json:"parent_id,omitempty"`
}

type FRDocumentsResponse struct {
	Results        []FRDocument `json:"results"`
	TotalDocuments int          `json:"total_documents"`
}

type FRAgenciesResponse []FRAgency

func NewFederalRegisterService(cfg *config.Config) *FederalRegisterService {
	return &FederalRegisterService{
		baseURL:  cfg.FederalRegisterAPIURL,
		timeout:  time.Duration(cfg.FederalRegisterTimeout) * time.Second,
		perPage:  cfg.FederalRegisterPerPage,
		maxPages: cfg.FederalRegisterMaxPages,
		client: &http.Client{
			Timeout: time.Duration(cfg.FederalRegisterTimeout) * time.Second,
		},
	}
}

func (s *FederalRegisterService) FetchRecentDocuments(ctx context.Context, days int) ([]FRDocument, error) {
	endDate := time.Now().UTC()
	startDate := endDate.AddDate(0, 0, -days)

	params := url.Values{
		"per_page":                      {fmt.Sprintf("%d", s.perPage)},
		"page":                          {"1"},
		"filter[publication_date][gte]": {startDate.Format("2006-01-02")},
		"filter[publication_date][lte]": {endDate.Format("2006-01-02")},
	}

	var allDocs []FRDocument

	for page := 1; page <= s.maxPages; page++ {
		params.Set("page", fmt.Sprintf("%d", page))

		reqURL := fmt.Sprintf("%s/documents?%s", s.baseURL, params.Encode())
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		resp, err := s.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("request failed: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
		}

		var result FRDocumentsResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}

		allDocs = append(allDocs, result.Results...)

		if len(result.Results) < s.perPage {
			break
		}

		time.Sleep(500 * time.Millisecond)
	}

	return allDocs, nil
}

func (s *FederalRegisterService) FetchAgencies(ctx context.Context) ([]FRAgency, error) {
	reqURL := fmt.Sprintf("%s/agencies", s.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var result FRAgenciesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}
