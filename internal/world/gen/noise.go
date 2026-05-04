package gen

import (
	"math"
	"math/rand"
)

// Perlin2D returns a classic Perlin noise value in the range [-1, 1]
// for the given coordinates, scale, and seed.
func Perlin2D(x, y, scale float64, seed int64) float64 {
	if scale == 0 {
		scale = 1
	}
	x /= scale
	y /= scale

	perm := buildPerm(seed)

	// Find unit grid cell containing the point
	X := int(math.Floor(x)) & 255
	Y := int(math.Floor(y)) & 255

	// Relative coords within the cell
	x -= math.Floor(x)
	y -= math.Floor(y)

	// Fade curves
	u := fade(x)
	v := fade(y)

	// Hash coordinates of the 4 corners
	A := perm[X] + Y
	AA := perm[A]
	AB := perm[A+1]
	BA := perm[X+1] + Y
	BA2 := perm[BA]
	BB := perm[BA+1]

	// Blend results from the 4 corners
	res := lerp(v,
		lerp(u, grad(perm, AA, x, y, 0), grad(perm, BA2, x-1, y, 0)),
		lerp(u, grad(perm, AB, x, y-1, 0), grad(perm, BB, x-1, y-1, 0)),
	)

	return res
}

// FBM2D returns fractal Brownian motion noise by summing multiple octaves
// of Perlin noise.
func FBM2D(x, y float64, octaves int, lacunarity, gain, scale float64, seed int64) float64 {
	if octaves <= 0 {
		octaves = 1
	}
	var total float64
	var amplitude float64 = 1.0
	var frequency float64 = 1.0
	var maxValue float64

	for i := 0; i < octaves; i++ {
		total += Perlin2D(x*frequency, y*frequency, scale, seed) * amplitude
		maxValue += amplitude
		amplitude *= gain
		frequency *= lacunarity
	}

	if maxValue == 0 {
		return 0
	}
	return total / maxValue
}

func buildPerm(seed int64) []int {
	rng := rand.New(rand.NewSource(seed))
	perm := make([]int, 512)
	p := make([]int, 256)
	for i := range p {
		p[i] = i
	}
	rng.Shuffle(256, func(i, j int) {
		p[i], p[j] = p[j], p[i]
	})
	for i := 0; i < 512; i++ {
		perm[i] = p[i&255]
	}
	return perm
}

func fade(t float64) float64 {
	return t * t * t * (t*(t*6-15) + 10)
}

func lerp(t, a, b float64) float64 {
	return a + t*(b-a)
}

func grad(perm []int, hash int, x, y, z float64) float64 {
	h := hash & 15
	u := x
	if h >= 8 {
		u = y
	}
	v := y
	if h >= 4 {
		v = x
		if h == 12 || h == 14 {
			v = z
		}
	}
	if (h & 1) == 1 {
		u = -u
	}
	if (h & 2) == 2 {
		v = -v
	}
	return u + v
}
