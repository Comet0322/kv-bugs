package kvstore

import (
	"os"
	"path/filepath"
	"testing"
)

// existingStore writes a store file holding {"k":"v"} and chmods it to perm.
func existingStore(t *testing.T, perm os.FileMode) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "x.json")
	if err := os.WriteFile(p, []byte(`{"k":"v"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(p, perm); err != nil {
		t.Fatal(err)
	}
	return p
}

func modeOf(t *testing.T, p string) os.FileMode {
	t.Helper()
	fi, err := os.Stat(p)
	if err != nil {
		t.Fatalf("stat %s: %v", p, err)
	}
	return fi.Mode().Perm()
}

func TestSetKeepsExistingFileMode(t *testing.T) {
	for _, tc := range []struct {
		name string
		perm os.FileMode
	}{
		{"AC2_set_keeps_0644", 0o644},
		{"AC2_set_keeps_0640", 0o640},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := existingStore(t, tc.perm)
			s, err := Open(p)
			if err != nil {
				t.Fatal(err)
			}
			if err := s.Set("a", "1"); err != nil {
				t.Fatalf("Set: got error %v, want nil", err)
			}
			r, err := Open(p)
			if err != nil {
				t.Fatal(err)
			}
			if got, ok := r.Get("a"); !ok || got != "1" {
				t.Fatalf("reopened Get(a) = %q, %v; want \"1\", true", got, ok)
			}
			if got := modeOf(t, p); got != tc.perm {
				t.Fatalf("mode after Set = %#o, want %#o", got, tc.perm)
			}
		})
	}
}

func TestDeleteKeepsExistingFileMode(t *testing.T) {
	for _, tc := range []struct {
		name string
		perm os.FileMode
	}{
		{"AC3_delete_keeps_0644", 0o644},
		{"AC3_delete_keeps_0640", 0o640},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := existingStore(t, tc.perm)
			s, err := Open(p)
			if err != nil {
				t.Fatal(err)
			}
			if err := s.Delete("k"); err != nil {
				t.Fatalf("Delete(k): got error %v, want nil", err)
			}
			r, err := Open(p)
			if err != nil {
				t.Fatal(err)
			}
			if got, ok := r.Get("k"); ok {
				t.Fatalf("reopened Get(k) = %q, true; want absent", got)
			}
			if got := modeOf(t, p); got != tc.perm {
				t.Fatalf("mode after Delete = %#o, want %#o", got, tc.perm)
			}
		})
	}
}

func TestI5_NewStoreIs0600(t *testing.T) {
	withUmask(t, 0o022)
	p := filepath.Join(t.TempDir(), "x.json")
	s, err := Open(p)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Set("a", "1"); err != nil {
		t.Fatalf("Set: got error %v, want nil", err)
	}
	if got := modeOf(t, p); got != 0o600 {
		t.Fatalf("mode of new store = %#o, want %#o", got, os.FileMode(0o600))
	}
}
