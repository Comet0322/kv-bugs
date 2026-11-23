package kvstore

import (
	"path/filepath"
	"testing"
)

// AC1: List sorts keys byte-wise ascending, not by length first.
func TestListSortsBytewise(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "x.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, kv := range [][2]string{{"b", "2"}, {"aa", "1"}, {"B", "0"}} {
		if err := s.Set(kv[0], kv[1]); err != nil {
			t.Fatal(err)
		}
	}
	want := []string{"B", "aa", "b"}
	l := s.List()
	if len(l) != len(want) {
		t.Fatalf("list len = %d, want %d", len(l), len(want))
	}
	for i, p := range l {
		if p.Key != want[i] {
			t.Fatalf("list[%d].Key = %q, want %q (full order %v)", i, p.Key, want[i], l)
		}
	}
}
