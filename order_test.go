package main

import "testing"

func TestAC1_KvListPrintsByteWiseAscending(t *testing.T) {
	setup(t)
	for _, kv := range [][2]string{{"b", "1"}, {"aa", "2"}, {"B", "3"}, {"a", "4"}} {
		if c, o, e := do(t, "set", kv[0], kv[1]); c != 0 {
			t.Fatalf("set %s: exit %d stdout %q stderr %q, want exit 0", kv[0], c, o, e)
		}
	}
	c, o, e := do(t, "list")
	want := "B=3\na=4\naa=2\nb=1\n"
	if c != 0 || o != want || e != "" {
		t.Fatalf("list: got exit %d stdout %q stderr %q, want exit 0 stdout %q stderr %q", c, o, e, want, "")
	}
}
