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
	if seed != 7 || width != 8 || height != 9 {
		t.Errorf("data mismatch: %d, %d, %d", seed, width, height)
	}
	if offset != 999 {
		t.Errorf("default npcSeedOffset = %d, want 999", offset)
	}
}

func TestSaveAndLoadQTable(t *testing.T) {
	path := t.TempDir() + "/test.db"
	s := NewSQLiteStore()
	if err := s.Open(path); err != nil {
		t.Fatalf("Open error: %v", err)
	}
	defer s.Close()

	data := map[string]map[string]float64{
		"hunger|plains|day": {
			"harvest": 0.5,
			"rest":    0.1,
		},
		"fatigue|forest|day": {
			"rest": 0.8,
		},
	}

	if err := s.SaveQTable(1, data); err != nil {
		t.Fatalf("SaveQTable error: %v", err)
	}

	loaded, err := s.LoadQTable(1)
	if err != nil {
		t.Fatalf("LoadQTable error: %v", err)
	}

	if len(loaded) != 2 {
		t.Fatalf("expected 2 states, got %d", len(loaded))
	}
	if loaded["hunger|plains|day"]["harvest"] != 0.5 {
		t.Errorf("harvest Q = %f, want 0.5", loaded["hunger|plains|day"]["harvest"])
	}
	if loaded["fatigue|forest|day"]["rest"] != 0.8 {
		t.Errorf("rest Q = %f, want 0.8", loaded["fatigue|forest|day"]["rest"])
	}
}

func TestLoadEmptyQTable(t *testing.T) {
	path := t.TempDir() + "/test.db"
	s := NewSQLiteStore()
	if err := s.Open(path); err != nil {
		t.Fatalf("Open error: %v", err)
	}
	defer s.Close()

	loaded, err := s.LoadQTable(1)
	if err != nil {
		t.Fatalf("LoadQTable error: %v", err)
	}
	if len(loaded) != 0 {
		t.Errorf("expected empty map, got %d entries", len(loaded))
	}
}

func TestQTablePersistenceAcrossOpenClose(t *testing.T) {
	path := t.TempDir() + "/test.db"

	s1 := NewSQLiteStore()
	if err := s1.Open(path); err != nil {
		t.Fatalf("Open error: %v", err)
	}
	data := map[string]map[string]float64{
		"state1": {"action1": 1.23},
	}
	if err := s1.SaveQTable(42, data); err != nil {
		t.Fatalf("SaveQTable error: %v", err)
	}
	if err := s1.Close(); err != nil {
		t.Fatalf("Close error: %v", err)
	}

	s2 := NewSQLiteStore()
	if err := s2.Open(path); err != nil {
		t.Fatalf("second Open error: %v", err)
	}
	defer s2.Close()

	loaded, err := s2.LoadQTable(42)
	if err != nil {
		t.Fatalf("LoadQTable error: %v", err)
	}
	if loaded["state1"]["action1"] != 1.23 {
		t.Errorf("Q-value = %f, want 1.23", loaded["state1"]["action1"])
	}
}
