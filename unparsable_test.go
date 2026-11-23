package main

import (
	"bytes"
	"os"
	"testing"
)

func TestI8_EveryCommandOnUnparsableStoreFailsAndLeavesItUnchanged(t *testing.T) {
	bad := []byte(`{"a":"1",bad`)
	for _, tc := range []struct {
		name string
		args []string
	}{
		{"I8_kv_set", []string{"set", "a", "2"}},
		{"I8_kv_get", []string{"get", "a"}},
		{"I8_kv_del", []string{"del", "a"}},
		{"I8_kv_list", []string{"list"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := setup(t)
			if err := os.WriteFile(p, bad, 0o644); err != nil {
				t.Fatalf("write store: %v", err)
			}
			if c, o, e := do(t, tc.args...); c != 1 || o != "" || e == "" {
				t.Errorf("kv %v = %d %q %q, want 1 \"\" <error>", tc.args, c, o, e)
			}
			got, err := os.ReadFile(p)
			if err != nil {
				t.Fatalf("read store: %v", err)
			}
			if !bytes.Equal(got, bad) {
				t.Errorf("store after kv %v = %q, want unchanged %q", tc.args, got, bad)
			}
		})
	}
}
