package client

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

type FederalRegisterDocument struct {
	ID                     int        `json:"id"`
	DocumentNumber         string     `json:"document_number"`
	Title                  string     `json:"title"`
	Type                   string     `json:"type"`
	Abstract               *string    `json:"abstract"`
	HTMLURL                string     `json:"html_url"`
	PublicationDate        string     `json:"publication_date"`
	PDFURL                 *string    `json:"pdf_url"`
	PublicInspectionPDFURL *string    `json:"public_inspection_pdf_url"`
	Excerpts               *string    `json:"excerpts"`
	Agencies               []FRAgency `json:"agencies"`
}

type FRAgency struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	ShortName   string  `json:"short_name"`
	Slug        string  `json:"slug"`
	URL         string  `json:"url"`
	ParentID    *int    `json:"parent_id"`
	Description *string `json:"description"`
	RawName     string  `json:"raw_name"`
	JSONURL     string  `json:"json_url"`
}

type FederalRegisterRecordsResponse struct {
	Description string                    `json:"description"`
	Count       int                       `json:"count"`
	TotalPages  int                       `json:"total_pages"`
	NextPageURL string                    `json:"next_page_url,omitempty"`
	Results     []FederalRegisterDocument `json:"results"`
}

type FederalRegisterDocumentWithRaw struct {
	Document FederalRegisterDocument
	RawJSON  []byte
}

type FRAgenciesResponse []FRAgency

type FederalRegisterClient struct {
	baseURL  string
	timeout  time.Duration
	perPage  int
	maxPages int
	client   *http.Client
}

func NewFederalRegisterClient(cfg *config.Config) *FederalRegisterClient {
	return &FederalRegisterClient{
		baseURL:  cfg.FederalRegisterAPIURL,
		timeout:  time.Duration(cfg.FederalRegisterTimeout) * time.Second,
		perPage:  cfg.FederalRegisterPerPage,
		maxPages: cfg.FederalRegisterMaxPages,
		client: &http.Client{
			Timeout: time.Duration(cfg.FederalRegisterTimeout) * time.Second,
		},
	}
}

func (s *FederalRegisterClient) Scrape(ctx context.Context, days int) ([]FederalRegisterDocumentWithRaw, error) {
	endDate := time.Now().UTC()
	startDate := endDate.AddDate(0, 0, -days)

	params := url.Values{
		"per_page":                      {fmt.Sprintf("%d", s.perPage)},
		"page":                          {"1"},
		"filter[publication_date][gte]": {startDate.Format("2006-01-02")},
		"filter[publication_date][lte]": {endDate.Format("2006-01-02")},
	}

	var allDocs []FederalRegisterDocumentWithRaw

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

		bodyBytes, _ := io.ReadAll(resp.Body)
		var result FederalRegisterRecordsResponse
		if err := json.Unmarshal(bodyBytes, &result); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}

		for _, frDoc := range result.Results {
			docRaw, _ := json.Marshal(frDoc)
			allDocs = append(allDocs, FederalRegisterDocumentWithRaw{
				Document: frDoc,
				RawJSON:  docRaw,
			})
		}

		if len(result.Results) < s.perPage {
			break
		}

		time.Sleep(500 * time.Millisecond)
	}

	return allDocs, nil
}

func (s *FederalRegisterClient) FetchAgencies(ctx context.Context) ([]FRAgency, error) {
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
