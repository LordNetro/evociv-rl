package ecs

import (
	"sync"
	"testing"
)

func TestNewComponentIDUnique(t *testing.T) {
	id1 := NewComponentID("position")
	id2 := NewComponentID("name")
	id3 := NewComponentID("position")

	if id1 == id2 {
		t.Error("NewComponentID returned same ID for different names")
	}
	if id1 != id3 {
		t.Error("NewComponentID returned different ID for same name")
	}
}

func TestNewComponentIDConcurrent(t *testing.T) {
	const n = 100
	var wg sync.WaitGroup
	ids := make([]ComponentID, n)

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			ids[idx] = NewComponentID("concurrent")
		}(i)
	}
	wg.Wait()

	for i := 1; i < n; i++ {
		if ids[i] != ids[0] {
			t.Error("concurrent NewComponentID returned different IDs for same name")
		}
	}
}

func TestComponentTypes(t *testing.T) {
	p := Position{X: 1.0, Y: 2.0, Z: 3}
	if p.X != 1.0 || p.Y != 2.0 || p.Z != 3 {
		t.Errorf("Position fields incorrect: got %+v", p)
	}

	n := Name{Name: "TestEntity"}
	if n.Name != "TestEntity" {
		t.Errorf("Name.Name = %q, want %q", n.Name, "TestEntity")
	}

	tags := Tags{Tags: []string{"hostile", "flying"}}
	if len(tags.Tags) != 2 || tags.Tags[0] != "hostile" {
		t.Errorf("Tags incorrect: got %+v", tags)
	}
}
