package data

import "testing"

func TestRegistryRegisterGet(t *testing.T) {
	r := NewRegistry()
	r.Register("biomes", []string{"plains", "forest"})

	got, ok := Get[[]string](r, "biomes")
	if !ok {
		t.Fatal("expected key to exist")
	}
	if len(got) != 2 || got[0] != "plains" || got[1] != "forest" {
		t.Errorf("got %v, want [plains forest]", got)
	}
}

func TestRegistryGetMissing(t *testing.T) {
	r := NewRegistry()
	_, ok := Get[int](r, "missing")
	if ok {
		t.Error("expected Get on missing key to return false")
	}
}

func TestRegistryAll(t *testing.T) {
	r := NewRegistry()
	r.Register("a", 1)
	r.Register("b", 2)
	r.Register("c", "not an int")

	got := All[int](r)
	if len(got) != 2 {
		t.Fatalf("All returned %d items, want 2", len(got))
	}
	found := make(map[int]bool)
	for _, v := range got {
		found[v] = true
	}
	if !found[1] || !found[2] {
		t.Errorf("All returned wrong values: %v", got)
	}
}

func TestRegistryTypes(t *testing.T) {
	r := NewRegistry()
	r.Register("x", 1)
	r.Register("y", "hello")

	types := r.Types()
	if len(types) != 2 {
		t.Fatalf("Types returned %d items, want 2", len(types))
	}
	found := make(map[string]bool)
	for _, ty := range types {
		found[ty] = true
	}
	if !found["int"] || !found["string"] {
		t.Errorf("Types returned wrong values: %v", types)
	}
}
