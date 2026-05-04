package store

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// SQLiteStore implements Store using SQLite.
type SQLiteStore struct {
	db   *sql.DB
	path string
}

// NewSQLiteStore creates a new SQLiteStore.
func NewSQLiteStore() *SQLiteStore {
	return &SQLiteStore{}
}

// Open opens the SQLite database at the given path.
func (s *SQLiteStore) Open(path string) error {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return fmt.Errorf("open sqlite: %w", err)
	}
	s.db = db
	s.path = path
	if err := db.Ping(); err != nil {
		return err
	}
	return s.migrate()
}

func (s *SQLiteStore) migrate() error {
	_, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS worlds (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		seed INTEGER NOT NULL,
		width INTEGER NOT NULL,
		height INTEGER NOT NULL,
		created_at TEXT DEFAULT (datetime('now'))
	)`)
	if err != nil {
		return fmt.Errorf("migrate worlds table: %w", err)
	}
	return nil
}

// SaveWorld stores a world generation record.
func (s *SQLiteStore) SaveWorld(seed int64, width, height int) error {
	_, err := s.db.Exec(
		"INSERT INTO worlds (seed, width, height) VALUES (?, ?, ?)",
		seed, width, height,
	)
	if err != nil {
		return fmt.Errorf("save world: %w", err)
	}
	return nil
}

// LoadLatestWorld retrieves the most recently saved world.
func (s *SQLiteStore) LoadLatestWorld() (seed int64, width, height int, err error) {
	row := s.db.QueryRow("SELECT seed, width, height FROM worlds ORDER BY id DESC LIMIT 1")
	if err := row.Scan(&seed, &width, &height); err != nil {
		return 0, 0, 0, fmt.Errorf("load latest world: %w", err)
	}
	return seed, width, height, nil
}

// Close closes the database connection.
func (s *SQLiteStore) Close() error {
	if s.db == nil {
		return nil
	}
	return s.db.Close()
}

// Health checks the database connection.
func (s *SQLiteStore) Health() error {
	if s.db == nil {
		return fmt.Errorf("database not open")
	}
	return s.db.Ping()
}

// Path returns the database file path.
func (s *SQLiteStore) Path() string {
	return s.path
}
