package kvstore

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// A store that was valid when opened but is unparsable when a write re-reads
// it must fail the write and leave the file's bytes alone.
func TestWriteOnStoreCorruptedSinceOpen(t *testing.T) {
	corrupt := []byte("{bad")
	for _, tc := range []struct {
		name  string
		write func(*Store) error
	}{
		{"I8_Set_on_store_corrupted_since_Open", func(s *Store) error { return s.Set("a", "1") }},
		{"I8_Delete_on_store_corrupted_since_Open", func(s *Store) error { return s.Delete("k") }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := filepath.Join(t.TempDir(), "x.json")
			if err := os.WriteFile(p, []byte(`{"k":"v"}`), 0o644); err != nil {
				t.Fatal(err)
			}
			s, err := Open(p)
			if err != nil {
				t.Fatalf("Open valid store: %v", err)
			}
			if err := os.WriteFile(p, corrupt, 0o644); err != nil {
				t.Fatal(err)
			}
			if err := tc.write(s); err == nil {
				t.Errorf("write on store corrupted since Open: err = nil, want a parse error")
			}
			got, err := os.ReadFile(p)
			if err != nil {
				t.Fatalf("read store: %v", err)
			}
			if !bytes.Equal(got, corrupt) {
				t.Errorf("store bytes after write = %q, want unchanged %q", got, corrupt)
			}
		})
	}
}
