package services

import (
	"encoding/json"
	"testing"

	"github.com/alex/opengov-go/internal/client"
	"github.com/alex/opengov-go/internal/domain"
)

func TestDerivePlaceholderSummary_PrefersExcerptsOverAbstract(t *testing.T) {
	abs := "abstract text"
	ex := "excerpts text"
	doc := client.FederalRegisterDocument{
		Abstract: &abs,
		Excerpts: &ex,
	}
	if got := derivePlaceholderSummary(doc); got != ex {
		t.Fatalf("expected excerpts %q, got %q", ex, got)
	}
}

func TestDerivePlaceholderSummary_Truncates(t *testing.T) {
	long := make([]byte, 1500)
	for i := range long {
		long[i] = 'a'
	}
	s := string(long)
	doc := client.FederalRegisterDocument{Excerpts: &s}
	got := derivePlaceholderSummary(doc)
	if len(got) != 1000 {
		t.Fatalf("expected 1000 chars, got %d", len(got))
	}
}

func TestCanonicalize_UnmarshalCompatibility(t *testing.T) {
	// Guardrail: the raw stored JSON created by client.Scrape (marshaled FederalRegisterDocument)
	// must still unmarshal into FederalRegisterDocument for canonicalization.
	in := client.FederalRegisterDocument{
		DocumentNumber:  "2025-01234",
		Title:           "A Title",
		Type:            "Notice",
		HTMLURL:         "https://example.com",
		PublicationDate: "2025-01-10",
	}
	b, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var out client.FederalRegisterDocument
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out.DocumentNumber != in.DocumentNumber || out.Title != in.Title || out.PublicationDate != in.PublicationDate {
		t.Fatalf("unexpected roundtrip: %#v", out)
	}
}

func TestNeedsEnrichment(t *testing.T) {
	impact := "medium"
	pol := 0

	tests := []struct {
		name string
		doc  func() *domain.PolicyDocument
		want bool
	}{
		{
			name: "missing impact_score",
			doc: func() *domain.PolicyDocument {
				return &domain.PolicyDocument{PoliticalScore: &pol, Keypoints: []string{"k"}}
			},
			want: true,
		},
		{
			name: "missing political_score",
			doc: func() *domain.PolicyDocument {
				return &domain.PolicyDocument{ImpactScore: &impact, Keypoints: []string{"k"}}
			},
			want: true,
		},
		{
			name: "empty keypoints",
			doc: func() *domain.PolicyDocument {
				return &domain.PolicyDocument{ImpactScore: &impact, PoliticalScore: &pol, Keypoints: []string{}}
			},
			want: true,
		},
		{
			name: "fully enriched fields present",
			doc: func() *domain.PolicyDocument {
				return &domain.PolicyDocument{ImpactScore: &impact, PoliticalScore: &pol, Keypoints: []string{"k"}}
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := needsEnrichment(tc.doc()); got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}
