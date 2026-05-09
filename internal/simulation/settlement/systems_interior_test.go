package settlement

import (
	"testing"
)

// TestInteriorGeneratorInterface verifies the InteriorGenerator implementation
func TestInteriorGeneratorInterface(t *testing.T) {
	t.Run("generator implements interface", func(t *testing.T) {
		var ig InteriorGeneratorInterface = &InteriorGenerator{}
		if ig == nil {
			t.Error("InteriorGenerator should implement InteriorGeneratorInterface")
		}
	})

	t.Run("generator generates interior with correct dimensions", func(t *testing.T) {
		g := &InteriorGenerator{}
		bi := g.Generate(12345, "house", 5, 4)

		if bi.Width != 5 {
			t.Errorf("Width should be 5, got %d", bi.Width)
		}
		if bi.Height != 4 {
			t.Errorf("Height should be 4, got %d", bi.Height)
		}
		if bi.BuildingSeed != 12345 {
			t.Errorf("Seed should be 12345, got %d", bi.BuildingSeed)
		}
		if bi.MaxWorkers != 2 {
			t.Errorf("MaxWorkers should be 2, got %d", bi.MaxWorkers)
		}
		if bi.WorkersInside != 0 {
			t.Errorf("WorkersInside should be 0, got %d", bi.WorkersInside)
		}
	})

	t.Run("generator creates grid with floor tiles", func(t *testing.T) {
		g := &InteriorGenerator{}
		bi := g.Generate(12345, "house", 5, 4)

		if bi.Grid == nil {
			t.Error("Grid should not be nil for real generator")
		}
		if len(bi.Grid) != 4 {
			t.Errorf("Grid should have 4 rows, got %d", len(bi.Grid))
		}
	})

	t.Run("default generator is set to real implementation", func(t *testing.T) {
		if DefaultInteriorGenerator == nil {
			t.Error("DefaultInteriorGenerator should not be nil")
		}
		// Should be the real InteriorGenerator, not the stub
		if _, ok := DefaultInteriorGenerator.(*InteriorGenerator); !ok {
			t.Error("DefaultInteriorGenerator should be *InteriorGenerator")
		}
	})
}