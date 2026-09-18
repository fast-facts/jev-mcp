package mcp

import (
	"context"
	"fmt"

	"github.com/fast-facts/jev-mcp/internal/jev"
	"github.com/fast-facts/jev-mcp/internal/policy"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const extractDesc = "Pick one exact span from a shortlist you already have (a quote, an id, an email, a config key). " +
	"The returned span is copied from your candidates, never invented. Returns null when none of the spans is the value."

type ExtractInput struct {
	Query      string      `json:"query" jsonschema:"What value to extract, e.g. the support email or the error id"`
	Candidates []Candidate `json:"candidates" jsonschema:"Verbatim spans to choose from. Up to 250; texts are truncated at 2000 chars"`
}

type ExtractOutput struct {
	Tool          string        `json:"tool"`
	Model         string        `json:"model"`
	Query         string        `json:"query"`
	ID            *string       `json:"id,omitempty"`
	Span          *string       `json:"span,omitempty"`
	Exists        float64       `json:"exists"`
	ExistsVerdict policy.Exists `json:"exists_verdict"`
	Usage         jev.Usage     `json:"usage"`
}

func (h *handlers) extract(ctx context.Context, _ *mcp.CallToolRequest, in ExtractInput) (*mcp.CallToolResult, ExtractOutput, error) {
	if in.Query == "" {
		return nil, ExtractOutput{}, fmt.Errorf("query is required")
	}
	items, err := bindCandidates(in.Candidates)
	if err != nil {
		return nil, ExtractOutput{}, err
	}
	questions := map[string]jev.Question{
		"which": jev.Choice(
			`Which span is exactly the value requested: "`+in.Query+`"? Choose none if none is. Do not invent text.`,
			noneCriteria(items, "None of these spans is the requested value"),
		),
		"exists": jev.Noul(
			`Is the requested value present as one of these spans: "`+in.Query+`"?`,
			"One span is the requested value, verbatim",
			"No span is the requested value",
		),
	}
	result, err := h.client.Evaluate(ctx, map[string]any{"query": in.Query, "candidates": items}, questions)
	if err != nil {
		return nil, ExtractOutput{}, err
	}
	exists := result.Answers["exists"].Noul
	out := ExtractOutput{
		Tool:          "jev_extract",
		Model:         result.Model,
		Query:         in.Query,
		Exists:        exists,
		ExistsVerdict: policy.ExistsVerdict(exists),
		Usage:         result.Usage,
	}
	choice := result.Answers["which"].Choice
	if policy.AcceptPick(choice, exists, idSet(items)) {
		span := textByID(items)[choice]
		out.ID = &choice
		out.Span = &span
	}
	return nil, out, nil
}
