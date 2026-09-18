package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/fast-facts/jev-mcp/internal/jev"
	"github.com/fast-facts/jev-mcp/internal/policy"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const verifyDesc = "Check each claim against provided evidence text with TypeSafe Jev. Returns per claim: " +
	"verdict (verified | contradicted | unsupported), full probability distribution, confidence, " +
	"and whether the verdict stands on its own (auto) or needs human review. " +
	"Pattern: docs.typesafe.ai/cookbooks/citation_check."

type VerifyInput struct {
	Claims     []string `json:"claims" jsonschema:"Claims to verify"`
	Evidence   any      `json:"evidence" jsonschema:"Evidence as a string, {id,text}, or an array of those"`
	AutoAccept *float64 `json:"auto_accept,omitempty" jsonschema:"Confidence at or above which a verdict is auto. Default 0.8"`
}

type VerifyClaim struct {
	ID                 string             `json:"id"`
	Claim              string             `json:"claim"`
	Verdict            string             `json:"verdict"`
	Probabilities      map[string]float64 `json:"probabilities,omitempty"`
	Confidence         *float64           `json:"confidence,omitempty"`
	Action             policy.Decision    `json:"action"`
	SupportingEvidence *string            `json:"supporting_evidence,omitempty"`
}

type VerifySummary struct {
	Verified     int `json:"verified"`
	Contradicted int `json:"contradicted"`
	Unsupported  int `json:"unsupported"`
	NeedsReview  int `json:"needs_review"`
}

type VerifyOutput struct {
	Tool       string        `json:"tool"`
	Model      string        `json:"model"`
	AutoAccept float64       `json:"auto_accept"`
	Summary    VerifySummary `json:"summary"`
	Results    []VerifyClaim `json:"results"`
	Usage      jev.Usage     `json:"usage"`
}

func (h *handlers) verify(ctx context.Context, _ *mcp.CallToolRequest, in VerifyInput) (*mcp.CallToolResult, VerifyOutput, error) {
	if len(in.Claims) == 0 {
		return nil, VerifyOutput{}, fmt.Errorf("claims is required")
	}
	evidence, err := parseEvidence(in.Evidence)
	if err != nil {
		return nil, VerifyOutput{}, err
	}
	autoAccept := policy.DefaultVerifyAutoAccept
	if in.AutoAccept != nil {
		autoAccept = *in.AutoAccept
	}
	claims := make([]policy.Item, len(in.Claims))
	for i, text := range in.Claims {
		claims[i] = policy.Item{Text: text}
	}
	claims = policy.UniqueIDs(claims, "claim")
	evidence = policy.UniqueIDs(evidence, "evidence")

	questions := make(map[string]jev.Question, len(claims)*2)
	for _, claim := range claims {
		questions["relation_"+claim.ID] = jev.Choice(
			"How does the evidence relate to claim `"+claim.ID+"` ("+claim.Text+")?",
			map[string]string{
				"supports":     "The evidence states the claim or directly implies that it is true",
				"contradicts":  "The evidence states the opposite of the claim or implies that it is false",
				"says_nothing": "The evidence does not address what the claim asserts, either way",
			},
		)
		if len(evidence) > 1 {
			criteria := make(map[string]any, len(evidence)+1)
			for _, e := range evidence {
				criteria[e.ID] = nil
			}
			criteria[policy.NoneOption] = "No single evidence item contains the content the claim depends on"
			questions["source_"+claim.ID] = jev.Choice(
				"Which evidence item does claim `"+claim.ID+"` ("+claim.Text+") rest on?",
				criteria,
			)
		}
	}
	state := map[string]any{"purpose": "Verify each claim in claims against the evidence in evidence.", "claims": claims, "evidence": evidence}
	result, err := h.client.Evaluate(ctx, state, questions)
	if err != nil {
		return nil, VerifyOutput{}, err
	}
	return nil, mapVerify(result, claims, autoAccept), nil
}

func mapVerify(result jev.Result, claims []policy.Item, autoAccept float64) VerifyOutput {
	out := VerifyOutput{Tool: "jev_verify", Model: result.Model, AutoAccept: autoAccept, Results: make([]VerifyClaim, len(claims)), Usage: result.Usage}
	for i, claim := range claims {
		relation := result.Answers["relation_"+claim.ID]
		source := result.Answers["source_"+claim.ID]
		verdict := policy.RelationToVerdict[relation.Choice]
		if verdict == "" {
			verdict = "unknown"
		}
		row := VerifyClaim{
			ID:            claim.ID,
			Claim:         claim.Text,
			Verdict:       verdict,
			Probabilities: relation.Probabilities,
			Action:        policy.DecisionReview,
		}
		if relation.Type == "choice" {
			conf := relation.Confidence
			row.Confidence = &conf
			row.Action = policy.VerifyAction(conf, autoAccept)
		}
		if source.Choice != "" && source.Choice != policy.NoneOption {
			id := source.Choice
			row.SupportingEvidence = &id
		}
		out.Results[i] = row
		switch verdict {
		case "verified":
			out.Summary.Verified++
		case "contradicted":
			out.Summary.Contradicted++
		case "unsupported":
			out.Summary.Unsupported++
		}
		if row.Action == policy.DecisionReview {
			out.Summary.NeedsReview++
		}
	}
	return out
}

func parseEvidence(raw any) ([]policy.Item, error) {
	if raw == nil {
		return nil, fmt.Errorf("evidence is required")
	}
	switch v := raw.(type) {
	case string:
		if v == "" {
			return nil, fmt.Errorf("evidence is required")
		}
		return []policy.Item{{ID: "evidence", Text: v}}, nil
	case json.RawMessage:
		return parseEvidenceJSON(v)
	case []byte:
		return parseEvidenceJSON(v)
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("evidence must be a string, object, or array")
		}
		return parseEvidenceJSON(b)
	}
}

func parseEvidenceJSON(raw json.RawMessage) ([]policy.Item, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("evidence is required")
	}
	var asString string
	if err := json.Unmarshal(raw, &asString); err == nil {
		return []policy.Item{{ID: "evidence", Text: asString}}, nil
	}
	var one struct {
		ID   string `json:"id"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &one); err == nil && one.Text != "" {
		return []policy.Item{{ID: one.ID, Text: one.Text}}, nil
	}
	var many []struct {
		ID   string `json:"id"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &many); err != nil {
		return nil, fmt.Errorf("evidence must be a string, object, or array")
	}
	if len(many) == 0 {
		return nil, fmt.Errorf("evidence is required")
	}
	out := make([]policy.Item, len(many))
	for i, e := range many {
		if e.Text == "" {
			return nil, fmt.Errorf("evidence[%d].text is required", i)
		}
		out[i] = policy.Item{ID: e.ID, Text: e.Text}
	}
	return out, nil
}
