package ecs

// ComponentStore maps entities to components of type T.
type ComponentStore[T any] struct {
	components map[Entity]T
}

// NewComponentStore creates a new empty ComponentStore.
func NewComponentStore[T any]() *ComponentStore[T] {
	return &ComponentStore[T]{
		components: make(map[Entity]T),
	}
}

// Get returns the component for the given entity and whether it exists.
func (cs *ComponentStore[T]) Get(e Entity) (T, bool) {
	c, ok := cs.components[e]
	return c, ok
}

// Set assigns a component to the given entity.
func (cs *ComponentStore[T]) Set(e Entity, c T) {
	cs.components[e] = c
}

// Delete removes the component for the given entity.
func (cs *ComponentStore[T]) Delete(e Entity) {
	delete(cs.components, e)
}

// Has reports whether the given entity has a component in this store.
func (cs *ComponentStore[T]) Has(e Entity) bool {
	_, ok := cs.components[e]
	return ok
}

// Len returns the number of entities with components in this store.
func (cs *ComponentStore[T]) Len() int {
	return len(cs.components)
}

// All returns a copy of the underlying component map.
func (cs *ComponentStore[T]) All() map[Entity]T {
	out := make(map[Entity]T, len(cs.components))
	for e, c := range cs.components {
		out[e] = c
	}
	return out
}
