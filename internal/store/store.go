package store

// Store defines the interface for persistent storage backends.
type Store interface {
	Open(path string) error
	Close() error
	Health() error
	Path() string
}
