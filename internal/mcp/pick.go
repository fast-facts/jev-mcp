package mcp

import (
	"context"
	"fmt"

	"github.com/fast-facts/jev-mcp/internal/jev"
	"github.com/fast-facts/jev-mcp/internal/policy"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const pickDesc = "Pick at most one candidate for a query, or none. Unlike jev_find, this does not return a ranked list: " +
	"Choice includes an explicit none option, and a Noul checks that some candidate actually answers the query. " +
	"Use for which file, skill, or error — or skip."

type PickInput struct {
	Query      string      `json:"query" jsonschema:"What you need one candidate for"`
	Candidates []Candidate `json:"candidates" jsonschema:"Candidates to choose from. Up to 250; texts are truncated at 2000 chars"`
}

type PickOutput struct {
	Tool          string        `json:"tool"`
	Model         string        `json:"model"`
	Query         string        `json:"query"`
	Selected      *string       `json:"selected,omitempty"`
	Exists        float64       `json:"exists"`
	ExistsVerdict policy.Exists `json:"exists_verdict"`
	Usage         jev.Usage     `json:"usage"`
}

func (h *handlers) pick(ctx context.Context, _ *mcp.CallToolRequest, in PickInput) (*mcp.CallToolResult, PickOutput, error) {
	if in.Query == "" {
		return nil, PickOutput{}, fmt.Errorf("query is required")
	}
	items, err := bindCandidates(in.Candidates)
	if err != nil {
		return nil, PickOutput{}, err
	}
	questions := map[string]jev.Question{
		"best": jev.Choice(
			`Which candidate is the genuine answer to: "`+in.Query+`"? Choose none if none is.`,
			noneCriteria(items, "No candidate genuinely answers the query"),
		),
		"exists": jev.Noul(
			`Does any candidate genuinely answer: "`+in.Query+`"?`,
			"At least one candidate states or directly implies the answer",
			"No candidate answers this; the closest match is not enough",
		),
	}
	result, err := h.client.Evaluate(ctx, map[string]any{"query": in.Query, "candidates": items}, questions)
	if err != nil {
		return nil, PickOutput{}, err
	}
	exists := result.Answers["exists"].Noul
	out := PickOutput{
		Tool:          "jev_pick",
		Model:         result.Model,
		Query:         in.Query,
		Exists:        exists,
		ExistsVerdict: policy.ExistsVerdict(exists),
		Usage:         result.Usage,
	}
	choice := result.Answers["best"].Choice
	if policy.AcceptPick(choice, exists, idSet(items)) {
		out.Selected = &choice
	}
	return nil, out, nil
}
