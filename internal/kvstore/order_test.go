package kvstore

import (
	"path/filepath"
	"testing"
)

func TestAC1_ListIsByteWiseAscending(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "x.json"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	for _, k := range []string{"b", "aa", "B", "a"} {
		if err := s.Set(k, "v"+k); err != nil {
			t.Fatalf("Set(%q): %v", k, err)
		}
	}
	want := []string{"B", "a", "aa", "b"}
	got := s.List()
	if len(got) != len(want) {
		t.Fatalf("List() returned %d pairs %v, want %d keys %v", len(got), got, len(want), want)
	}
	for i, k := range want {
		if got[i].Key != k || got[i].Value != "v"+k {
			t.Fatalf("List()[%d] = %v, want {%s v%s}; full got %v, want keys %v", i, got[i], k, k, got, want)
		}
	}
}
