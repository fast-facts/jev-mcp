package mcp

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/fast-facts/jev-mcp/internal/jev"
	"github.com/fast-facts/jev-mcp/internal/policy"
)

func Test_verify_maps_choice_to_verdict(t *testing.T) {
	h := newTestHandlers(t, map[string]jev.Answer{
		"relation_claim0": {
			Type:          "choice",
			Choice:        "contradicts",
			Confidence:    1,
			Probabilities: map[string]float64{"supports": 0, "contradicts": 1, "says_nothing": 0},
		},
	})
	_, out, err := h.verify(context.Background(), nil, VerifyInput{
		Claims:   []string{"Helmets are optional."},
		Evidence: json.RawMessage(`"Every rider must wear a helmet."`),
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Results[0].Verdict != "contradicted" {
		t.Fatalf("verdict %s", out.Results[0].Verdict)
	}
	if out.Results[0].Action != policy.DecisionAuto {
		t.Fatalf("action %s", out.Results[0].Action)
	}
	if out.Summary.Contradicted != 1 {
		t.Fatalf("summary %+v", out.Summary)
	}
}

func Test_verify_accepts_object_and_array_evidence(t *testing.T) {
	h := newTestHandlers(t, map[string]jev.Answer{
		"relation_claim0": {
			Type:          "choice",
			Choice:        "supports",
			Confidence:    0.9,
			Probabilities: map[string]float64{"supports": 0.9, "contradicts": 0.05, "says_nothing": 0.05},
		},
		"source_claim0": {Type: "choice", Choice: "ord"},
	})
	_, out, err := h.verify(context.Background(), nil, VerifyInput{
		Claims:   []string{"Riders need helmets."},
		Evidence: json.RawMessage(`[{"id":"ord","text":"Every rider must wear a helmet."},{"id":"other","text":"Parking rules."}]`),
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Results[0].Verdict != "verified" {
		t.Fatalf("verdict %s", out.Results[0].Verdict)
	}
	if out.Results[0].SupportingEvidence == nil || *out.Results[0].SupportingEvidence != "ord" {
		t.Fatalf("source %+v", out.Results[0].SupportingEvidence)
	}
}

func Test_verify_rejects_empty_claims(t *testing.T) {
	h := &handlers{}
	_, _, err := h.verify(context.Background(), nil, VerifyInput{
		Claims:   nil,
		Evidence: json.RawMessage(`"x"`),
	})
	if err == nil {
		t.Fatal("expected error")
	}
}
