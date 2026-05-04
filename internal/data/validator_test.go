package data

import "testing"

func TestValidatorUniqueIDsPass(t *testing.T) {
	r := NewRegistry()
	r.Register("biomes", []any{
		map[string]any{"id": "plains", "name": "Llanuras"},
		map[string]any{"id": "forest", "name": "Bosque"},
	})

	v := &uniqueIDValidator{kind: "biomes", key: "id"}
	errs := v.Validate(r)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestValidatorUniqueIDsFail(t *testing.T) {
	r := NewRegistry()
	r.Register("biomes", []any{
		map[string]any{"id": "plains", "name": "Llanuras"},
		map[string]any{"id": "plains", "name": "Duplicate"},
	})

	v := &uniqueIDValidator{kind: "biomes", key: "id"}
	errs := v.Validate(r)
	if len(errs) == 0 {
		t.Error("expected errors for duplicate IDs")
	}
}

func TestValidatorRequiredFieldsPass(t *testing.T) {
	r := NewRegistry()
	r.Register("biomes", []any{
		map[string]any{"id": "plains", "name": "Llanuras"},
	})

	v := &requiredFieldValidator{kind: "biomes", fields: []string{"id", "name"}}
	errs := v.Validate(r)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestValidatorRequiredFieldsFail(t *testing.T) {
	r := NewRegistry()
	r.Register("biomes", []any{
		map[string]any{"id": "", "name": "Llanuras"},
		map[string]any{"id": "forest", "name": ""},
	})

	v := &requiredFieldValidator{kind: "biomes", fields: []string{"id", "name"}}
	errs := v.Validate(r)
	if len(errs) != 2 {
		t.Errorf("expected 2 errors, got %d", len(errs))
	}
}
