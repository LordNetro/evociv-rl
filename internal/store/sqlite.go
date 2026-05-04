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
	return db.Ping()
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
