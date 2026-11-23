package kvstore

import (
	"os"
	"path/filepath"
	"testing"
)

// writeStoreWithMode writes a store file with known contents and then sets its
// permission bits exactly. os.WriteFile alone is subject to the umask, so the
// explicit Chmod is what makes the starting mode the one the test asserts on.
func writeStoreWithMode(t *testing.T, path string, mode os.FileMode) {
	t.Helper()
	if err := os.WriteFile(path, []byte(`{"a":"1","b":"2"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, mode); err != nil {
		t.Fatal(err)
	}
}

// modeOf returns the permission bits of the file at path.
func modeOf(t *testing.T, path string) os.FileMode {
	t.Helper()
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	return fi.Mode().Perm()
}

// AC2: Set on an existing store file keeps its permission bits.
func TestSetPreservesExistingMode(t *testing.T) {
	cases := []struct {
		name string
		mode os.FileMode
	}{
		{"AC2_set_keeps_0644", 0o644},
		{"AC2_set_keeps_0640", 0o640},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := filepath.Join(t.TempDir(), "x.json")
			writeStoreWithMode(t, p, tc.mode)
			s, err := Open(p)
			if err != nil {
				t.Fatal(err)
			}
			if err := s.Set("c", "3"); err != nil {
				t.Fatal(err)
			}
			if got := modeOf(t, p); got != tc.mode {
				t.Fatalf("mode after Set = %o, want %o", got, tc.mode)
			}
		})
	}
}

// AC2: Delete on an existing store file keeps its permission bits.
func TestDeletePreservesExistingMode(t *testing.T) {
	cases := []struct {
		name string
		mode os.FileMode
	}{
		{"AC2_delete_keeps_0644", 0o644},
		{"AC2_delete_keeps_0640", 0o640},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := filepath.Join(t.TempDir(), "x.json")
			writeStoreWithMode(t, p, tc.mode)
			s, err := Open(p)
			if err != nil {
				t.Fatal(err)
			}
			if err := s.Delete("a"); err != nil {
				t.Fatal(err)
			}
			if got := modeOf(t, p); got != tc.mode {
				t.Fatalf("mode after Delete = %o, want %o", got, tc.mode)
			}
		})
	}
}

// I13: a store file that does not yet exist is created with mode 0600.
// This already holds on main; the test pins it so a regression is caught.
func TestI13_NewStoreIs0600(t *testing.T) {
	p := filepath.Join(t.TempDir(), "new.json")
	s, err := Open(p)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Set("a", "1"); err != nil {
		t.Fatal(err)
	}
	if got := modeOf(t, p); got != 0o600 {
		t.Fatalf("mode of new store file = %o, want 600", got)
	}
}
