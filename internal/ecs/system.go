package ecs

// System is an interface for ECS systems that process entities.
type System interface {
	Update(w *World, dt float64) error
	Name() string
}

// SystemManager holds and runs registered systems.
type SystemManager struct {
	systems []System
}

// NewSystemManager creates a new SystemManager.
func NewSystemManager() *SystemManager {
	return &SystemManager{
		systems: make([]System, 0),
	}
}

// AddSystem registers a system.
func (sm *SystemManager) AddSystem(s System) {
	sm.systems = append(sm.systems, s)
}

// Systems returns the list of registered systems.
func (sm *SystemManager) Systems() []System {
	out := make([]System, len(sm.systems))
	copy(out, sm.systems)
	return out
}

// UpdateAll runs all systems in order.
func (sm *SystemManager) UpdateAll(w *World, dt float64) error {
	for _, s := range sm.systems {
		if err := s.Update(w, dt); err != nil {
			return err
		}
	}
	return nil
}
