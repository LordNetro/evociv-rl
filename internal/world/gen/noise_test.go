package gen

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

var updateGolden = flag.Bool("update", false, "update golden files")

func TestPerlinDeterministic(t *testing.T) {
	seed := int64(42)
	cases := []struct {
		x, y  float64
		scale float64
	}{
		{0, 0, 100},
		{0.5, 1.5, 100},
		{-3, 7, 100},
	}
	var first []float64
	for _, c := range cases {
		v := Perlin2D(c.x, c.y, c.scale, seed)
		first = append(first, v)
	}
	// Same seed must produce same values
	for i, c := range cases {
		v := Perlin2D(c.x, c.y, c.scale, seed)
		if v != first[i] {
			t.Errorf("Perlin2D(%v,%v,%v,%d) = %v, want %v (deterministic failure)", c.x, c.y, c.scale, seed, v, first[i])
		}
	}
}

func TestPerlinSeedsDiffer(t *testing.T) {
	v1 := Perlin2D(0.5, 0.5, 100, 42)
	v2 := Perlin2D(0.5, 0.5, 100, 99)
	if v1 == v2 {
		t.Errorf("seeds 42 and 99 produced same value %v", v1)
	}
}

func TestPerlinRange(t *testing.T) {
	// Sample many points and verify they are in [-1, 1]
	seed := int64(42)
	for x := 0; x < 10; x++ {
		for y := 0; y < 10; y++ {
			v := Perlin2D(float64(x), float64(y), 100, seed)
			if v < -1.0 || v > 1.0 {
				t.Errorf("Perlin2D(%d,%d) = %v, out of range [-1,1]", x, y, v)
			}
		}
	}
}

func TestFBMDeterministic(t *testing.T) {
	seed := int64(42)
	v1 := FBM2D(1.0, 2.0, 4, 2.0, 0.5, 100.0, seed)
	v2 := FBM2D(1.0, 2.0, 4, 2.0, 0.5, 100.0, seed)
	if v1 != v2 {
		t.Errorf("FBM2D not deterministic: %v != %v", v1, v2)
	}
}

func TestFBMOctavesIncreaseDetail(t *testing.T) {
	seed := int64(42)
	v1 := FBM2D(1.0, 2.0, 1, 2.0, 0.5, 100.0, seed)
	v4 := FBM2D(1.0, 2.0, 4, 2.0, 0.5, 100.0, seed)
	if v1 == v4 {
		t.Errorf("1 octave and 4 octaves produced same value %v", v1)
	}
}

func TestFBMRange(t *testing.T) {
	seed := int64(42)
	for x := 0; x < 10; x++ {
		for y := 0; y < 10; y++ {
			v := FBM2D(float64(x), float64(y), 6, 2.0, 0.5, 100.0, seed)
			if v < -1.5 || v > 1.5 {
				t.Errorf("FBM2D(%d,%d) = %v, unexpectedly out of reasonable range", x, y, v)
			}
		}
	}
}

func TestPerlin2DGolden(t *testing.T) {
	v := Perlin2D(0, 0, 100, 42)
	output := fmt.Sprintf("%.10f", v)
	goldenFile := filepath.Join("testdata", "perlin.golden")

	if *updateGolden {
		if err := os.MkdirAll("testdata", 0755); err != nil {
			t.Fatalf("failed to create testdata: %v", err)
		}
		if err := os.WriteFile(goldenFile, []byte(output), 0644); err != nil {
			t.Fatalf("failed to write golden file: %v", err)
		}
	}

	expected, err := os.ReadFile(goldenFile)
	if err != nil {
		t.Fatalf("reading golden file: %v", err)
	}
	if output != string(expected) {
		t.Errorf("output %q doesn't match golden file %q", output, string(expected))
	}
}
