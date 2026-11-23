// Package kvstore is a tiny string key-value store persisted as a JSON object.
package kvstore

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

// ErrNotFound is returned by Delete when the key does not exist.
var ErrNotFound = errors.New("key not found")

// Pair is a key/value entry.
type Pair struct {
	Key   string
	Value string
}

// Store is a file-backed key-value store.
type Store struct {
	mu   sync.Mutex // serialises the methods of one handle, which may be shared by goroutines
	path string
	data map[string]string
}

// Open loads the store at path; a missing file is an empty store.
func Open(path string) (*Store, error) {
	s := &Store{path: path, data: map[string]string{}}
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return s, nil
	}
	if err := json.Unmarshal(b, &s.data); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if s.data == nil {
		s.data = map[string]string{}
	}
	return s, nil
}

// Get returns the value for key.
func (s *Store) Get(key string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.data[key]
	return v, ok
}

// Set stores value under key and persists atomically. Another process may write the same
// file, so Set holds the store's directory lock and re-reads the file before applying the change.
func (s *Store) Set(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	unlock, err := lockDir(filepath.Dir(s.path))
	if err != nil {
		return err
	}
	defer unlock()

	fresh, err := Open(s.path)
	if err != nil {
		return err
	}
	fresh.data[key] = value
	if err := s.save(fresh.data); err != nil {
		return err
	}
	s.data = fresh.data
	return nil
}

// Delete removes key, returning ErrNotFound if absent. Like Set, it works on a fresh copy of
// the file, read under the directory lock.
func (s *Store) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	unlock, err := lockDir(filepath.Dir(s.path))
	if err != nil {
		return err
	}
	defer unlock()

	fresh, err := Open(s.path)
	if err != nil {
		return err
	}
	if _, ok := fresh.data[key]; !ok {
		return ErrNotFound
	}
	delete(fresh.data, key)
	if err := s.save(fresh.data); err != nil {
		return err
	}
	s.data = fresh.data
	return nil
}

// List returns all pairs sorted by key.
func (s *Store) List() []Pair {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Pair, 0, len(s.data))
	for k, v := range s.data {
		out = append(out, Pair{k, v})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Key < out[j].Key
	})
	return out
}

// save writes data to a temp file in the same directory, then renames over the target.
// An existing store file keeps its permission bits; a new one keeps the 0600 of CreateTemp.
func (s *Store) save(data map[string]string) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	info, err := os.Stat(s.path)
	exists := err == nil
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(s.path), ".kv-*.tmp")
	if err != nil {
		return err
	}
	name := tmp.Name()
	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		os.Remove(name)
		return err
	}
	if exists {
		if err := tmp.Chmod(info.Mode().Perm()); err != nil {
			tmp.Close()
			os.Remove(name)
			return err
		}
	}
	if err := tmp.Close(); err != nil {
		os.Remove(name)
		return err
	}
	if err := os.Rename(name, s.path); err != nil {
		os.Remove(name)
		return err
	}
	return nil
}
