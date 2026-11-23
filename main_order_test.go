package main

import "testing"

// AC1: kv list prints keys in plain byte-wise ascending order (not length-first).
func TestListByteOrder(t *testing.T) {
	setup(t)
	do(t, "set", "b", "2")
	do(t, "set", "aa", "1")
	do(t, "set", "B", "0")
	want := "B=0\naa=1\nb=2\n"
	if c, o, _ := do(t, "list"); c != 0 || o != want {
		t.Fatalf("list: code=%d stdout=%q, want code=0 stdout=%q", c, o, want)
	}
}
