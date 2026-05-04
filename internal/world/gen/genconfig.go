package gen

import (
	"fmt"
	"io/fs"

	"gopkg.in/yaml.v3"
)

// GenConfig holds world generation parameters.
type GenConfig struct {
	Seed       int64   `yaml:"seed"`
	Width      int     `yaml:"width"`
	Height     int     `yaml:"height"`
	Octaves    int     `yaml:"octaves"`
	Lacunarity float64 `yaml:"lacunarity"`
	Gain       float64 `yaml:"gain"`
	Scale      float64 `yaml:"scale"`
}

// Validate checks that the config has valid dimensions.
func (c GenConfig) Validate() error {
	if c.Width <= 0 {
		return fmt.Errorf("width must be > 0, got %d", c.Width)
	}
	if c.Height <= 0 {
		return fmt.Errorf("height must be > 0, got %d", c.Height)
	}
	return nil
}

// LoadGenConfig loads a generation config from a YAML file.
func LoadGenConfig(path string, fsys fs.FS) (GenConfig, error) {
	data, err := fs.ReadFile(fsys, path)
	if err != nil {
		return GenConfig{}, fmt.Errorf("read file: %w", err)
	}

	var doc struct {
		Kind string    `yaml:"kind"`
		Data GenConfig `yaml:"data"`
	}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return GenConfig{}, fmt.Errorf("parse yaml: %w", err)
	}
	if doc.Kind != "" && doc.Kind != "gen-config" {
		return GenConfig{}, fmt.Errorf("expected kind gen-config, got %q", doc.Kind)
	}

	cfg := doc.Data
	if cfg.Octaves == 0 {
		cfg.Octaves = 6
	}
	if cfg.Lacunarity == 0 {
		cfg.Lacunarity = 2.0
	}
	if cfg.Gain == 0 {
		cfg.Gain = 0.5
	}
	if cfg.Scale == 0 {
		cfg.Scale = 100.0
	}

	return cfg, nil
}
