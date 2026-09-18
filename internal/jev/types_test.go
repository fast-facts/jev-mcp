package jev

import "testing"

func Test_ValidState(t *testing.T) {
	ok := []any{"hello", map[string]any{"k": "v"}, []any{"a"}}
	for _, v := range ok {
		if err := ValidState(v); err != nil {
			t.Fatalf("%v: %v", v, err)
		}
	}
	for _, v := range []any{nil, 3.2, true} {
		if ValidState(v) == nil {
			t.Fatalf("accepted %v", v)
		}
	}
}

func Test_Question_Validate_matches_api_contract(t *testing.T) {
	if err := (Question{Type: "noul", Instructions: "urgent?"}).Validate(); err != nil {
		t.Fatal(err)
	}
	if (Question{Type: "noul"}).Validate() == nil {
		t.Fatal("noul needs instructions")
	}
	if (Question{Type: "choice", Instructions: "which?", Criteria: map[string]string{}}).Validate() == nil {
		t.Fatal("empty choice criteria")
	}
	if err := (Question{Type: "choice", Instructions: "which?", Criteria: map[string]string{"a": "A"}}).Validate(); err != nil {
		t.Fatal(err)
	}
	if (Question{Type: "score", Instructions: "how?", Criteria: []string{"only"}}).Validate() == nil {
		t.Fatal("one score level")
	}
	if err := (Question{Type: "score", Instructions: "how?", Criteria: []string{"low", "high"}}).Validate(); err != nil {
		t.Fatal(err)
	}
}
