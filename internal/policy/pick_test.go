package policy

import "testing"

func Test_AcceptPick_requires_answered_and_a_real_id(t *testing.T) {
	ids := map[string]struct{}{"auth": {}}
	if !AcceptPick("auth", 0.98, ids) {
		t.Fatal("high exists and real id should pick")
	}
	if AcceptPick("auth", 0.46, ids) {
		t.Fatal("partial exists should not pick")
	}
	if AcceptPick("auth", 0.14, ids) {
		t.Fatal("absent exists should not pick")
	}
	if AcceptPick(NoneOption, 0.99, ids) {
		t.Fatal("none should not pick")
	}
	if AcceptPick("missing", 0.99, ids) {
		t.Fatal("unknown id should not pick")
	}
}
