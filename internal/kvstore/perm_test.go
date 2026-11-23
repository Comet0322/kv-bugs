//go:build unix

package kvstore

import (
	"os"
	"path/filepath"
	"testing"
)

// existingStore creates a store file holding {"keep":"k","gone":"g"} and
// chmods it to mode (chmod is not subject to the umask).
func existingStore(t *testing.T, mode os.FileMode) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "s.json")
	if err := os.WriteFile(p, []byte(`{"keep":"k","gone":"g"}`), 0o600); err != nil {
		t.Fatalf("write store: %v", err)
	}
	if err := os.Chmod(p, mode); err != nil {
		t.Fatalf("chmod store: %v", err)
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

func TestStoreSetPreservesPermissions(t *testing.T) {
	for _, tc := range []struct {
		name string
		mode os.FileMode
	}{
		{"AC2_store_set_keeps_0644", 0o644},
		{"AC2_store_set_keeps_0640", 0o640},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := existingStore(t, tc.mode)
			s, err := Open(p)
			if err != nil {
				t.Fatalf("Open: %v", err)
			}
			if err := s.Set("new", "v"); err != nil {
				t.Fatalf("Set: %v", err)
			}
			if got := modeOf(t, p); got != tc.mode {
				t.Errorf("mode after Set = %#o, want %#o", got, tc.mode)
			}
			r, err := Open(p)
			if err != nil {
				t.Fatalf("reopen: %v", err)
			}
			if v, ok := r.Get("new"); !ok || v != "v" {
				t.Errorf("after reopen Get(new) = %q, %v; want \"v\", true", v, ok)
			}
		})
	}
}

func TestStoreDeletePreservesPermissions(t *testing.T) {
	for _, tc := range []struct {
		name string
		mode os.FileMode
	}{
		{"AC3_store_delete_keeps_0644", 0o644},
		{"AC3_store_delete_keeps_0640", 0o640},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := existingStore(t, tc.mode)
			s, err := Open(p)
			if err != nil {
				t.Fatalf("Open: %v", err)
			}
			if err := s.Delete("gone"); err != nil {
				t.Fatalf("Delete: %v", err)
			}
			if got := modeOf(t, p); got != tc.mode {
				t.Errorf("mode after Delete = %#o, want %#o", got, tc.mode)
			}
			r, err := Open(p)
			if err != nil {
				t.Fatalf("reopen: %v", err)
			}
			if v, ok := r.Get("gone"); ok {
				t.Errorf("after reopen Get(gone) = %q, true; want deleted", v)
			}
			if v, ok := r.Get("keep"); !ok || v != "k" {
				t.Errorf("after reopen Get(keep) = %q, %v; want \"k\", true", v, ok)
			}
		})
	}
}
