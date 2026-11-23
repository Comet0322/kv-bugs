package kvstore

import (
	"path/filepath"
	"testing"
)

// AC1: Store.List returns pairs in plain byte-wise ascending key order,
// not ordered by key length first.
func TestListIsByteWiseAscending(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "x.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, kv := range [][2]string{{"b", "1"}, {"aa", "2"}, {"B", "3"}, {"a", "4"}} {
		if err := s.Set(kv[0], kv[1]); err != nil {
			t.Fatal(err)
		}
	}
	want := []string{"B", "a", "aa", "b"}
	got := s.List()
	if len(got) != len(want) {
		t.Fatalf("List() returned %d pairs, want %d: %v", len(got), len(want), got)
	}
	for i, p := range got {
		if p.Key != want[i] {
			keys := make([]string, 0, len(got))
			for _, q := range got {
				keys = append(keys, q.Key)
			}
			t.Fatalf("List() key order = %v, want %v (first mismatch at index %d)", keys, want, i)
		}
	}
}
