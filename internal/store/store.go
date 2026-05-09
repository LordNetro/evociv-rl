package store

// Store defines the interface for persistent storage backends.
type Store interface {
	Open(path string) error
	Close() error
	Health() error
	Path() string
	SaveWorld(seed int64, width, height int, npcSeedOffset int64) error
	LoadLatestWorld() (seed int64, width, height int, npcSeedOffset int64, err error)
	SaveQTable(npcID int, qTable map[string]map[string]float64) error
	LoadQTable(npcID int) (map[string]map[string]float64, error)
}
