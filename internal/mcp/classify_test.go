package mcp

import (
	"context"
	"strings"
	"testing"

	"github.com/fast-facts/jev-mcp/internal/jev"
	"github.com/fast-facts/jev-mcp/internal/policy"
)

func Test_classify_maps_opaque_keys_back_to_caller_ids(t *testing.T) {
	h := newTestHandlers(t, map[string]jev.Answer{
		"i0": {
			Type:          "choice",
			Choice:        "c0",
			Confidence:    1,
			Probabilities: map[string]float64{"c0": 0.95, "c1": 0.05},
		},
	})
	_, out, err := h.classify(context.Background(), nil, ClassifyInput{
		Items:   []ClassifyItem{{ID: "m1", Text: "charged twice"}},
		Classes: []ClassifyClass{{ID: "billing", Description: "payments"}, {ID: "sales", Description: "pricing"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Results[0].Classification == nil || *out.Results[0].Classification != "billing" {
		t.Fatalf("class %+v", out.Results[0].Classification)
	}
	if out.Results[0].Decision != policy.DecisionAuto {
		t.Fatalf("decision %s", out.Results[0].Decision)
	}
}

func Test_classify_rejects_duplicate_ids(t *testing.T) {
	h := &handlers{}
	_, _, err := h.classify(context.Background(), nil, ClassifyInput{
		Items:   []ClassifyItem{{ID: "m1", Text: "a"}, {ID: "m1", Text: "b"}},
		Classes: []ClassifyClass{{ID: "a", Description: "a"}, {ID: "b", Description: "b"}},
	})
	if err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("err=%v", err)
	}
}

func Test_classify_marks_invalid_response(t *testing.T) {
	h := newTestHandlers(t, map[string]jev.Answer{
		"i0": {Type: "choice", Choice: "not-a-class", Probabilities: map[string]float64{"x": 1}},
	})
	_, out, err := h.classify(context.Background(), nil, ClassifyInput{
		Items:   []ClassifyItem{{ID: "m1", Text: "hello"}},
		Classes: []ClassifyClass{{ID: "billing", Description: "pay"}, {ID: "sales", Description: "price"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Results[0].Status != "invalid_response" {
		t.Fatalf("status %s", out.Results[0].Status)
	}
	if out.Summary.InvalidResponse != 1 || out.Summary.Auto != 0 {
		t.Fatalf("summary %+v", out.Summary)
	}
}
