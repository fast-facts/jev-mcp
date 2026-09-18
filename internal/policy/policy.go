// Package policy maps Jev probabilities to verdicts and tool limits.
package policy

import (
	"regexp"
	"slices"
	"strconv"
)

const (
	MaxCandidates     = 250
	MaxCandidateChars = 2000
	MaxClasses        = 250
	MaxItems          = 64
	MaxItemChars      = 2000
	MaxItemClassPairs = 8000

	DefaultVerifyAutoAccept = 0.8
	DefaultScreenBlockAt    = 0.75
	DefaultScreenReviewAt   = 0.25
	DefaultClassifyAccept   = 0.85
	DefaultClassifyMargin   = 0.5
	DefaultFindTopK         = 5
	MaxFindTopK             = 50

	ExistsAnswered = 0.7
	ExistsAbsent   = 0.35
	SkipSubstance  = 0.3
	SkipRelevance  = 0.3
)

var (
	unsafeID = regexp.MustCompile(`[^A-Za-z0-9_.-]+`)
	edgeID   = regexp.MustCompile(`^_+|_+$`)
)

type Decision string

const (
	DecisionAuto   Decision = "auto"
	DecisionReview Decision = "review"
)

type ScreenAction string

const (
	ScreenPass   ScreenAction = "pass"
	ScreenReview ScreenAction = "review"
	ScreenBlock  ScreenAction = "block"
	ScreenSkip   ScreenAction = "skip"
)

type Exists string

const (
	ExistsAnsweredVerdict Exists = "answered"
	ExistsPartialVerdict  Exists = "partial"
	ExistsAbsentVerdict   Exists = "absent"
)

var RelationToVerdict = map[string]string{
	"supports":     "verified",
	"contradicts":  "contradicted",
	"says_nothing": "unsupported",
}

type Item struct {
	ID   string
	Text string
}

type Ranked struct {
	ID          string
	Text        string
	Probability float64
}

type ScreenInput struct {
	Injection float64
	Substance *float64
	Relevance *float64
	BlockAt   float64
	ReviewAt  float64
}

type ScreenResult struct {
	Action ScreenAction `json:"action"`
	Reason string       `json:"reason"`
}

type ClassifyGates struct {
	AutoAccept    float64
	MinimumMargin float64
}

func SanitizeID(id string) string {
	cleaned := edgeID.ReplaceAllString(unsafeID.ReplaceAllString(id, "_"), "")
	if len(cleaned) > 64 {
		return cleaned[:64]
	}
	return cleaned
}

func UniqueIDs(items []Item, prefix string) []Item {
	used := make(map[string]struct{}, len(items))
	out := make([]Item, len(items))
	for i, item := range items {
		base := SanitizeID(item.ID)
		if base == "" {
			base = prefix + strconv.Itoa(i)
		}
		id := base
		n := 1
		for {
			if _, ok := used[id]; !ok {
				break
			}
			id = base + "_" + strconv.Itoa(n)
			n++
		}
		used[id] = struct{}{}
		out[i] = Item{ID: id, Text: item.Text}
	}
	return out
}

func Truncate(text string, maxChars int) string {
	runes := []rune(text)
	if len(runes) <= maxChars {
		return text
	}
	return string(runes[:maxChars]) + " […truncated]"
}

func VerifyAction(confidence, autoAccept float64) Decision {
	if confidence >= autoAccept {
		return DecisionAuto
	}
	return DecisionReview
}

func Screen(in ScreenInput) ScreenResult {
	if in.Injection >= in.BlockAt {
		return ScreenResult{
			Action: ScreenBlock,
			Reason: "injection probability " + format2(in.Injection) + " >= block threshold " + format2(in.BlockAt),
		}
	}
	if in.Injection >= in.ReviewAt {
		return ScreenResult{
			Action: ScreenReview,
			Reason: "injection probability " + format2(in.Injection) + " >= review threshold " + format2(in.ReviewAt),
		}
	}
	if in.Substance != nil && *in.Substance < SkipSubstance {
		return ScreenResult{
			Action: ScreenSkip,
			Reason: "little substantive content (substance " + format2(*in.Substance) + ")",
		}
	}
	if in.Relevance != nil && *in.Relevance < SkipRelevance {
		return ScreenResult{
			Action: ScreenSkip,
			Reason: "not relevant to the stated purpose (relevance " + format2(*in.Relevance) + ")",
		}
	}
	return ScreenResult{Action: ScreenPass, Reason: "no signals above thresholds"}
}

func ExistsVerdict(exists float64) Exists {
	if exists >= ExistsAnswered {
		return ExistsAnsweredVerdict
	}
	if exists < ExistsAbsent {
		return ExistsAbsentVerdict
	}
	return ExistsPartialVerdict
}

func Rank(items []Item, probabilities map[string]float64) []Ranked {
	out := make([]Ranked, len(items))
	for i, item := range items {
		out[i] = Ranked{ID: item.ID, Text: item.Text, Probability: probabilities[item.ID]}
	}
	slices.SortStableFunc(out, func(a, b Ranked) int {
		switch {
		case a.Probability > b.Probability:
			return -1
		case a.Probability < b.Probability:
			return 1
		default:
			return 0
		}
	})
	return out
}

func MarginOf(probabilities map[string]float64) float64 {
	if len(probabilities) < 2 {
		return 0
	}
	ranked := make([]float64, 0, len(probabilities))
	for _, p := range probabilities {
		ranked = append(ranked, p)
	}
	slices.Sort(ranked)
	return ranked[len(ranked)-1] - ranked[len(ranked)-2]
}

func ClassifyDecision(top, margin float64, gates ClassifyGates) Decision {
	if top >= gates.AutoAccept && margin >= gates.MinimumMargin {
		return DecisionAuto
	}
	return DecisionReview
}

func ValidChoice(choice string, probabilities map[string]float64, expected map[string]struct{}) bool {
	if _, ok := expected[choice]; !ok {
		return false
	}
	if len(probabilities) < len(expected) {
		return false
	}
	sum := 0.0
	for k, p := range probabilities {
		if _, ok := expected[k]; !ok {
			return false
		}
		if p < 0 || p > 1 {
			return false
		}
		sum += p
	}
	if sum < 0.99 || sum > 1.01 {
		return false
	}
	return true
}

func format2(v float64) string {
	return strconv.FormatFloat(v, 'f', 2, 64)
}
