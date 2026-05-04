package data

import (
	"fmt"
	"io/fs"
	"path"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Loader reads YAML data files into a Registry.
type Loader struct {
	fsys fs.FS
}

// NewLoader creates a Loader using the provided filesystem.
func NewLoader(fsys fs.FS) *Loader {
	return &Loader{fsys: fsys}
}

// LoadAll reads all .yaml files in the given directory and registers their contents.
func (l *Loader) LoadAll(dir string, registry *Registry) error {
	entries, err := fs.ReadDir(l.fsys, dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}
		data, err := fs.ReadFile(l.fsys, path.Join(dir, entry.Name()))
		if err != nil {
			return err
		}
		var doc struct {
			Kind string `yaml:"kind"`
			Data any    `yaml:"data"`
		}
		if err := yaml.Unmarshal(data, &doc); err != nil {
			return fmt.Errorf("parse %s: %w", entry.Name(), err)
		}
		if doc.Kind != "" {
			registry.Register(doc.Kind, doc.Data)
		}
	}
	return nil
}
