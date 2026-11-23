package kvstore

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestOpenMissingFileIsEmpty(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "x.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(s.List()) != 0 {
		t.Fatal("expected empty")
	}
}

func TestSetGetDeleteList(t *testing.T) {
	p := filepath.Join(t.TempDir(), "x.json")
	s, _ := Open(p)
	for _, kv := range [][2]string{{"b", "2"}, {"a", "1"}, {"c", "3"}} {
		if err := s.Set(kv[0], kv[1]); err != nil {
			t.Fatal(err)
		}
	}
	if v, ok := s.Get("a"); !ok || v != "1" {
		t.Fatalf("get a = %q %v", v, ok)
	}
	if _, ok := s.Get("zz"); ok {
		t.Fatal("zz should be missing")
	}
	l := s.List()
	if len(l) != 3 || l[0].Key != "a" || l[1].Key != "b" || l[2].Key != "c" {
		t.Fatalf("bad order %v", l)
	}
	if err := s.Delete("b"); err != nil {
		t.Fatal(err)
	}
	if err := s.Delete("b"); err != ErrNotFound {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
	s2, err := Open(p)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := s2.Get("b"); ok || len(s2.List()) != 2 {
		t.Fatal("persistence failed")
	}
}

func TestNoTempFilesLeft(t *testing.T) {
	d := t.TempDir()
	s, _ := Open(filepath.Join(d, "x.json"))
	s.Set("a", "1")
	s.Set("a", "2")
	s.Delete("a")
	es, _ := os.ReadDir(d)
	if len(es) != 1 || es[0].Name() != "x.json" {
		t.Fatalf("unexpected files: %v", es)
	}
}

func TestSetFailureCleansTemp(t *testing.T) {
	d := t.TempDir()
	// target is a directory: Rename fails, temp must be removed.
	target := filepath.Join(d, "x.json")
	os.Mkdir(target, 0o755)
	os.WriteFile(filepath.Join(target, "f"), nil, 0o644)
	s := &Store{path: target, data: map[string]string{}}
	if err := s.Set("a", "1"); err == nil {
		t.Fatal("expected error")
	}
	es, _ := os.ReadDir(d)
	if len(es) != 1 {
		t.Fatalf("temp left: %v", es)
	}
}

func TestOpenInvalidJSON(t *testing.T) {
	p := filepath.Join(t.TempDir(), "x.json")
	os.WriteFile(p, []byte("{nope"), 0o644)
	if _, err := Open(p); err == nil {
		t.Fatal("expected error")
	}
}

func TestListByteOrder(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "x.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Set("b", "2"); err != nil {
		t.Fatal(err)
	}
	if err := s.Set("aa", "1"); err != nil {
		t.Fatal(err)
	}
	l := s.List()
	if len(l) != 2 {
		t.Fatalf("List() len = %d, want 2: %v", len(l), l)
	}
	if l[0].Key != "aa" || l[1].Key != "b" {
		t.Fatalf("List() keys = [%q %q], want [\"aa\" \"b\"] (plain byte-wise order)", l[0].Key, l[1].Key)
	}
}

// AC3: Set keeps the permission bits of an existing store file.
func TestSetPreservesMode(t *testing.T) {
	p := filepath.Join(t.TempDir(), "x.json")
	s, err := Open(p)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Set("a", "1"); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(p, 0o640); err != nil {
		t.Fatal(err)
	}
	if err := s.Set("b", "2"); err != nil {
		t.Fatal(err)
	}
	fi, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := fi.Mode().Perm(), os.FileMode(0o640); got != want {
		t.Fatalf("mode after Set = %v, want %v", got, want)
	}
	s2, err := Open(p)
	if err != nil {
		t.Fatal(err)
	}
	if v, ok := s2.Get("b"); !ok || v != "2" {
		t.Fatalf("reopened b = %q %v, want \"2\" true", v, ok)
	}
}

// I9: a store file created by Set where none existed has mode 0600.
func TestI9_NewStoreIs0600(t *testing.T) {
	p := filepath.Join(t.TempDir(), "x.json")
	s, err := Open(p)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Set("a", "1"); err != nil {
		t.Fatal(err)
	}
	fi, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := fi.Mode().Perm(), os.FileMode(0o600); got != want {
		t.Fatalf("mode of new store = %v, want %v", got, want)
	}
}

// AC3: Delete keeps the permission bits of an existing store file.
func TestDelPreservesMode(t *testing.T) {
	p := filepath.Join(t.TempDir(), "x.json")
	s, err := Open(p)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Set("a", "1"); err != nil {
		t.Fatal(err)
	}
	if err := s.Set("b", "2"); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(p, 0o640); err != nil {
		t.Fatal(err)
	}
	if err := s.Delete("a"); err != nil {
		t.Fatal(err)
	}
	fi, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := fi.Mode().Perm(), os.FileMode(0o640); got != want {
		t.Fatalf("mode after Delete = %v, want %v", got, want)
	}
	s2, err := Open(p)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := s2.Get("a"); ok {
		t.Fatal("reopened store still has deleted key a")
	}
	if v, ok := s2.Get("b"); !ok || v != "2" {
		t.Fatalf("reopened b = %q %v, want \"2\" true", v, ok)
	}
}

// openHandles opens n independent Store handles on path, all before any of
// them writes, so every handle starts from the same snapshot.
func openHandles(t *testing.T, path string, n int) []*Store {
	t.Helper()
	hs := make([]*Store, n)
	for i := range hs {
		s, err := Open(path)
		if err != nil {
			t.Fatalf("Open handle %d: %v", i, err)
		}
		hs[i] = s
	}
	return hs
}

// AC4: concurrent Set through independent handles keeps every write.
func TestConcurrentSetKeepsAllWrites(t *testing.T) {
	const n = 20
	p := filepath.Join(t.TempDir(), "x.json")
	hs := openHandles(t, p, n)
	errs := make([]error, n)
	var wg sync.WaitGroup
	for i := range hs {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			errs[i] = hs[i].Set(fmt.Sprintf("k%d", i), fmt.Sprintf("v%d", i))
		}(i)
	}
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			t.Fatalf("concurrent Set k%d: error %v, want nil", i, err)
		}
	}
	s, err := Open(p)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < n; i++ {
		k, want := fmt.Sprintf("k%d", i), fmt.Sprintf("v%d", i)
		if v, ok := s.Get(k); !ok || v != want {
			t.Errorf("after concurrent Set, Get(%q) = %q, %v; want %q, true", k, v, ok, want)
		}
	}
}
