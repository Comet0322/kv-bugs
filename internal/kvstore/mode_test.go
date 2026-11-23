package kvstore

import (
	"os"
	"path/filepath"
	"testing"
)

// AC2: Set and Delete keep the permission bits of an existing store file.
func TestSetPreservesExistingMode(t *testing.T) {
	rows := []struct {
		name string
		mode os.FileMode
	}{
		{"AC2_mode_0644", 0o644},
		{"AC2_mode_0640", 0o640},
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			p := filepath.Join(t.TempDir(), "x.json")
			if err := os.WriteFile(p, []byte("{}"), 0o600); err != nil {
				t.Fatal(err)
			}
			if err := os.Chmod(p, row.mode); err != nil {
				t.Fatal(err)
			}

			s, err := Open(p)
			if err != nil {
				t.Fatal(err)
			}
			if err := s.Set("a", "9"); err != nil {
				t.Fatalf("Set: unexpected error %v", err)
			}
			assertStoreMode(t, p, row.mode, "after Set")
			assertStoredValue(t, p, "a", "9", true)

			if err := s.Delete("a"); err != nil {
				t.Fatalf("Delete: unexpected error %v", err)
			}
			assertStoreMode(t, p, row.mode, "after Delete")
			assertStoredValue(t, p, "a", "", false)
		})
	}
}

// I6: a store file created by Set on a path that does not exist has mode 0600.
func TestNewStoreFileMode0600(t *testing.T) {
	p := filepath.Join(t.TempDir(), "new.json")
	s, err := Open(p)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Set("a", "1"); err != nil {
		t.Fatalf("Set: unexpected error %v", err)
	}
	assertStoreMode(t, p, 0o600, "for a new store file")
}

// assertStoreMode fails the test unless the file at p has exactly the permission bits want.
func assertStoreMode(t *testing.T, p string, want os.FileMode, when string) {
	t.Helper()
	info, err := os.Stat(p)
	if err != nil {
		t.Fatalf("stat %s %s: %v", p, when, err)
	}
	if got := info.Mode().Perm(); got != want {
		t.Fatalf("mode of %s %s = %04o, want %04o", p, when, got, want)
	}
}

// assertStoredValue reopens the store at p and checks that key k is present (with value v) or absent.
func assertStoredValue(t *testing.T, p, k, v string, present bool) {
	t.Helper()
	s, err := Open(p)
	if err != nil {
		t.Fatalf("reopen %s: %v", p, err)
	}
	got, ok := s.Get(k)
	if ok != present || (present && got != v) {
		t.Fatalf("reopened store: key %q = %q (present %v), want %q (present %v)", k, got, ok, v, present)
	}
}
