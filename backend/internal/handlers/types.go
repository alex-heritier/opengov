package handlers

import "time"

type StatsResponse struct {
	TotalArticles  int        `json:"total_articles"`
	LastScrapeTime *time.Time `json:"last_scrape_time,omitempty"`
	LastScrapeAge  string     `json:"last_scrape_human,omitempty"`
}
