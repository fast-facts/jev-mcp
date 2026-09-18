package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/fast-facts/jev-mcp/internal/jev"
	"github.com/fast-facts/jev-mcp/internal/policy"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const classifyDesc = "Assign each item to one class from a shared catalog with TypeSafe Jev, in one batched request: " +
	"the class catalog is sent once and every item becomes an independent Choice question. " +
	"Returns per item: the chosen class, the full distribution, confidence, winner-to-runner-up margin, " +
	"and an auto-versus-review decision. Auto requires both a high top probability (default 0.85) and a " +
	"clear margin (default 0.50). Include a manual_review class if you want an explicit escape hatch."

type ClassifyItem struct {
	ID   string `json:"id,omitempty" jsonschema:"Short identifier for this item"`
	Text string `json:"text" jsonschema:"The item text"`
}

type ClassifyClass struct {
	ID          string `json:"id,omitempty" jsonschema:"Short identifier for this class"`
	Description string `json:"description" jsonschema:"What belongs in this class"`
}

type ClassifyInput struct {
	Items         []ClassifyItem  `json:"items" jsonschema:"Items to classify. Text is truncated at 2000 characters"`
	Classes       []ClassifyClass `json:"classes" jsonschema:"Shared class catalog"`
	Purpose       string          `json:"purpose,omitempty" jsonschema:"What this classification is for"`
	Context       json.RawMessage `json:"context,omitempty" jsonschema:"Shared context for every item"`
	AutoAccept    *float64        `json:"auto_accept,omitempty" jsonschema:"Minimum top probability for auto. Default 0.85"`
	MinimumMargin *float64        `json:"minimum_margin,omitempty" jsonschema:"Minimum winner-to-runner-up gap for auto. Default 0.5"`
}

type labeled struct {
	external string
	key      string
	text     string
}

type ClassifyRow struct {
	ID             string             `json:"id"`
	Status         string             `json:"status,omitempty"`
	Classification *string            `json:"classification,omitempty"`
	Probabilities  map[string]float64 `json:"probabilities,omitempty"`
	Confidence     *float64           `json:"confidence,omitempty"`
	Margin         *float64           `json:"margin,omitempty"`
	TopProbability *float64           `json:"top_probability,omitempty"`
	Decision       policy.Decision    `json:"decision"`
}

type ClassifySummary struct {
	Items           int            `json:"items"`
	Auto            int            `json:"auto"`
	Review          int            `json:"review"`
	InvalidResponse int            `json:"invalid_response"`
	ByClass         map[string]int `json:"by_class"`
}

type ClassifyOutput struct {
	Tool       string             `json:"tool"`
	Model      string             `json:"model"`
	Summary    ClassifySummary    `json:"summary"`
	Thresholds map[string]float64 `json:"thresholds"`
	Results    []ClassifyRow      `json:"results"`
	Usage      jev.Usage          `json:"usage"`
}

func (h *handlers) classify(ctx context.Context, _ *mcp.CallToolRequest, in ClassifyInput) (*mcp.CallToolResult, ClassifyOutput, error) {
	items, classes, err := bindClassify(in)
	if err != nil {
		return nil, ClassifyOutput{}, err
	}
	autoAccept := policy.DefaultClassifyAccept
	if in.AutoAccept != nil {
		autoAccept = *in.AutoAccept
	}
	minMargin := policy.DefaultClassifyMargin
	if in.MinimumMargin != nil {
		minMargin = *in.MinimumMargin
	}
	purpose := in.Purpose
	if purpose == "" {
		purpose = "Assign each item to exactly one class."
	}
	stateClasses := make([]map[string]string, len(classes))
	criteria := make(map[string]any, len(classes))
	for i, c := range classes {
		stateClasses[i] = map[string]string{"id": c.key, "description": c.text}
		criteria[c.key] = nil
	}
	questions := make(map[string]jev.Question, len(items))
	for _, item := range items {
		questions[item.key] = jev.Choice(
			map[string]any{"task": "Which class does this item belong to?", "item": map[string]string{"id": item.key, "text": item.text}},
			criteria,
		)
	}
	var contextVal any
	if len(in.Context) > 0 {
		if err := json.Unmarshal(in.Context, &contextVal); err != nil {
			return nil, ClassifyOutput{}, fmt.Errorf("context: %w", err)
		}
	}
	result, err := h.client.Evaluate(ctx, map[string]any{"purpose": purpose, "context": contextVal, "classes": stateClasses}, questions)
	if err != nil {
		return nil, ClassifyOutput{}, err
	}
	return nil, mapClassify(result, items, classes, policy.ClassifyGates{AutoAccept: autoAccept, MinimumMargin: minMargin}), nil
}

