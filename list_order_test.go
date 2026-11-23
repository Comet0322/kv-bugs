package main

import "testing"

// AC1: `kv list` prints pairs in plain byte-wise ascending key order,
// not ordered by key length first.
func TestListOutputIsByteWiseAscending(t *testing.T) {
	setup(t)
	for _, kv := range [][2]string{{"b", "1"}, {"aa", "2"}, {"B", "3"}, {"a", "4"}} {
		if c, o, e := do(t, "set", kv[0], kv[1]); c != 0 {
			t.Fatalf("set %s: exit %d stdout %q stderr %q", kv[0], c, o, e)
		}
	}
	c, o, e := do(t, "list")
	if c != 0 || e != "" {
		t.Fatalf("list: exit %d stderr %q, want exit 0 and empty stderr", c, e)
	}
	want := "B=3\na=4\naa=2\nb=1\n"
	if o != want {
		t.Fatalf("list output = %q, want %q", o, want)
	}
}
