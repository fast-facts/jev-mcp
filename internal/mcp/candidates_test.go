package mcp

import (
	"strings"
	"testing"
)

func Test_bindCandidates_rejects_empty_and_reserved_id(t *testing.T) {
	if _, err := bindCandidates(nil); err == nil {
		t.Fatal("empty should fail")
	}
	_, err := bindCandidates([]Candidate{{ID: "none", Text: "x"}})
	if err == nil || !strings.Contains(err.Error(), "reserved") {
		t.Fatalf("err=%v", err)
	}
}
