package ecs

import "testing"

type dummySystem struct {
	name    string
	counter int
}

func (s *dummySystem) Update(w *World, dt float64) error {
	s.counter++
	return nil
}

func (s *dummySystem) Name() string {
	return s.name
}

func TestSystemManagerAddSystem(t *testing.T) {
	sm := NewSystemManager()
	s := &dummySystem{name: "dummy"}
	sm.AddSystem(s)
	if len(sm.Systems()) != 1 {
		t.Fatalf("expected 1 system, got %d", len(sm.Systems()))
	}
}

func TestSystemManagerUpdateAll(t *testing.T) {
	sm := NewSystemManager()
	s1 := &dummySystem{name: "s1"}
	s2 := &dummySystem{name: "s2"}
	sm.AddSystem(s1)
	sm.AddSystem(s2)

	err := sm.UpdateAll(NewWorld(), 1.0)
	if err != nil {
		t.Fatalf("UpdateAll error: %v", err)
	}
	if s1.counter != 1 {
		t.Errorf("s1.counter = %d, want 1", s1.counter)
	}
	if s2.counter != 1 {
		t.Errorf("s2.counter = %d, want 1", s2.counter)
	}
}

func TestWorldUpdate(t *testing.T) {
	w := NewWorld()
	s := &dummySystem{name: "world-dummy"}
	w.AddSystem(s)

	err := w.Update(1.0)
	if err != nil {
		t.Fatalf("Update error: %v", err)
	}
	if s.counter != 1 {
		t.Errorf("counter = %d, want 1", s.counter)
	}
}
