package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	return &Store{dir: dir}
}

func TestSetAndGet(t *testing.T) {
	s := newTestStore(t)

	type payload struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}

	in := payload{Name: "hello", Count: 42}
	if err := s.Set("test-key", in, time.Hour); err != nil {
		t.Fatalf("Set: %v", err)
	}

	var out payload
	if !s.Get("test-key", &out) {
		t.Fatal("Get returned false for existing key")
	}
	if out.Name != in.Name || out.Count != in.Count {
		t.Errorf("Get = %+v, want %+v", out, in)
	}
}

func TestGetMissingKey(t *testing.T) {
	s := newTestStore(t)

	var out string
	if s.Get("nonexistent", &out) {
		t.Error("Get should return false for missing key")
	}
}

func TestGetExpired(t *testing.T) {
	s := newTestStore(t)

	if err := s.Set("short-lived", "value", time.Millisecond); err != nil {
		t.Fatalf("Set: %v", err)
	}

	time.Sleep(5 * time.Millisecond)

	var out string
	if s.Get("short-lived", &out) {
		t.Error("Get should return false for expired entry")
	}
}

func TestGetCorruptedJSON(t *testing.T) {
	s := newTestStore(t)

	// Write garbage to the cache file.
	path := filepath.Join(s.dir, "corrupt.json")
	if err := os.WriteFile(path, []byte("not json"), 0o644); err != nil {
		t.Fatal(err)
	}

	var out string
	if s.Get("corrupt", &out) {
		t.Error("Get should return false for corrupted JSON")
	}
}

func TestDelete(t *testing.T) {
	s := newTestStore(t)

	if err := s.Set("del-me", "value", time.Hour); err != nil {
		t.Fatalf("Set: %v", err)
	}

	s.Delete("del-me")

	var out string
	if s.Get("del-me", &out) {
		t.Error("Get should return false after Delete")
	}
}

func TestClear(t *testing.T) {
	s := newTestStore(t)

	for _, key := range []string{"a", "b", "c"} {
		if err := s.Set(key, key, time.Hour); err != nil {
			t.Fatalf("Set(%s): %v", key, err)
		}
	}

	if err := s.Clear(); err != nil {
		t.Fatalf("Clear: %v", err)
	}

	for _, key := range []string{"a", "b", "c"} {
		var out string
		if s.Get(key, &out) {
			t.Errorf("Get(%s) should return false after Clear", key)
		}
	}
}

func TestDir(t *testing.T) {
	s := newTestStore(t)
	if s.Dir() == "" {
		t.Error("Dir should not be empty")
	}
}

func TestSetOverwrite(t *testing.T) {
	s := newTestStore(t)

	if err := s.Set("key", "old", time.Hour); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if err := s.Set("key", "new", time.Hour); err != nil {
		t.Fatalf("Set overwrite: %v", err)
	}

	var out string
	if !s.Get("key", &out) {
		t.Fatal("Get returned false")
	}
	if out != "new" {
		t.Errorf("Get = %q, want %q", out, "new")
	}
}
