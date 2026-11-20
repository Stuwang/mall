package callpath

import "testing"

func TestNormalizeTargets(t *testing.T) {
	got := NormalizeTargets([]string{"svc-a, svc-b", "svc-a", "svc-c"})
	want := []string{"svc-a", "svc-b", "svc-c"}
	if len(got) != len(want) {
		t.Fatalf("len=%d want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v want %v", got, want)
		}
	}
}

func TestAppendTrace(t *testing.T) {
	if got := AppendTrace("", "biz-a"); got != "biz-a" {
		t.Fatalf("got %q", got)
	}
	if got := AppendTrace("client", "biz-a"); got != "client>biz-a" {
		t.Fatalf("got %q", got)
	}
}

func TestClampDepth(t *testing.T) {
	if got := ClampDepth(0); got != 1 {
		t.Fatalf("got %d", got)
	}
	if got := ClampDepth(2); got != 2 {
		t.Fatalf("got %d", got)
	}
	if got := ClampDepth(9); got != 3 {
		t.Fatalf("got %d", got)
	}
}

func TestDownstream(t *testing.T) {
	got := Downstream("svc-b", []string{"svc-a", "svc-b", "svc-c"})
	if len(got) != 2 || got[0] != "svc-a" || got[1] != "svc-c" {
		t.Fatalf("got %v", got)
	}
}
