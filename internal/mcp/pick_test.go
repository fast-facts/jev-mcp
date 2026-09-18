package mcp

import (
	"context"
	"testing"

	"github.com/fast-facts/jev-mcp/internal/jev"
	"github.com/fast-facts/jev-mcp/internal/policy"
)

func Test_pick_selects_when_exists_is_high(t *testing.T) {
	h := newTestHandlers(t, map[string]jev.Answer{
		"best":   {Type: "choice", Choice: "auth"},
		"exists": {Type: "noul", Noul: 0.99},
	})
	_, out, err := h.pick(context.Background(), nil, PickInput{
		Query: "rotate keys",
		Candidates: []Candidate{
			{ID: "billing", Text: "invoices"},
			{ID: "auth", Text: "rotate an API key"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Selected == nil || *out.Selected != "auth" {
		t.Fatalf("selected %+v", out.Selected)
	}
}

func Test_pick_returns_null_when_exists_is_low(t *testing.T) {
	h := newTestHandlers(t, map[string]jev.Answer{
		"best":   {Type: "choice", Choice: "auth"},
		"exists": {Type: "noul", Noul: 0.14},
	})
	_, out, err := h.pick(context.Background(), nil, PickInput{
		Query:      "arbitration",
		Candidates: []Candidate{{ID: "auth", Text: "rotate an API key"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Selected != nil {
		t.Fatalf("selected %s", *out.Selected)
	}
	if out.ExistsVerdict != policy.ExistsAbsentVerdict {
		t.Fatalf("exists %s", out.ExistsVerdict)
	}
}

func Test_pick_returns_null_when_model_chooses_none(t *testing.T) {
	h := newTestHandlers(t, map[string]jev.Answer{
		"best":   {Type: "choice", Choice: policy.NoneOption},
		"exists": {Type: "noul", Noul: 0.99},
	})
	_, out, err := h.pick(context.Background(), nil, PickInput{
		Query:      "arbitration",
		Candidates: []Candidate{{ID: "auth", Text: "rotate an API key"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Selected != nil {
		t.Fatalf("selected %s", *out.Selected)
	}
}
