package kvstore

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestAC1_ListIsByteWiseAscending(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "kv.json"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	for _, kv := range [][2]string{{"b", "2"}, {"aa", "1"}, {"B", "3"}, {"a", "0"}} {
		if err := s.Set(kv[0], kv[1]); err != nil {
			t.Fatalf("Set(%q): %v", kv[0], err)
		}
	}
	got := s.List()
	want := []Pair{{"B", "3"}, {"a", "0"}, {"aa", "1"}, {"b", "2"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("List() order: got %v, want %v", got, want)
	}
}
