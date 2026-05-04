package data

import "reflect"

// Registry holds typed data by name.
type Registry struct {
	stores map[string]any
}

// NewRegistry creates a new empty Registry.
func NewRegistry() *Registry {
	return &Registry{
		stores: make(map[string]any),
	}
}

// Register stores data under the given name.
func (r *Registry) Register(name string, data any) {
	r.stores[name] = data
}

// Get retrieves data of type T by name.
func Get[T any](r *Registry, name string) (T, bool) {
	var zero T
	v, ok := r.stores[name]
	if !ok {
		return zero, false
	}
	typed, ok := v.(T)
	if !ok {
		return zero, false
	}
	return typed, true
}

// All returns all stored values of type T.
func All[T any](r *Registry) []T {
	var out []T
	var target reflect.Type
	var zero T
	target = reflect.TypeOf(zero)
	for _, v := range r.stores {
		if reflect.TypeOf(v) == target {
			out = append(out, v.(T))
		}
	}
	return out
}

// Types returns the reflect.Type names of all registered values.
func (r *Registry) Types() []string {
	seen := make(map[string]bool)
	for _, v := range r.stores {
		seen[reflect.TypeOf(v).String()] = true
	}
	out := make([]string, 0, len(seen))
	for ty := range seen {
		out = append(out, ty)
	}
	return out
}
