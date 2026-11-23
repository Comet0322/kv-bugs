package main

import (
	"bytes"
	"os"
	"testing"
)

func TestCLICorruptStoreIsErrorAndUntouched(t *testing.T) {
	corrupt := []byte("{bad")
	for _, tc := range []struct {
		name string
		args []string
	}{
		{"I8_kv_set_on_corrupt_store", []string{"set", "a", "1"}},
		{"I8_kv_get_on_corrupt_store", []string{"get", "a"}},
		{"I8_kv_del_on_corrupt_store", []string{"del", "a"}},
		{"I8_kv_list_on_corrupt_store", []string{"list"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := setup(t)
			if err := os.WriteFile(p, corrupt, 0o644); err != nil {
				t.Fatal(err)
			}
			if c, o, e := do(t, tc.args...); c != 1 || o != "" || e == "" {
				t.Errorf("kv %v on corrupt store: exit %d stdout %q stderr %q, want exit 1, empty stdout, an error on stderr", tc.args, c, o, e)
			}
			got, err := os.ReadFile(p)
			if err != nil {
				t.Fatalf("read store after kv %v: %v", tc.args, err)
			}
			if !bytes.Equal(got, corrupt) {
				t.Errorf("store bytes after kv %v = %q, want unchanged %q", tc.args, got, corrupt)
			}
		})
	}
}
