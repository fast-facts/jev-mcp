package mcp

import (
	"context"
	"fmt"

	"github.com/fast-facts/jev-mcp/internal/jev"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const askDesc = "Escape hatch: send shared state plus named noul/choice/score questions to TypeSafe Jev. " +
	"Use when no other jev_* recipe fits. Jev does not write code or prose. Questions in one call run in parallel. " +
	"Returns typed answers, probabilities, confidence, and usage."

type AskInput struct {
	State     any                    `json:"state" jsonschema:"Shared state: a string, object, or array"`
	Questions map[string]AskQuestion `json:"questions" jsonschema:"Named noul, choice, or score questions"`
}

type AskQuestion struct {
	Type         string `json:"type" jsonschema:"noul, choice, or score"`
	Instructions any    `json:"instructions,omitempty" jsonschema:"Question text or structured instructions"`
	Criteria     any    `json:"criteria,omitempty" jsonschema:"Choice map, score levels, or noul true/false"`
}

type AskOutput struct {
	Tool    string                `json:"tool"`
	Model   string                `json:"model"`
	Answers map[string]jev.Answer `json:"answers"`
	Usage   jev.Usage             `json:"usage"`
}

func (h *handlers) ask(ctx context.Context, _ *mcp.CallToolRequest, in AskInput) (*mcp.CallToolResult, AskOutput, error) {
	if err := jev.ValidState(in.State); err != nil {
		return nil, AskOutput{}, err
	}
	if len(in.Questions) == 0 {
		return nil, AskOutput{}, fmt.Errorf("questions is required")
	}
	questions := make(map[string]jev.Question, len(in.Questions))
	for id, q := range in.Questions {
		jq := jev.Question{Type: q.Type, Instructions: q.Instructions, Criteria: q.Criteria}
		if err := jq.Validate(); err != nil {
			return nil, AskOutput{}, fmt.Errorf("question %s: %w", id, err)
		}
		questions[id] = jq
	}
	result, err := h.client.Evaluate(ctx, in.State, questions)
	if err != nil {
		return nil, AskOutput{}, err
	}
	return nil, AskOutput{Tool: "jev_ask", Model: result.Model, Answers: result.Answers, Usage: result.Usage}, nil
}
