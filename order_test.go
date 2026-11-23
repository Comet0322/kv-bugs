package main

import (
	"strings"
	"testing"
)

func TestAC1_KvListPrintsByteWiseAscending(t *testing.T) {
	setup(t)
	// The AC's example keys (b, aa, B, a) plus more, so that an unsorted
	// (map-iteration) order cannot match the expected order by chance.
	for _, k := range []string{"b", "aa", "B", "a", "ba", "Z", "ab", "0", "zz", "A", "b0", "~", "a0", "AA"} {
		if c, o, e := do(t, "set", k, "v"+k); c != 0 {
			t.Fatalf("set %s: exit %d stdout %q stderr %q, want exit 0", k, c, o, e)
		}
	}
	var want strings.Builder
	for _, k := range []string{"0", "A", "AA", "B", "Z", "a", "a0", "aa", "ab", "b", "b0", "ba", "zz", "~"} {
		want.WriteString(k + "=v" + k + "\n")
	}
	c, o, e := do(t, "list")
	if c != 0 || o != want.String() || e != "" {
		t.Fatalf("list: got exit %d stdout %q stderr %q, want exit 0 stdout %q stderr %q", c, o, e, want.String(), "")
	}
}
