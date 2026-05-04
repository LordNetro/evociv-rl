package ecs

import "testing"

func TestComponentStoreSetGet(t *testing.T) {
	store := NewComponentStore[Position]()
	e := Entity(1)
	pos := Position{X: 10.0, Y: 20.0, Z: 0}

	store.Set(e, pos)
	got, ok := store.Get(e)
	if !ok {
		t.Fatal("expected component to exist")
	}
	if got != pos {
		t.Errorf("got %+v, want %+v", got, pos)
	}
}

func TestComponentStoreGetMissing(t *testing.T) {
	store := NewComponentStore[Position]()
	_, ok := store.Get(Entity(99))
	if ok {
		t.Error("expected Get on missing entity to return false")
	}
}

func TestComponentStoreDelete(t *testing.T) {
	store := NewComponentStore[Name]()
	e := Entity(1)
	store.Set(e, Name{Name: "Alice"})
	store.Delete(e)

	if store.Has(e) {
		t.Error("expected entity to be deleted")
	}
}

func TestComponentStoreHas(t *testing.T) {
	store := NewComponentStore[Tags]()
	e := Entity(1)
	if store.Has(e) {
		t.Error("expected Has to be false for unset entity")
	}
	store.Set(e, Tags{Tags: []string{"a"}})
	if !store.Has(e) {
		t.Error("expected Has to be true after Set")
	}
}

func TestComponentStoreLen(t *testing.T) {
	store := NewComponentStore[Position]()
	if store.Len() != 0 {
		t.Errorf("Len = %d, want 0", store.Len())
	}
	store.Set(Entity(1), Position{X: 1, Y: 1})
	store.Set(Entity(2), Position{X: 2, Y: 2})
	if store.Len() != 2 {
		t.Errorf("Len = %d, want 2", store.Len())
	}
}

func TestComponentStoreAll(t *testing.T) {
	store := NewComponentStore[Name]()
	store.Set(Entity(1), Name{Name: "A"})
	store.Set(Entity(2), Name{Name: "B"})

	all := store.All()
	if len(all) != 2 {
		t.Fatalf("All returned %d items, want 2", len(all))
	}

	found := make(map[Entity]Name)
	for e, n := range all {
		found[e] = n
	}
	if len(found) != 2 {
		t.Error("All map size incorrect")
	}
	if found[Entity(1)].Name != "A" || found[Entity(2)].Name != "B" {
		t.Errorf("All returned wrong values: %+v", found)
	}
}
