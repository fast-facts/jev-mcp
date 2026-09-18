package mcp

import (
	"context"
	"testing"

	"github.com/fast-facts/jev-mcp/internal/jev"
)

func Test_extract_returns_verbatim_span(t *testing.T) {
	h := newTestHandlers(t, map[string]jev.Answer{
		"which":  {Type: "choice", Choice: "e1"},
		"exists": {Type: "noul", Noul: 0.99},
	})
	_, out, err := h.extract(context.Background(), nil, ExtractInput{
		Query: "the support email",
		Candidates: []Candidate{
			{ID: "e0", Text: "not-an-email"},
			{ID: "e1", Text: "support@example.com"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Span == nil || *out.Span != "support@example.com" {
		t.Fatalf("span %+v", out.Span)
	}
	if out.ID == nil || *out.ID != "e1" {
		t.Fatalf("id %+v", out.ID)
	}
}

func Test_extract_does_not_invent_a_span(t *testing.T) {
	h := newTestHandlers(t, map[string]jev.Answer{
		"which":  {Type: "choice", Choice: "invented"},
		"exists": {Type: "noul", Noul: 0.99},
	})
	_, out, err := h.extract(context.Background(), nil, ExtractInput{
		Query:      "the support email",
		Candidates: []Candidate{{ID: "e1", Text: "support@example.com"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Span != nil {
		t.Fatalf("invented span %s", *out.Span)
	}
}

func Test_extract_returns_null_when_exists_is_low(t *testing.T) {
	h := newTestHandlers(t, map[string]jev.Answer{
		"which":  {Type: "choice", Choice: "e1"},
		"exists": {Type: "noul", Noul: 0.1},
	})
	_, out, err := h.extract(context.Background(), nil, ExtractInput{
		Query:      "the support email",
		Candidates: []Candidate{{ID: "e1", Text: "support@example.com"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Span != nil {
		t.Fatalf("span %s", *out.Span)
	}
}
