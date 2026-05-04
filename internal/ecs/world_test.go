package ecs

import "testing"

func TestNewWorld(t *testing.T) {
	w := NewWorld()
	if w == nil {
		t.Fatal("NewWorld returned nil")
	}
}

func TestWorldNewEntity(t *testing.T) {
	w := NewWorld()
	e1 := w.NewEntity()
	e2 := w.NewEntity()
	if e1 == EntityInvalid {
		t.Error("NewEntity returned invalid entity")
	}
	if e1 == e2 {
		t.Error("NewEntity returned duplicate IDs")
	}
}

func TestWorldRegisterAndGetStore(t *testing.T) {
	w := NewWorld()
	store := NewComponentStore[Position]()
	id := NewComponentID("position")

	RegisterComponentStore(w, id, store)
	got := w.GetStore(id)
	if got == nil {
		t.Fatal("GetStore returned nil")
	}
	if got != store {
		t.Error("GetStore returned different store")
	}
}

func TestWorldAddGetComponent(t *testing.T) {
	w := NewWorld()
	posID := NewComponentID("position")
	RegisterComponentStore(w, posID, NewComponentStore[Position]())

	e := w.NewEntity()
	pos := Position{X: 5.0, Y: 10.0}
	AddComponent(w, e, pos)

	got, ok := GetComponent[Position](w, e)
	if !ok {
		t.Fatal("expected component to exist")
	}
	if got != pos {
		t.Errorf("got %+v, want %+v", got, pos)
	}
}

func TestWorldGetComponentMissing(t *testing.T) {
	w := NewWorld()
	posID := NewComponentID("position")
	RegisterComponentStore(w, posID, NewComponentStore[Position]())

	_, ok := GetComponent[Position](w, Entity(999))
	if ok {
		t.Error("expected GetComponent to return false for missing entity")
	}
}

func TestWorldRemoveEntity(t *testing.T) {
	w := NewWorld()
	posID := NewComponentID("position")
	nameID := NewComponentID("name")
	RegisterComponentStore(w, posID, NewComponentStore[Position]())
	RegisterComponentStore(w, nameID, NewComponentStore[Name]())

	e := w.NewEntity()
	AddComponent(w, e, Position{X: 1, Y: 2})
	AddComponent(w, e, Name{Name: "E1"})
	w.RemoveEntity(e)

	if _, ok := GetComponent[Position](w, e); ok {
		t.Error("expected Position to be removed")
	}
	if _, ok := GetComponent[Name](w, e); ok {
		t.Error("expected Name to be removed")
	}
}

func TestWorldEntities(t *testing.T) {
	w := NewWorld()
	e1 := w.NewEntity()
	e2 := w.NewEntity()

	entities := w.Entities()
	if len(entities) != 2 {
		t.Fatalf("len(entities) = %d, want 2", len(entities))
	}
	found := make(map[Entity]bool)
	for _, e := range entities {
		found[e] = true
	}
	if !found[e1] || !found[e2] {
		t.Error("Entities missing expected values")
	}
}
