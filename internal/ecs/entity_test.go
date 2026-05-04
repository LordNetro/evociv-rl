package ecs

import "testing"

func TestEntityInvalid(t *testing.T) {
	if EntityInvalid != 0 {
		t.Errorf("EntityInvalid = %d, want 0", EntityInvalid)
	}
}

func TestEntityType(t *testing.T) {
	var e Entity = 42
	if e != 42 {
		t.Errorf("Entity = %d, want 42", e)
	}
}
