package ecs

import (
	"reflect"
	"sync/atomic"
)

// World manages entities, component stores, and systems.
type World struct {
	entities     []Entity
	nextID       uint64
	stores       map[ComponentID]any
	typeRegistry map[reflect.Type]ComponentID
	systems      *SystemManager
}

// NewWorld creates a new empty World.
func NewWorld() *World {
	return &World{
		entities:     make([]Entity, 0),
		stores:       make(map[ComponentID]any),
		typeRegistry: make(map[reflect.Type]ComponentID),
		systems:      NewSystemManager(),
	}
}

// NewEntity creates a new unique entity.
func (w *World) NewEntity() Entity {
	id := Entity(atomic.AddUint64(&w.nextID, 1))
	w.entities = append(w.entities, id)
	return id
}

// RegisterComponentStore registers a component store for a given component ID.
func RegisterComponentStore[T any](w *World, id ComponentID, store *ComponentStore[T]) {
	w.stores[id] = store
	var zero T
	w.typeRegistry[reflect.TypeOf(zero)] = id
}

// GetStore returns the registered store for the given component ID.
func (w *World) GetStore(id ComponentID) any {
	return w.stores[id]
}

// AddComponent adds a component of type T to the given entity.
func AddComponent[T any](w *World, e Entity, c T) {
	var zero T
	t := reflect.TypeOf(zero)
	id, ok := w.typeRegistry[t]
	if !ok {
		panic("component type not registered")
	}
	store, ok := w.stores[id].(*ComponentStore[T])
	if !ok {
		panic("component store type mismatch")
	}
	store.Set(e, c)
}

// GetComponent returns the component of type T for the given entity.
func GetComponent[T any](w *World, e Entity) (T, bool) {
	var zero T
	t := reflect.TypeOf(zero)
	id, ok := w.typeRegistry[t]
	if !ok {
		return zero, false
	}
	store, ok := w.stores[id].(*ComponentStore[T])
	if !ok {
		return zero, false
	}
	return store.Get(e)
}

// RemoveEntity removes an entity and all its components.
func (w *World) RemoveEntity(e Entity) {
	for _, store := range w.stores {
		switch s := store.(type) {
		case *ComponentStore[Position]:
			s.Delete(e)
		case *ComponentStore[Name]:
			s.Delete(e)
		case *ComponentStore[Tags]:
			s.Delete(e)
		default:
			// Use reflection to call Delete if the store has a Delete method.
			v := reflect.ValueOf(store)
			m := v.MethodByName("Delete")
			if m.IsValid() {
				m.Call([]reflect.Value{reflect.ValueOf(e)})
			}
		}
	}
	for i, ent := range w.entities {
		if ent == e {
			w.entities = append(w.entities[:i], w.entities[i+1:]...)
			break
		}
	}
}

// Entities returns a copy of all active entities.
func (w *World) Entities() []Entity {
	out := make([]Entity, len(w.entities))
	copy(out, w.entities)
	return out
}

// AddSystem registers a system with the world.
func (w *World) AddSystem(s System) {
	w.systems.AddSystem(s)
}

// Systems returns all registered systems.
func (w *World) Systems() []System {
	return w.systems.Systems()
}

// Update runs all systems with the given delta time.
func (w *World) Update(dt float64) error {
	return w.systems.UpdateAll(w, dt)
}
