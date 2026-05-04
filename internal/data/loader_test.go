package data

import (
	"testing"
	"testing/fstest"
)

func TestLoaderLoadAllValid(t *testing.T) {
	fsys := fstest.MapFS{
		"biomes.yaml": &fstest.MapFile{
			Data: []byte(`kind: biomes
data:
  - id: plains
    name: Llanuras
    temperature: 0.6
    humidity: 0.5
    color: "#90EE90"
`),
		},
	}

	loader := NewLoader(fsys)
	registry := NewRegistry()
	err := loader.LoadAll(".", registry)
	if err != nil {
		t.Fatalf("LoadAll error: %v", err)
	}

	biomes, ok := Get[[]any](registry, "biomes")
	if !ok {
		t.Fatal("expected biomes to be registered")
	}
	if len(biomes) != 1 {
		t.Fatalf("expected 1 biome, got %d", len(biomes))
	}
	b, ok := biomes[0].(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any, got %T", biomes[0])
	}
	if b["id"] != "plains" {
		t.Errorf("id = %v, want plains", b["id"])
	}
}

func TestLoaderLoadAllInvalidYAML(t *testing.T) {
	fsys := fstest.MapFS{
		"bad.yaml": &fstest.MapFile{Data: []byte(`{ not yaml `)},
	}

	loader := NewLoader(fsys)
	registry := NewRegistry()
	err := loader.LoadAll(".", registry)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestLoaderLoadAllEmptyDir(t *testing.T) {
	fsys := fstest.MapFS{}

	loader := NewLoader(fsys)
	registry := NewRegistry()
	err := loader.LoadAll(".", registry)
	if err != nil {
		t.Fatalf("LoadAll error on empty dir: %v", err)
	}
}
