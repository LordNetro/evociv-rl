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

	if err := s.SaveWorld(42, 64, 64, 999); err != nil {
		t.Fatalf("SaveWorld error: %v", err)
	}

	seed, width, height, offset, err := s.LoadLatestWorld()
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
	if offset != 999 {
		t.Errorf("npcSeedOffset = %d, want 999", offset)
	}
}

func TestLoadEmptyStore(t *testing.T) {
	path := t.TempDir() + "/test.db"
	s := NewSQLiteStore()
	if err := s.Open(path); err != nil {
		t.Fatalf("Open error: %v", err)
	}
	defer s.Close()

	_, _, _, _, err := s.LoadLatestWorld()
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
		if err := s.SaveWorld(seed, 32, 32, 999); err != nil {
			t.Fatalf("SaveWorld error: %v", err)
		}
	}

	seed, width, height, offset, err := s.LoadLatestWorld()
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
	if offset != 999 {
		t.Errorf("npcSeedOffset = %d, want 999", offset)
	}
}

func TestIdempotentMigration(t *testing.T) {
	path := t.TempDir() + "/test.db"
	s1 := NewSQLiteStore()
	if err := s1.Open(path); err != nil {
		t.Fatalf("first Open error: %v", err)
	}
	if err := s1.SaveWorld(1, 16, 16, 999); err != nil {
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

	seed, width, height, offset, err := s2.LoadLatestWorld()
	if err != nil {
		t.Fatalf("LoadLatestWorld error: %v", err)
	}
	if seed != 1 {
		t.Errorf("seed = %d, want 1", seed)
	}
	if width != 16 || height != 16 {
		t.Errorf("size = %dx%d, want 16x16", width, height)
	}
	if offset != 999 {
		t.Errorf("npcSeedOffset = %d, want 999", offset)
	}
}

func TestSaveAndLoadWorldCustomOffset(t *testing.T) {
	path := t.TempDir() + "/test.db"
	s := NewSQLiteStore()
	if err := s.Open(path); err != nil {
		t.Fatalf("Open error: %v", err)
	}
	defer s.Close()

	if err := s.SaveWorld(42, 64, 64, 1234); err != nil {
		t.Fatalf("SaveWorld error: %v", err)
	}

	_, _, _, offset, err := s.LoadLatestWorld()
	if err != nil {
		t.Fatalf("LoadLatestWorld error: %v", err)
	}
	if offset != 1234 {
		t.Errorf("npcSeedOffset = %d, want 1234", offset)
	}
}

func TestDeterministicRegenerationAfterSaveLoad(t *testing.T) {
	path := t.TempDir() + "/test.db"
	s := NewSQLiteStore()
	if err := s.Open(path); err != nil {
		t.Fatalf("Open error: %v", err)
	}
	defer s.Close()

	seed := int64(12345)
	width := 64
	height := 64
	offset := int64(888)

	if err := s.SaveWorld(seed, width, height, offset); err != nil {
		t.Fatalf("SaveWorld error: %v", err)
	}

	loadedSeed, loadedWidth, loadedHeight, loadedOffset, err := s.LoadLatestWorld()
	if err != nil {
		t.Fatalf("LoadLatestWorld error: %v", err)
	}

	if loadedSeed != seed {
		t.Errorf("seed = %d, want %d", loadedSeed, seed)
	}
	if loadedWidth != width {
		t.Errorf("width = %d, want %d", loadedWidth, width)
	}
	if loadedHeight != height {
		t.Errorf("height = %d, want %d", loadedHeight, height)
	}
	if loadedOffset != offset {
		t.Errorf("offset = %d, want %d", loadedOffset, offset)
	}

	// Verify that seed+offset can be used to regenerate the same world
	if loadedSeed+loadedOffset != seed+offset {
		t.Error("seed+offset mismatch prevents deterministic regeneration")
	}
}

func TestMigrationAddsColumnToExistingDB(t *testing.T) {
	path := t.TempDir() + "/legacy.db"

	// Manually create an old-style worlds table without npc_seed_offset
	s1 := NewSQLiteStore()
	if err := s1.Open(path); err != nil {
		t.Fatalf("Open error: %v", err)
	}
	_, err := s1.db.Exec(`CREATE TABLE IF NOT EXISTS worlds (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		seed INTEGER NOT NULL,
		width INTEGER NOT NULL,
		height INTEGER NOT NULL,
		created_at TEXT DEFAULT (datetime('now'))
	)`)
	if err != nil {
		t.Fatalf("create legacy table: %v", err)
	}
	_, err = s1.db.Exec("INSERT INTO worlds (seed, width, height) VALUES (7, 8, 9)")
	if err != nil {
		t.Fatalf("insert legacy row: %v", err)
	}
	if err := s1.Close(); err != nil {
		t.Fatalf("Close error: %v", err)
	}

	// Re-open with new code — migration should add column
	s2 := NewSQLiteStore()
	if err := s2.Open(path); err != nil {
		t.Fatalf("second Open error: %v", err)
	}
	defer s2.Close()

	seed, width, height, offset, err := s2.LoadLatestWorld()
	if err != nil {
		t.Fatalf("LoadLatestWorld error: %v", err)
	}
	if seed != 7 || width != 8 || height != 9 {
		t.Errorf("data mismatch: %d, %d, %d", seed, width, height)
	}
	if offset != 999 {
		t.Errorf("default npcSeedOffset = %d, want 999", offset)
	}
}
