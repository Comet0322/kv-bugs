package kvstore

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// A store file that becomes unparsable after Open must make Set and Delete
// fail and stay untouched: they re-read the file before writing.
func TestI8_StoreWritesOnFileTurnedUnparsableFailAndLeaveIt(t *testing.T) {
	bad := []byte(`{"a":"1",bad`)
	for _, tc := range []struct {
		name  string
		write func(s *Store) error
	}{
		{"I8_Store_Set", func(s *Store) error { return s.Set("b", "2") }},
		{"I8_Store_Delete", func(s *Store) error { return s.Delete("a") }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := filepath.Join(t.TempDir(), "x.json")
			if err := os.WriteFile(p, []byte(`{"a":"1"}`), 0o600); err != nil {
				t.Fatalf("write store: %v", err)
			}
			s, err := Open(p)
			if err != nil {
				t.Fatalf("Open: %v", err)
			}
			if err := os.WriteFile(p, bad, 0o600); err != nil {
				t.Fatalf("corrupt store: %v", err)
			}
			if err := tc.write(s); err == nil {
				t.Errorf("%s on unparsable store: err = nil, want parse error", tc.name)
			}
			got, err := os.ReadFile(p)
			if err != nil {
				t.Fatalf("read store: %v", err)
			}
			if !bytes.Equal(got, bad) {
				t.Errorf("store after %s = %q, want unchanged %q", tc.name, got, bad)
			}
		})
	}
}
