package main

import "testing"

func TestAC1_KvListIsByteWiseAscending(t *testing.T) {
	setup(t)
	for _, kv := range [][2]string{{"b", "2"}, {"aa", "1"}, {"B", "3"}, {"a", "0"}} {
		if c, o, e := do(t, "set", kv[0], kv[1]); c != 0 {
			t.Fatalf("set %s: exit %d, stdout %q, stderr %q", kv[0], c, o, e)
		}
	}
	c, o, e := do(t, "list")
	want := "B=3\na=0\naa=1\nb=2\n"
	if c != 0 || o != want {
		t.Fatalf("kv list: got exit %d stdout %q (stderr %q), want exit 0 stdout %q", c, o, e, want)
	}
}
