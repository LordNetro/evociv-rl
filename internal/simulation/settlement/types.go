package settlement

// SettlementDef defines a settlement type loaded from YAML.
type SettlementDef struct {
	ID          string   `yaml:"id"`
	Name        string   `yaml:"name"`
	Symbol      string   `yaml:"symbol"`
	Color       string   `yaml:"color"`
	Radius      int      `yaml:"radius"`
	Biomes      []string `yaml:"biomes"`
	Buildings   []string `yaml:"buildings"`
	SpawnWeight float64  `yaml:"spawn_weight"`
}

// BuildingDef defines a building type loaded from YAML.
type BuildingDef struct {
	ID         string             `yaml:"id"`
	Name       string             `yaml:"name"`
	Role       string             `yaml:"role"`
	Produces   map[string]float64 `yaml:"produces"`
	Consumes   map[string]float64 `yaml:"consumes"`
	MaxWorkers int                `yaml:"max_workers"`
}

// GrowthThreshold defines resource requirements for settlement level-up.
type GrowthThreshold struct {
	Level        int     `yaml:"level"`
	Food         float64 `yaml:"food"`
	Tools        float64 `yaml:"tools"`
	Gold         float64 `yaml:"gold"`
	NewRadius    int     `yaml:"new_radius"`
	NewBuildings []string `yaml:"new_buildings"`
}

// SettlementRenderInfo carries everything the TUI needs to draw a settlement overlay.
type SettlementRenderInfo struct {
	Entity         int
	Symbol         rune
	Color          string
	Name           string
	WorldX, WorldY int
	Population     int
	Level          int
	Food           float64
	Gold           float64
	Tools          float64
	HasResources   bool
}

// Name pools for procedural settlement naming.
var (
	settlementNamePrefixes = []string{
		"Norte", "Sur", "Este", "Oeste", "Alto", "Bajo",
		"Nuevo", "Viejo", "Valle", "Monte", "Rio", "Lago",
	}
	settlementNameSuffixes = []string{
		"del Valle", "de la Colina", "del Bosque", "del Rio",
		"Dorado", "Plateado", "Verde", "del Sol",
	}
)
