package store

// Store defines the interface for persistent storage backends.
type Store interface {
	Open(path string) error
	Close() error
	Health() error
	Path() string
	SaveWorld(seed int64, width, height int) error
	LoadLatestWorld() (seed int64, width, height int, err error)
}
