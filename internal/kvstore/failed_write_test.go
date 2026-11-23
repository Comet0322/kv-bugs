package kvstore

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// I11: a failed Set leaves the in-memory store as it was before the call.
func TestI11_FailedSetLeavesInMemoryStoreUnchanged(t *testing.T) {
	p := filepath.Join(t.TempDir(), "x.json")
	s, err := Open(p)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Set("a", "1"); err != nil {
		t.Fatal(err)
	}
	// Make the next save fail: the store path becomes a non-empty directory,
	// so renaming the temp file over it fails.
	if err := os.Remove(p); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(p, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(p, "f"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	before := s.List()

	if err := s.Set("b", "2"); err == nil {
		t.Fatal("Set(new key) on failing store: got nil error, want an error")
	}
	if err := s.Set("a", "9"); err == nil {
		t.Fatal("Set(existing key) on failing store: got nil error, want an error")
	}
	if v, ok := s.Get("b"); ok {
		t.Fatalf("after failed Set, Get(\"b\") = %q, true; want absent", v)
	}
	if v, ok := s.Get("a"); !ok || v != "1" {
		t.Fatalf("after failed Set, Get(\"a\") = %q, %v; want \"1\", true", v, ok)
	}
	if got := s.List(); !reflect.DeepEqual(got, before) {
		t.Fatalf("after failed Set, List() = %v; want %v (unchanged)", got, before)
	}
}
