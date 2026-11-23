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
	data, err := load(path)
	if err != nil {
		return nil, err
	}
	return &Store{path: path, data: data}, nil
}

// load reads the store file at path; a missing or empty file is an empty map.
func load(path string) (map[string]string, error) {
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

// Get returns the value for key.
func (s *Store) Get(key string) (string, bool) {
	v, ok := s.data[key]
	return v, ok
}

// Set stores value under key and persists atomically.
func (s *Store) Set(key, value string) error {
	return s.update(func(data map[string]string) error {
		data[key] = value
		return nil
	})
}

// Delete removes key, returning ErrNotFound if absent.
func (s *Store) Delete(key string) error {
	return s.update(func(data map[string]string) error {
		if _, ok := data[key]; !ok {
			return ErrNotFound
		}
		delete(data, key)
		return nil
	})
}

// update applies mutate to the store under an exclusive lock on its
// directory: it re-reads the file so changes made by other processes since
// Open are kept, mutates that fresh data and saves it before unlocking.
func (s *Store) update(mutate func(map[string]string) error) error {
	unlock, err := lockDir(filepath.Dir(s.path))
	if err != nil {
		return err
	}
	defer unlock()
	data, err := load(s.path)
	if err != nil {
		return err
	}
	if err := mutate(data); err != nil {
		return err
	}
	old := s.data
	s.data = data
	if err := s.save(); err != nil {
		s.data = old
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
func (s *Store) save() error {
	b, err := json.Marshal(s.data)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(s.path), ".kv-*.tmp")
	if err != nil {
		return err
	}
	name := tmp.Name()
	if err := keepMode(tmp, s.path); err != nil {
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

// keepMode gives tmp the permission bits of the existing file at path, so the
// rename does not change them; a path that does not exist yet is left alone.
func keepMode(tmp *os.File, path string) error {
	fi, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	return tmp.Chmod(fi.Mode().Perm())
}
