package store

// Store defines the interface for persistent storage backends.
type Store interface {
	Open(path string) error
	Close() error
	Health() error
	Path() string
	SaveWorld(seed int64, width, height int, npcSeedOffset int64) error
	LoadLatestWorld() (seed int64, width, height int, npcSeedOffset int64, err error)
}
