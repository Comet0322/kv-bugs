package main

import (
	"bytes"
	"os"
	"testing"
)

// I4: set against an unparseable store exits 1 and leaves the file byte-identical.
func TestI4_SetOnUnparseableStoreExits1AndLeavesFileIdentical(t *testing.T) {
	p := setup(t)
	orig := []byte("{bad")
	if err := os.WriteFile(p, orig, 0o644); err != nil {
		t.Fatal(err)
	}
	c, o, e := do(t, "set", "a", "1")
	if c != 1 || o != "" || e == "" {
		t.Fatalf("set on unparseable store: exit %d stdout %q stderr %q, want exit 1, empty stdout, non-empty stderr", c, o, e)
	}
	got, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, orig) {
		t.Fatalf("store bytes after failed set = %q, want %q (byte-identical)", got, orig)
	}
}

// I4: del against an unparseable store exits 1 and leaves the file byte-identical.
func TestI4_DelOnUnparseableStoreExits1AndLeavesFileIdentical(t *testing.T) {
	p := setup(t)
	orig := []byte("{bad")
	if err := os.WriteFile(p, orig, 0o644); err != nil {
		t.Fatal(err)
	}
	c, o, e := do(t, "del", "a")
	if c != 1 || o != "" || e == "" {
		t.Fatalf("del on unparseable store: exit %d stdout %q stderr %q, want exit 1, empty stdout, non-empty stderr", c, o, e)
	}
	got, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, orig) {
		t.Fatalf("store bytes after failed del = %q, want %q (byte-identical)", got, orig)
	}
}

// I5: a 0-byte store file is an empty store, and set succeeds on it.
func TestI5_EmptyStoreIsEmptyAndSetSucceeds(t *testing.T) {
	p := setup(t)
	if err := os.WriteFile(p, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if c, o, e := do(t, "list"); c != 0 || o != "" || e != "" {
		t.Fatalf("list on 0-byte store: exit %d stdout %q stderr %q, want exit 0 and no output", c, o, e)
	}
	if c, o, e := do(t, "set", "z", "9"); c != 0 || o != "" || e != "" {
		t.Fatalf("set on 0-byte store: exit %d stdout %q stderr %q, want exit 0 and no output", c, o, e)
	}
	if c, o, _ := do(t, "list"); c != 0 || o != "z=9\n" {
		t.Fatalf("list after set on 0-byte store: exit %d stdout %q, want exit 0 and %q", c, o, "z=9\n")
	}
}
