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

func TestSaveAndLoadWorld(t *testing.T) {
	path := t.TempDir() + "/test.db"
	s := NewSQLiteStore()
	if err := s.Open(path); err != nil {
		t.Fatalf("Open error: %v", err)
	}
	defer s.Close()

	if err := s.SaveWorld(42, 64, 64); err != nil {
		t.Fatalf("SaveWorld error: %v", err)
	}

	seed, width, height, err := s.LoadLatestWorld()
	if err != nil {
		t.Fatalf("LoadLatestWorld error: %v", err)
	}
	if seed != 42 {
		t.Errorf("seed = %d, want 42", seed)
	}
	if width != 64 {
		t.Errorf("width = %d, want 64", width)
	}
	if height != 64 {
		t.Errorf("height = %d, want 64", height)
	}
}

func TestLoadEmptyStore(t *testing.T) {
	path := t.TempDir() + "/test.db"
	s := NewSQLiteStore()
	if err := s.Open(path); err != nil {
		t.Fatalf("Open error: %v", err)
	}
	defer s.Close()

	_, _, _, err := s.LoadLatestWorld()
	if err == nil {
		t.Error("expected error for empty store")
	}
}

func TestSaveMultipleLoadLatest(t *testing.T) {
	path := t.TempDir() + "/test.db"
	s := NewSQLiteStore()
	if err := s.Open(path); err != nil {
		t.Fatalf("Open error: %v", err)
	}
	defer s.Close()

	seeds := []int64{10, 20, 30}
	for _, seed := range seeds {
		if err := s.SaveWorld(seed, 32, 32); err != nil {
			t.Fatalf("SaveWorld error: %v", err)
		}
	}

	seed, width, height, err := s.LoadLatestWorld()
	if err != nil {
		t.Fatalf("LoadLatestWorld error: %v", err)
	}
	if seed != 30 {
		t.Errorf("seed = %d, want 30 (latest)", seed)
	}
	if width != 32 {
		t.Errorf("width = %d, want 32", width)
	}
	if height != 32 {
		t.Errorf("height = %d, want 32", height)
	}
}

func TestIdempotentMigration(t *testing.T) {
	path := t.TempDir() + "/test.db"
	s1 := NewSQLiteStore()
	if err := s1.Open(path); err != nil {
		t.Fatalf("first Open error: %v", err)
	}
	if err := s1.SaveWorld(1, 16, 16); err != nil {
		t.Fatalf("SaveWorld error: %v", err)
	}
	if err := s1.Close(); err != nil {
		t.Fatalf("Close error: %v", err)
	}

	// Re-open same database
	s2 := NewSQLiteStore()
	if err := s2.Open(path); err != nil {
		t.Fatalf("second Open error: %v", err)
	}
	defer s2.Close()

	seed, width, height, err := s2.LoadLatestWorld()
	if err != nil {
		t.Fatalf("LoadLatestWorld error: %v", err)
	}
	if seed != 1 {
		t.Errorf("seed = %d, want 1", seed)
	}
	if width != 16 || height != 16 {
		t.Errorf("size = %dx%d, want 16x16", width, height)
	}
}