func bindClassify(in ClassifyInput) ([]labeled, []labeled, error) {
	if len(in.Items) == 0 || len(in.Items) > policy.MaxItems {
		return nil, nil, fmt.Errorf("items must have between 1 and %d entries", policy.MaxItems)
	}
	if len(in.Classes) < 2 || len(in.Classes) > policy.MaxClasses {
		return nil, nil, fmt.Errorf("classes must have between 2 and %d entries", policy.MaxClasses)
	}
	if len(in.Items)*len(in.Classes) > policy.MaxItemClassPairs {
		return nil, nil, fmt.Errorf("batch too large: %d items x %d classes exceeds %d", len(in.Items), len(in.Classes), policy.MaxItemClassPairs)
	}
	items, err := labelList(in.Items, "item", func(it ClassifyItem) (string, string) { return it.ID, it.Text })
	if err != nil {
		return nil, nil, err
	}
	classes, err := labelList(in.Classes, "class", func(c ClassifyClass) (string, string) { return c.ID, c.Description })
	if err != nil {
		return nil, nil, err
	}
	return items, classes, nil
}

func labelList[T any](in []T, kind string, pick func(T) (string, string)) ([]labeled, error) {
	seen := make(map[string]struct{}, len(in))
	out := make([]labeled, len(in))
	for i, raw := range in {
		id, text := pick(raw)
		if id != "" {
			if _, ok := seen[id]; ok {
				return nil, fmt.Errorf("duplicate %s id: %s", kind, id)
			}
			seen[id] = struct{}{}
		} else {
			id = kind + strconv.Itoa(i)
		}
		out[i] = labeled{external: id, key: string(kind[0]) + strconv.Itoa(i), text: policy.Truncate(text, policy.MaxItemChars)}
	}
	return out, nil
}

func mapClassify(result jev.Result, items, classes []labeled, gates policy.ClassifyGates) ClassifyOutput {
	expected := make(map[string]struct{}, len(classes))
	keyToExt := make(map[string]string, len(classes))
	for _, c := range classes {
		expected[c.key] = struct{}{}
		keyToExt[c.key] = c.external
	}
	out := ClassifyOutput{
		Tool:       "jev_classify",
		Model:      result.Model,
		Thresholds: map[string]float64{"auto_accept": gates.AutoAccept, "minimum_margin": gates.MinimumMargin},
		Results:    make([]ClassifyRow, len(items)),
		Usage:      result.Usage,
		Summary:    ClassifySummary{Items: len(items), ByClass: map[string]int{}},
	}
	for i, item := range items {
		answer := result.Answers[item.key]
		if !policy.ValidChoice(answer.Choice, answer.Probabilities, expected) {
			out.Results[i] = ClassifyRow{ID: item.external, Status: "invalid_response", Decision: policy.DecisionReview}
			out.Summary.InvalidResponse++
			continue
		}
		extProbs := make(map[string]float64, len(classes))
		for _, c := range classes {
			extProbs[c.external] = answer.Probabilities[c.key]
		}
		margin := policy.MarginOf(answer.Probabilities)
		top := answer.Probabilities[answer.Choice]
		className := keyToExt[answer.Choice]
		conf := answer.Confidence
		decision := policy.ClassifyDecision(top, margin, gates)
		out.Results[i] = ClassifyRow{
			ID:             item.external,
			Classification: &className,
			Probabilities:  extProbs,
			Confidence:     &conf,
			Margin:         &margin,
			TopProbability: &top,
			Decision:       decision,
		}
		out.Summary.ByClass[className]++
		if decision == policy.DecisionAuto {
			out.Summary.Auto++
		} else {
			out.Summary.Review++
		}
	}
	return out
}
