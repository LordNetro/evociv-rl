package data

import (
	"fmt"
)

// Validator validates data inside a Registry.
type Validator interface {
	Validate(registry *Registry) []error
}

// uniqueIDValidator checks that the specified key is unique across items.
type uniqueIDValidator struct {
	kind string
	key  string
}

func (v *uniqueIDValidator) Validate(r *Registry) []error {
	var errs []error
	data, ok := Get[[]any](r, v.kind)
	if !ok {
		return nil
	}
	seen := make(map[string]bool)
	for _, item := range data {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		idVal, ok := m[v.key].(string)
		if !ok || idVal == "" {
			continue
		}
		if seen[idVal] {
			errs = append(errs, fmt.Errorf("duplicate %s %s: %s", v.kind, v.key, idVal))
		}
		seen[idVal] = true
	}
	return errs
}

// requiredFieldValidator checks that required fields are non-empty strings.
type requiredFieldValidator struct {
	kind   string
	fields []string
}

func (v *requiredFieldValidator) Validate(r *Registry) []error {
	var errs []error
	data, ok := Get[[]any](r, v.kind)
	if !ok {
		return nil
	}
	for _, item := range data {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		for _, f := range v.fields {
			val, ok := m[f].(string)
			if !ok || val == "" {
				errs = append(errs, fmt.Errorf("missing required field %q in %s", f, v.kind))
			}
		}
	}
	return errs
}
