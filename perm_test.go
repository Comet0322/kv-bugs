package main

import (
	"os"
	"testing"
)

// existingStoreFile writes {"k":"v"} to the KV_FILE store and chmods it to perm.
func existingStoreFile(t *testing.T, perm os.FileMode) string {
	t.Helper()
	p := setup(t)
	if err := os.WriteFile(p, []byte(`{"k":"v"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(p, perm); err != nil {
		t.Fatal(err)
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

func TestCLISetKeepsExistingFileMode(t *testing.T) {
	for _, tc := range []struct {
		name string
		perm os.FileMode
	}{
		{"AC2_kv_set_keeps_0644", 0o644},
		{"AC2_kv_set_keeps_0640", 0o640},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := existingStoreFile(t, tc.perm)
			if c, o, e := do(t, "set", "a", "1"); c != 0 {
				t.Fatalf("kv set a 1: exit %d (stdout %q, stderr %q), want 0", c, o, e)
			}
			if c, o, _ := do(t, "get", "a"); c != 0 || o != "1\n" {
				t.Fatalf("kv get a: exit %d stdout %q, want 0 %q", c, o, "1\n")
			}
			if got := permOf(t, p); got != tc.perm {
				t.Fatalf("mode after kv set = %#o, want %#o", got, tc.perm)
			}
		})
	}
}
