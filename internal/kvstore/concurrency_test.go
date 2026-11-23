package kvstore

import (
	"fmt"
	"path/filepath"
	"sync"
	"testing"
)

// AC3: concurrent Sets through separate handles all persist.
// Twenty handles are opened on the same path, then each Sets a distinct key
// at the same time. A fresh Open must see all twenty keys.
func TestConcurrentSetsAllPersist(t *testing.T) {
	const n = 20
	p := filepath.Join(t.TempDir(), "x.json")

	handles := make([]*Store, n)
	for i := range handles {
		s, err := Open(p)
		if err != nil {
			t.Fatal(err)
		}
		handles[i] = s
	}

	errs := make(chan error, n)
	var wg sync.WaitGroup
	for i, s := range handles {
		wg.Add(1)
		go func(i int, s *Store) {
			defer wg.Done()
			errs <- s.Set(fmt.Sprintf("k%02d", i), fmt.Sprintf("v%02d", i))
		}(i, s)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent Set returned error: %v", err)
		}
	}

	fresh, err := Open(p)
	if err != nil {
		t.Fatal(err)
	}
	if got := len(fresh.List()); got != n {
		t.Fatalf("keys persisted = %d, want %d", got, n)
	}
	for i := 0; i < n; i++ {
		k := fmt.Sprintf("k%02d", i)
		if v, ok := fresh.Get(k); !ok || v != fmt.Sprintf("v%02d", i) {
			t.Fatalf("key %s = %q (found %v), want %q", k, v, ok, fmt.Sprintf("v%02d", i))
		}
	}
}

// AC4: a Delete is not undone by a concurrent Set of another key.
// Ten keys are set first. Then, at the same time, ten handles (opened before
// any change, so each holds a snapshot that still has all ten keys) delete
// those keys and ten other handles set new keys. Afterwards no pre-set key may
// remain and all ten new keys must be present.
func TestConcurrentDeleteNotResurrected(t *testing.T) {
	const n = 10
	p := filepath.Join(t.TempDir(), "x.json")

	pre, err := Open(p)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < n; i++ {
		if err := pre.Set(fmt.Sprintf("d%02d", i), "old"); err != nil {
			t.Fatal(err)
		}
	}

	delHandles := make([]*Store, n)
	setHandles := make([]*Store, n)
	for i := 0; i < n; i++ {
		if delHandles[i], err = Open(p); err != nil {
			t.Fatal(err)
		}
		if setHandles[i], err = Open(p); err != nil {
			t.Fatal(err)
		}
	}

	errs := make(chan error, 2*n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(2)
		go func(i int) {
			defer wg.Done()
			errs <- delHandles[i].Delete(fmt.Sprintf("d%02d", i))
		}(i)
		go func(i int) {
			defer wg.Done()
			errs <- setHandles[i].Set(fmt.Sprintf("n%02d", i), "new")
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent write returned error: %v", err)
		}
	}

	fresh, err := Open(p)
	if err != nil {
		t.Fatal(err)
	}
	var resurrected []string
	for i := 0; i < n; i++ {
		k := fmt.Sprintf("d%02d", i)
		if _, ok := fresh.Get(k); ok {
			resurrected = append(resurrected, k)
		}
	}
	if len(resurrected) != 0 {
		t.Fatalf("deleted keys came back: %v (want none)", resurrected)
	}
	var missing []string
	for i := 0; i < n; i++ {
		k := fmt.Sprintf("n%02d", i)
		if _, ok := fresh.Get(k); !ok {
			missing = append(missing, k)
		}
	}
	if len(missing) != 0 {
		t.Fatalf("set keys lost: %v (want none)", missing)
	}
}
