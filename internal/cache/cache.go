// Package cache provides client-side disk caching for lokai.
//
// All data is stored under ~/.cache/lokai/ (or platform equivalent).
// Each cache entry has a TTL; stale entries are ignored and overwritten
// on the next write. Nothing is ever sent to a remote server.
package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Store manages a JSON-based disk cache under a single directory.
type Store struct {
	dir string
}

// New creates (or opens) a cache store in the default user cache directory.
func New() (*Store, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		base = os.TempDir()
	}
	dir := filepath.Join(base, "lokai")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &Store{dir: dir}, nil
}

// Dir returns the cache directory path.
func (s *Store) Dir() string { return s.dir }

// entry is the on-disk envelope for every cached value.
type entry struct {
	Data      json.RawMessage `json:"data"`
	CachedAt  time.Time       `json:"cached_at"`
	TTLMillis int64           `json:"ttl_ms"`
}

// Get loads a cached value into dst. Returns false if the key is missing
// or the entry has expired.
func (s *Store) Get(key string, dst any) bool {
	path := s.path(key)
	raw, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	var e entry
	if err := json.Unmarshal(raw, &e); err != nil {
		return false
	}
	ttl := time.Duration(e.TTLMillis) * time.Millisecond
	if time.Since(e.CachedAt) > ttl {
		return false // expired
	}
	return json.Unmarshal(e.Data, dst) == nil
}

// Set writes a value to the cache with the given TTL.
func (s *Store) Set(key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	e := entry{
		Data:      data,
		CachedAt:  time.Now(),
		TTLMillis: ttl.Milliseconds(),
	}
	raw, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path(key), raw, 0o644)
}

// Delete removes a cached key.
func (s *Store) Delete(key string) {
	_ = os.Remove(s.path(key))
}

// Clear removes all cached data.
func (s *Store) Clear() error {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if !e.IsDir() {
			_ = os.Remove(filepath.Join(s.dir, e.Name()))
		}
	}
	return nil
}

func (s *Store) path(key string) string {
	return filepath.Join(s.dir, key+".json")
}
