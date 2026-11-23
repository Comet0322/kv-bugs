// Package kvstore is a tiny string key-value store persisted as a JSON object.
package kvstore

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
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
	path string
	data map[string]string
}

// Open loads the store at path; a missing file is an empty store.
func Open(path string) (*Store, error) {
	data, err := readData(path)
	if err != nil {
		return nil, err
	}
	return &Store{path: path, data: data}, nil
}

// readData parses the store file at path with the rules Open uses.
func readData(path string) (map[string]string, error) {
	data := map[string]string{}
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return data, nil
	}
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return data, nil
	}
	if err := json.Unmarshal(b, &data); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if data == nil {
		data = map[string]string{}
	}
	return data, nil
}

// reload replaces s.data with the current contents of the store file.
func (s *Store) reload() error {
	data, err := readData(s.path)
	if err != nil {
		return err
	}
	s.data = data
	return nil
}

// Get returns the value for key.
func (s *Store) Get(key string) (string, bool) {
	v, ok := s.data[key]
	return v, ok
}

// Set stores value under key and persists atomically. The store file is
// re-read under the directory lock, so writes made through other handles since
// Open are kept.
func (s *Store) Set(key, value string) error {
	unlock, err := lockDir(filepath.Dir(s.path))
	if err != nil {
		return err
	}
	defer unlock()
	if err := s.reload(); err != nil {
		return err
	}
	old, had := s.data[key]
	s.data[key] = value
	if err := s.save(); err != nil {
		if had {
			s.data[key] = old
		} else {
			delete(s.data, key)
		}
		return err
	}
	return nil
}

// Delete removes key, returning ErrNotFound if absent. Like Set, it works on
// the store file as it is under the directory lock, so a key another handle has
// already removed is not written back.
func (s *Store) Delete(key string) error {
	unlock, err := lockDir(filepath.Dir(s.path))
	if err != nil {
		return err
	}
	defer unlock()
	if err := s.reload(); err != nil {
		return err
	}
	old, ok := s.data[key]
	if !ok {
		return ErrNotFound
	}
	delete(s.data, key)
	if err := s.save(); err != nil {
		s.data[key] = old
		return err
	}
	return nil
}

// List returns all pairs sorted by key.
func (s *Store) List() []Pair {
	out := make([]Pair, 0, len(s.data))
	for k, v := range s.data {
		out = append(out, Pair{k, v})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Key < out[j].Key
	})
	return out
}

// save writes to a temp file in the same directory, then renames over the target.
// An existing store file keeps its permission bits; a new one is created with 0600.
func (s *Store) save() error {
	b, err := json.Marshal(s.data)
	if err != nil {
		return err
	}
	mode := os.FileMode(0o600)
	fi, err := os.Stat(s.path)
	switch {
	case err == nil:
		mode = fi.Mode().Perm()
	case !errors.Is(err, os.ErrNotExist):
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(s.path), ".kv-*.tmp")
	if err != nil {
		return err
	}
	name := tmp.Name()
	if err := tmp.Chmod(mode); err != nil {
		tmp.Close()
		os.Remove(name)
		return err
	}
	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		os.Remove(name)
		return err
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
