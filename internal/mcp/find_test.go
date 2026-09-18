package mcp

import (
	"context"
	"testing"

	"github.com/fast-facts/jev-mcp/internal/jev"
	"github.com/fast-facts/jev-mcp/internal/policy"
)

func Test_find_ranks_and_sets_exists_verdict(t *testing.T) {
	h := newTestHandlers(t, map[string]jev.Answer{
		"best": {
			Type:          "choice",
			Choice:        "auth",
			Probabilities: map[string]float64{"auth": 0.99, "billing": 0.01},
		},
		"exists": {Type: "noul", Noul: 0.99},
	})
	_, out, err := h.find(context.Background(), nil, FindInput{
		Query: "rotate keys",
		Candidates: []Candidate{
			{ID: "billing", Text: "invoices"},
			{ID: "auth", Text: "rotate an API key"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.ExistsVerdict != policy.ExistsAnsweredVerdict {
		t.Fatalf("exists %s", out.ExistsVerdict)
	}
	if out.Top[0].ID != "auth" {
		t.Fatalf("top %v", out.Top)
	}
}

func Test_find_rejects_empty_query(t *testing.T) {
	h := &handlers{}
	_, _, err := h.find(context.Background(), nil, FindInput{
		Query:      "",
		Candidates: []Candidate{{ID: "a", Text: "x"}},
	})
	if err == nil {
		t.Fatal("expected error")
	}
}
