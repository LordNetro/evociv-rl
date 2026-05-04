package store

import (
	"testing"
)

func TestSQLiteStoreOpenClose(t *testing.T) {
	path := t.TempDir() + "/test.db"
	s := NewSQLiteStore()

	if err := s.Open(path); err != nil {
		t.Fatalf("Open error: %v", err)
	}
	if s.Path() != path {
		t.Errorf("Path = %q, want %q", s.Path(), path)
	}
	if err := s.Health(); err != nil {
		t.Errorf("Health error: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("Close error: %v", err)
	}
}

func TestSQLiteStoreOpenInvalidPath(t *testing.T) {
	s := NewSQLiteStore()
	// An invalid path should still open a DB in memory or fail gracefully.
	// For modernc.org/sqlite, paths with parent dirs that don't exist may fail.
	err := s.Open("/nonexistent/path/to/db.sqlite")
	if err == nil {
		t.Error("expected error for invalid path")
	}
}
