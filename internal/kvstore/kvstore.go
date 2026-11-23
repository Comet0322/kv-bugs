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
	v, ok := s.data[key]
	return v, ok
}

// Set stores value under key and persists atomically. The directory lock is
// held across a reload from disk, so a concurrent writer's keys are kept.
func (s *Store) Set(key, value string) error {
	unlock, err := lockDir(filepath.Dir(s.path))
	if err != nil {
		return err
	}
	defer unlock()
	data, err := reload(s.path)
	if err != nil {
		return err
	}
	data[key] = value
	if err := (&Store{path: s.path, data: data}).save(); err != nil {
		return err
	}
	s.data = data
	return nil
}

// reload reads the store file from disk the way Open does and returns its
// map. A missing or empty file is an empty map; a parse error returns an error
// and the caller writes nothing.
func reload(path string) (map[string]string, error) {
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return map[string]string{}, nil
	}
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return map[string]string{}, nil
	}
	data := map[string]string{}
	if err := json.Unmarshal(b, &data); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if data == nil {
		data = map[string]string{}
	}
	return data, nil
}

// Delete removes key, returning ErrNotFound if absent. Like Set, it reloads
// the file from disk under the directory lock before it writes.
func (s *Store) Delete(key string) error {
	unlock, err := lockDir(filepath.Dir(s.path))
	if err != nil {
		return err
	}
	defer unlock()
	data, err := reload(s.path)
	if err != nil {
		return err
	}
	if _, ok := data[key]; !ok {
		return ErrNotFound
	}
	delete(data, key)
	if err := (&Store{path: s.path, data: data}).save(); err != nil {
		return err
	}
	s.data = data
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
	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		os.Remove(name)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(name)
		return err
	}
	// Keep the permission bits of an existing store file. A new file keeps
	// CreateTemp's 0600. os.Chmod, not the creation mode, so umask does not apply.
	if fi, err := os.Stat(s.path); err == nil {
		if err := os.Chmod(name, fi.Mode().Perm()); err != nil {
			os.Remove(name)
			return err
		}
	}
	if err := os.Rename(name, s.path); err != nil {
		os.Remove(name)
		return err
	}
	return nil
}
