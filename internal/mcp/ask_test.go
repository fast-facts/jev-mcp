package mcp

import (
	"context"
	"testing"

	"github.com/fast-facts/jev-mcp/internal/jev"
)

func Test_ask_rejects_bad_input(t *testing.T) {
	h := &handlers{}
	cases := []AskInput{
		{State: "hi", Questions: map[string]AskQuestion{"q": {Type: "choice", Instructions: "which?"}}},
		{State: "hi", Questions: map[string]AskQuestion{"q": {Type: "chat", Instructions: "write a poem"}}},
		{Questions: map[string]AskQuestion{"q": {Type: "noul", Instructions: "yes?"}}},
	}
	for _, in := range cases {
		if _, _, err := h.ask(context.Background(), nil, in); err == nil {
			t.Fatalf("expected error for %+v", in)
		}
	}
}

func Test_ask_returns_raw_answers(t *testing.T) {
	h := newTestHandlers(t, map[string]jev.Answer{
		"dept": {Type: "choice", Choice: "billing", Confidence: 0.8, Probabilities: map[string]float64{"billing": 0.8, "tech": 0.2}},
	})
	_, out, err := h.ask(context.Background(), nil, AskInput{
		State: "I was charged twice",
		Questions: map[string]AskQuestion{
			"dept": {Type: "choice", Instructions: "Which team?", Criteria: map[string]string{"billing": "pay", "tech": "bugs"}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Answers["dept"].Choice != "billing" {
		t.Fatalf("%+v", out.Answers)
	}
}
