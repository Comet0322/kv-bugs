//go:build unix

package main

import (
	"os"
	"testing"
)

// existingStoreCLI points KV_FILE at a store holding keep=k, gone=g with the
// given permission bits.
func existingStoreCLI(t *testing.T, mode os.FileMode) string {
	t.Helper()
	p := setup(t)
	if err := os.WriteFile(p, []byte(`{"keep":"k","gone":"g"}`), 0o600); err != nil {
		t.Fatalf("write store: %v", err)
	}
	if err := os.Chmod(p, mode); err != nil {
		t.Fatalf("chmod store: %v", err)
	}
	return p
}

func permOf(t *testing.T, p string) os.FileMode {
	t.Helper()
	fi, err := os.Stat(p)
	if err != nil {
		t.Fatalf("stat %s: %v", p, err)
	}
	return fi.Mode().Perm()
}

func TestCLISetPreservesPermissions(t *testing.T) {
	for _, tc := range []struct {
		name string
		mode os.FileMode
	}{
		{"AC2_kv_set_keeps_0644", 0o644},
		{"AC2_kv_set_keeps_0640", 0o640},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := existingStoreCLI(t, tc.mode)
			if c, o, e := do(t, "set", "new", "v"); c != 0 || o != "" || e != "" {
				t.Fatalf("kv set = %d %q %q, want 0 \"\" \"\"", c, o, e)
			}
			if got := permOf(t, p); got != tc.mode {
				t.Errorf("mode after kv set = %#o, want %#o", got, tc.mode)
			}
			if c, o, _ := do(t, "get", "new"); c != 0 || o != "v\n" {
				t.Errorf("kv get new = %d %q, want 0 \"v\\n\"", c, o)
			}
		})
	}
}
