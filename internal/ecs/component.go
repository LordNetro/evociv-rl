package ecs

import (
	"sync"
	"sync/atomic"
)

// ComponentID uniquely identifies a component type.
type ComponentID uint64

var (
	componentCounter uint64
	componentNames   = make(map[string]ComponentID)
	componentMu      sync.RWMutex
)

// NewComponentID returns a unique ComponentID for the given name.
// Calling with the same name always returns the same ID.
func NewComponentID(name string) ComponentID {
	componentMu.RLock()
	if id, ok := componentNames[name]; ok {
		componentMu.RUnlock()
		return id
	}
	componentMu.RUnlock()

	componentMu.Lock()
	defer componentMu.Unlock()
	if id, ok := componentNames[name]; ok {
		return id
	}
	id := ComponentID(atomic.AddUint64(&componentCounter, 1))
	componentNames[name] = id
	return id
}

// Position represents a 3D position component.
type Position struct {
	X, Y float64
	Z    int
}

// Name represents an entity name component.
type Name struct {
	Name string
}

// Tags represents a set of string tags attached to an entity.
type Tags struct {
	Tags []string
}
