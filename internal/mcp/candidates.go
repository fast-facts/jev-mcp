package mcp

import (
	"fmt"

	"github.com/fast-facts/jev-mcp/internal/policy"
)

type Candidate struct {
	ID   string `json:"id,omitempty" jsonschema:"Short identifier for this candidate"`
	Text string `json:"text" jsonschema:"The candidate text"`
}

func bindCandidates(raw []Candidate) ([]policy.Item, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("candidates is required")
	}
	if len(raw) > policy.MaxCandidates {
		return nil, fmt.Errorf("at most %d candidates", policy.MaxCandidates)
	}
	items := make([]policy.Item, len(raw))
	for i, c := range raw {
		items[i] = policy.Item{ID: c.ID, Text: policy.Truncate(c.Text, policy.MaxCandidateChars)}
	}
	items = policy.UniqueIDs(items, "candidate")
	for _, it := range items {
		if it.ID == policy.NoneOption {
			return nil, fmt.Errorf("candidate id %q is reserved", policy.NoneOption)
		}
	}
	return items, nil
}

func noneCriteria(items []policy.Item, noneDesc string) map[string]any {
	criteria := make(map[string]any, len(items)+1)
	for _, c := range items {
		criteria[c.ID] = nil
	}
	criteria[policy.NoneOption] = noneDesc
	return criteria
}

func idSet(items []policy.Item) map[string]struct{} {
	ids := make(map[string]struct{}, len(items))
	for _, it := range items {
		ids[it.ID] = struct{}{}
	}
	return ids
}

func textByID(items []policy.Item) map[string]string {
	out := make(map[string]string, len(items))
	for _, it := range items {
		out[it.ID] = it.Text
	}
	return out
}
