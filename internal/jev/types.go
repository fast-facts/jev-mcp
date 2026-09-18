// Package jev calls POST /v1/systemone on the TypeSafe API.
package jev

import (
	"errors"
	"fmt"
)

const (
	PathSystemOne    = "/v1/systemone"
	MaxChoiceOptions = 255
	MinScoreLevels   = 2
)

type Question struct {
	Type         string `json:"type"`
	Instructions any    `json:"instructions"`
	Criteria     any    `json:"criteria,omitempty"`
}

type requestBody struct {
	Model     string              `json:"model"`
	State     any                 `json:"state"`
	Questions map[string]Question `json:"questions"`
}

type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type Answer struct {
	Type          string             `json:"type"`
	Noul          float64            `json:"noul,omitempty"`
	Choice        string             `json:"choice,omitempty"`
	Score         float64            `json:"score,omitempty"`
	Confidence    float64            `json:"confidence,omitempty"`
	Probabilities map[string]float64 `json:"probabilities,omitempty"`
	Legend        map[string]string  `json:"legend,omitempty"`
}

type Result struct {
	Model   string
	Answers map[string]Answer
	Usage   Usage
}

func Noul(instructions any, yes, no string) Question {
	return Question{
		Type:         "noul",
		Instructions: instructions,
		Criteria:     map[string]string{"true": yes, "false": no},
	}
}

func Choice(instructions any, criteria any) Question {
	return Question{Type: "choice", Instructions: instructions, Criteria: criteria}
}

func Score(instructions any, levels []string) Question {
	return Question{Type: "score", Instructions: instructions, Criteria: levels}
}

func ValidState(state any) error {
	if state == nil {
		return errors.New("state is required")
	}
	switch state.(type) {
	case string, map[string]any, []any:
		return nil
	default:
		return errors.New("state must be a string, object, or array")
	}
}

func (q Question) Validate() error {
	if q.Instructions == nil {
		return errors.New("instructions is required")
	}
	if s, ok := q.Instructions.(string); ok && s == "" {
		return errors.New("instructions is required")
	}
	switch q.Type {
	case "noul":
		return nil
	case "choice":
		n := mapLen(q.Criteria)
		if n < 1 {
			return errors.New("choice criteria is required")
		}
		if n > MaxChoiceOptions {
			return fmt.Errorf("choice criteria has at most %d options", MaxChoiceOptions)
		}
		return nil
	case "score":
		n := sliceLen(q.Criteria)
		if n < MinScoreLevels {
			return fmt.Errorf("score criteria needs at least %d levels", MinScoreLevels)
		}
		return nil
	default:
		return errors.New("type must be noul, choice, or score")
	}
}

func mapLen(v any) int {
	switch m := v.(type) {
	case map[string]any:
		return len(m)
	case map[string]string:
		return len(m)
	default:
		return 0
	}
}

func sliceLen(v any) int {
	switch s := v.(type) {
	case []any:
		return len(s)
	case []string:
		return len(s)
	default:
		return 0
	}
}
