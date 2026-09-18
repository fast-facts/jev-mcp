package mcp

import (
	"context"
	"fmt"

	"github.com/fast-facts/jev-mcp/internal/jev"
	"github.com/fast-facts/jev-mcp/internal/policy"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const findDesc = "Rank candidates against a plain-language query with TypeSafe Jev — no embeddings needed. " +
	"One Choice scores every candidate id by how well it answers the query, plus a Noul checks whether " +
	"any candidate addresses the query at all (so a confident top hit cannot masquerade as an answer). " +
	"Pattern: docs.typesafe.ai/cookbooks/semantic_find."

type FindInput struct {
	Query      string      `json:"query" jsonschema:"What you are looking for, in natural language"`
	Candidates []Candidate `json:"candidates" jsonschema:"Candidates to search. Up to 250; texts are truncated at 2000 chars"`
	TopK       *int        `json:"top_k,omitempty" jsonschema:"How many ranked candidates to return. Default 5"`
}

type FindHit struct {
	ID          string  `json:"id"`
	Probability float64 `json:"probability"`
	Text        string  `json:"text"`
}

type FindOutput struct {
	Tool          string        `json:"tool"`
	Model         string        `json:"model"`
	Query         string        `json:"query"`
	Exists        float64       `json:"exists"`
	ExistsVerdict policy.Exists `json:"exists_verdict"`
	Top           []FindHit     `json:"top"`
	Usage         jev.Usage     `json:"usage"`
}

func (h *handlers) find(ctx context.Context, _ *mcp.CallToolRequest, in FindInput) (*mcp.CallToolResult, FindOutput, error) {
	if in.Query == "" {
		return nil, FindOutput{}, fmt.Errorf("query is required")
	}
	items, err := bindCandidates(in.Candidates)
	if err != nil {
		return nil, FindOutput{}, err
	}
	topK := policy.DefaultFindTopK
	if in.TopK != nil {
		topK = *in.TopK
	}
	if topK < 1 || topK > policy.MaxFindTopK {
		return nil, FindOutput{}, fmt.Errorf("top_k must be between 1 and %d", policy.MaxFindTopK)
	}
	criteria := make(map[string]any, len(items))
	for _, c := range items {
		criteria[c.ID] = nil
	}
	questions := map[string]jev.Question{
		"best": jev.Choice(`Which candidate contains the best answer to: "`+in.Query+`"?`, criteria),
		"exists": jev.Noul(
			`Does any candidate address or answer: "`+in.Query+`"?`,
			"At least one candidate states or directly implies the answer",
			"No candidate addresses this",
		),
	}
	result, err := h.client.Evaluate(ctx, map[string]any{"query": in.Query, "candidates": items}, questions)
	if err != nil {
		return nil, FindOutput{}, err
	}
	probs := result.Answers["best"].Probabilities
	if probs == nil {
		probs = map[string]float64{}
	}
	ranked := policy.Rank(items, probs)
	if topK > len(ranked) {
		topK = len(ranked)
	}
	top := make([]FindHit, topK)
	for i, r := range ranked[:topK] {
		top[i] = FindHit{ID: r.ID, Probability: r.Probability, Text: r.Text}
	}
	exists := result.Answers["exists"].Noul
	return nil, FindOutput{
		Tool:          "jev_find",
		Model:         result.Model,
		Query:         in.Query,
		Exists:        exists,
		ExistsVerdict: policy.ExistsVerdict(exists),
		Top:           top,
		Usage:         result.Usage,
	}, nil
}
