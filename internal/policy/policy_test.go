package policy

import (
	"strings"
	"testing"
)

func Test_SanitizeID_keeps_safe_characters(t *testing.T) {
	if got := SanitizeID("src/lib.ts"); got != "src_lib.ts" {
		t.Fatalf("got %q", got)
	}
	if got := SanitizeID("note: hello?!"); got != "note_hello" {
		t.Fatalf("got %q", got)
	}
	if got := SanitizeID("???"); got != "" {
		t.Fatalf("got %q", got)
	}
	if got := SanitizeID(strings.Repeat("a", 100)); len(got) != 64 {
		t.Fatalf("len=%d", len(got))
	}
}

func Test_UniqueIDs_assigns_fallbacks_and_resolves_collisions(t *testing.T) {
	got := UniqueIDs([]Item{
		{ID: "src/lib.ts", Text: "a"},
		{ID: "src/lib.ts", Text: "b"},
		{Text: "c"},
	}, "candidate")
	want := []string{"src_lib.ts", "src_lib.ts_1", "candidate2"}
	for i, id := range want {
		if got[i].ID != id {
			t.Fatalf("index %d: got %q want %q", i, got[i].ID, id)
		}
	}
}

func Test_Truncate_marks_cut_text(t *testing.T) {
	out := Truncate("abcdef", 3)
	if !strings.HasSuffix(out, "…truncated]") {
		t.Fatalf("missing marker: %q", out)
	}
	if Truncate("abc", 3) != "abc" {
		t.Fatal("short text should stay whole")
	}
}

func Test_VerifyAction_gates_on_auto_accept(t *testing.T) {
	if VerifyAction(0.8, 0.8) != DecisionAuto {
		t.Fatal("equal to threshold should auto")
	}
	if VerifyAction(0.79, 0.8) != DecisionReview {
		t.Fatal("below threshold should review")
	}
}

func Test_Screen_escalates_injection_then_skips_junk(t *testing.T) {
	cases := []struct {
		name string
		in   ScreenInput
		want ScreenAction
	}{
		{"block", ScreenInput{Injection: 0.9, BlockAt: 0.75, ReviewAt: 0.25}, ScreenBlock},
		{"review", ScreenInput{Injection: 0.4, BlockAt: 0.75, ReviewAt: 0.25}, ScreenReview},
		{"pass", ScreenInput{Injection: 0.01, BlockAt: 0.75, ReviewAt: 0.25}, ScreenPass},
		{"skip substance", ScreenInput{Injection: 0.01, Substance: ptr(0.1), BlockAt: 0.75, ReviewAt: 0.25}, ScreenSkip},
		{"skip relevance", ScreenInput{Injection: 0.01, Relevance: ptr(0.05), BlockAt: 0.75, ReviewAt: 0.25}, ScreenSkip},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Screen(tc.in); got.Action != tc.want {
				t.Fatalf("got %s want %s", got.Action, tc.want)
			}
		})
	}
}

func Test_ExistsVerdict_uses_cookbook_thresholds(t *testing.T) {
	if ExistsVerdict(0.98) != ExistsAnsweredVerdict {
		t.Fatal("high exists should be answered")
	}
	if ExistsVerdict(0.46) != ExistsPartialVerdict {
		t.Fatal("mid exists should be partial")
	}
	if ExistsVerdict(0.14) != ExistsAbsentVerdict {
		t.Fatal("low exists should be absent")
	}
}

func Test_Rank_orders_by_probability_and_keeps_caller_order_on_ties(t *testing.T) {
	items := []Item{{ID: "a", Text: "1"}, {ID: "b", Text: "2"}, {ID: "c", Text: "3"}}
	ranked := Rank(items, map[string]float64{"a": 0.1, "b": 0.5, "c": 0.1})
	got := []string{ranked[0].ID, ranked[1].ID, ranked[2].ID}
	if got[0] != "b" || got[1] != "a" || got[2] != "c" {
		t.Fatalf("got %v", got)
	}
	missing := Rank([]Item{{ID: "x", Text: "1"}}, map[string]float64{})
	if missing[0].Probability != 0 {
		t.Fatal("missing probability should be 0")
	}
}

func Test_MarginOf_measures_winner_to_runner_up_gap(t *testing.T) {
	if d := MarginOf(map[string]float64{"a": 0.7, "b": 0.2, "c": 0.1}) - 0.5; d > 1e-9 || d < -1e-9 {
		t.Fatalf("want 0.5, delta %v", d)
	}
	if MarginOf(map[string]float64{"a": 0.5, "b": 0.5}) != 0 {
		t.Fatal("tie margin should be 0")
	}
	if MarginOf(map[string]float64{"only": 0.8}) != 0 {
		t.Fatal("lone probability has no runner-up")
	}
	if MarginOf(nil) != 0 {
		t.Fatal("nil map margin should be 0")
	}
}

func Test_ClassifyDecision_requires_top_probability_and_margin(t *testing.T) {
	gates := ClassifyGates{AutoAccept: 0.85, MinimumMargin: 0.5}
	if ClassifyDecision(0.9, 0.6, gates) != DecisionAuto {
		t.Fatal("both gates met")
	}
	if ClassifyDecision(0.9, 0.4, gates) != DecisionReview {
		t.Fatal("thin margin")
	}
	if ClassifyDecision(0.8, 0.8, gates) != DecisionReview {
		t.Fatal("low top probability")
	}
	if ClassifyDecision(0.85, 0.5, gates) != DecisionAuto {
		t.Fatal("exactly at both gates")
	}
}

func Test_caps_stay_inside_choice_limit(t *testing.T) {
	if MaxCandidates > 255 || MaxClasses > 255 {
		t.Fatal("choice option cap is 255")
	}
}

func ptr(v float64) *float64 { return &v }
