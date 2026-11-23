package kvstore

import (
	"fmt"
	"path/filepath"
	"sync"
	"testing"
)

// AC3: writes from several handles opened on one file before any of them writes are all kept.
// This models the CLI, where each `kv set` process opens the store, then writes.
func TestConcurrentSetKeepsAllWrites(t *testing.T) {
	p := filepath.Join(t.TempDir(), "x.json")
	const n = 20
	handles := make([]*Store, n)
	for i := range handles {
		s, err := Open(p)
		if err != nil {
			t.Fatalf("Open handle %d: %v", i, err)
		}
		handles[i] = s
	}

	errs := make([]error, n)
	var wg sync.WaitGroup
	for i, s := range handles {
		wg.Add(1)
		go func(i int, s *Store) {
			defer wg.Done()
			errs[i] = s.Set(fmt.Sprintf("k%02d", i), fmt.Sprintf("v%02d", i))
		}(i, s)
	}
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			t.Fatalf("Set from handle %d: %v", i, err)
		}
	}

	final, err := Open(p)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < n; i++ {
		key, want := fmt.Sprintf("k%02d", i), fmt.Sprintf("v%02d", i)
		if got, ok := final.Get(key); !ok || got != want {
			t.Errorf("after concurrent Sets: %s = %q (present %v), want %q", key, got, ok, want)
		}
	}
	if got := len(final.List()); got != n {
		t.Errorf("after concurrent Sets: store holds %d keys, want %d", got, n)
	}
}

// AC4: a key deleted by one handle is not brought back by a Set from a handle opened before the delete.
func TestDeleteNotResurrectedByStaleHandle(t *testing.T) {
	p := filepath.Join(t.TempDir(), "x.json")
	seed, err := Open(p)
	if err != nil {
		t.Fatal(err)
	}
	if err := seed.Set("d0", "0"); err != nil {
		t.Fatalf("seed Set: %v", err)
	}

	a, err := Open(p)
	if err != nil {
		t.Fatal(err)
	}
	b, err := Open(p)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Delete("d0"); err != nil {
		t.Fatalf("A.Delete(d0): %v", err)
	}
	if err := b.Set("x", "1"); err != nil {
		t.Fatalf("B.Set(x): %v", err)
	}

	final, err := Open(p)
	if err != nil {
		t.Fatal(err)
	}
	if v, ok := final.Get("d0"); ok {
		t.Errorf("after A deleted d0 and B set x: d0 = %q, want absent", v)
	}
	if v, ok := final.Get("x"); !ok || v != "1" {
		t.Errorf("after A deleted d0 and B set x: x = %q (present %v), want \"1\"", v, ok)
	}
}
