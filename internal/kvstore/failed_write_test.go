package kvstore

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// newStoreWithFailingSave opens a store at a fresh path, seeds it with a=1 and b=2, then makes every
// later save fail: the store path is replaced by a non-empty directory, so renaming the temp file over
// it fails. It returns the handle and the directory that holds the store path.
func newStoreWithFailingSave(t *testing.T) (*Store, string) {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "x.json")
	s, err := Open(p)
	if err != nil {
		t.Fatal(err)
	}
	for _, kv := range [][2]string{{"a", "1"}, {"b", "2"}} {
		if err := s.Set(kv[0], kv[1]); err != nil {
			t.Fatalf("seed Set(%q): %v", kv[0], err)
		}
	}
	if err := os.Remove(p); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(p, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(p, "f"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	return s, dir
}

// I5: a failed Set or Delete leaves the handle's in-memory map unchanged and leaves no temp file behind.
// Each row performs one failing write on a fresh handle and checks the map and the directory afterwards.
func TestFailedWriteKeepsMemory(t *testing.T) {
	rows := []struct {
		name string
		op   func(s *Store) error
	}{
		{"I5_set_new_key", func(s *Store) error { return s.Set("c", "3") }},
		{"I5_set_overwrite", func(s *Store) error { return s.Set("a", "9") }},
		{"I5_delete_existing", func(s *Store) error { return s.Delete("a") }},
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			s, dir := newStoreWithFailingSave(t)
			before := s.List()

			if err := row.op(s); err == nil {
				t.Fatal("expected the write to fail, got nil error")
			}

			if after := s.List(); !reflect.DeepEqual(after, before) {
				t.Errorf("in-memory map after failed write = %v, want unchanged %v", after, before)
			}

			es, err := os.ReadDir(dir)
			if err != nil {
				t.Fatal(err)
			}
			if len(es) != 1 || es[0].Name() != "x.json" {
				names := []string{}
				for _, e := range es {
					names = append(names, e.Name())
				}
				t.Errorf("directory after failed write = %v, want only [x.json] (no temp file left)", names)
			}
		})
	}
}
