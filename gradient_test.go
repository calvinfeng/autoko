package autokeepout

import "testing"

func BenchmarkGradientMask(b *testing.B) {
	m := randomMat(1000, 1000)

	b.Run("RegularGradientMask", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			GradientMask(m)
		}
	})

	b.Run("ParallelGradientMask with 1 go-routine", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ParallelGradientMask(m, 1)
		}
	})

	b.Run("ParallelGradientMask with 2 go-routines", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ParallelGradientMask(m, 2)
		}
	})

	b.Run("ParallelGradientMask with 4 go-routines", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ParallelGradientMask(m, 4)
		}
	})

	b.Run("ParallelGradientMask with 32 go-routines", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ParallelGradientMask(m, 32)
		}
	})
}
func TestGradientDirection(t *testing.T) {
	cases := map[string]*Gradient{
		N:  &Gradient{X: 0, Y: 1},
		NE: &Gradient{X: 1, Y: 1},
		E:  &Gradient{X: 1, Y: 0},
		SE: &Gradient{X: 1, Y: -1},
		S:  &Gradient{X: 0, Y: -1},
		SW: &Gradient{X: -1, Y: -1},
		W:  &Gradient{X: -1, Y: 0},
		NW: &Gradient{X: -1, Y: 1},
	}

	t.Run("Directions", func(t *testing.T) {
		for expected, g := range cases {
			g.SetDirection()

			if g.Dir != expected {
				t.Errorf("graident direction is not %s", expected)
			}
		}
	})

	t.Run("DirectionForZeroGrad", func(t *testing.T) {
		g := &Gradient{}
		g.SetDirection()

		if g.Dir != "" {
			t.Error("gradient direction should be zero valued")
		}
	})

	t.Run("TrickyDirections", func(t *testing.T) {
		var g *Gradient

		g = &Gradient{X: -0.174, Y: 0.985}
		g.SetDirection()

		if g.Dir != N {
			t.Errorf("gradient direction should be %s", N)
		}

		g = &Gradient{X: -0.731, Y: -0.682}
		g.SetDirection()

		if g.Dir != SW {
			t.Errorf("gradient direction should be %s", SW)
		}

		g = &Gradient{X: 0.961, Y: -0.276}
		g.SetDirection()

		if g.Dir != E {
			t.Errorf("gradient direction should be %s", E)
		}
	})
}

func TestNonMaximumSuppression(t *testing.T) {
	mask := [][]*Gradient{
		{&Gradient{Y: 1, X: 0}, &Gradient{Y: 1, X: 0}, &Gradient{Y: 1, X: 0}},
		{&Gradient{Y: 1, X: 0}, &Gradient{Y: 11, X: 0}, &Gradient{Y: 1, X: 0}},
		{&Gradient{Y: 1, X: 0}, &Gradient{Y: 1, X: 0}, &Gradient{Y: 1, X: 0}},
	}

	for i := 0; i < len(mask); i++ {
		for j := 0; j < len(mask); j++ {
			mask[i][j].SetDirection()
		}
	}

	NonMaximumSuppression(mask, 10)
	t.Run("Center", func(t *testing.T) {
		if !mask[1][1].IsLocalMax {
			t.Error("gradient at (1, 1) should be a local maximum")
		}
	})

	t.Run("Corners", func(t *testing.T) {
		if mask[0][0].IsLocalMax {
			t.Error("gradient at (0, 0) should NOT be a local maximum")
		}

		if mask[2][2].IsLocalMax {
			t.Error("gradient at (2, 2) should NOT be a local maximum")
		}

		if mask[0][2].IsLocalMax {
			t.Error("gradient at (0, 2) should NOT be a local maximum")
		}

		if mask[2][0].IsLocalMax {
			t.Error("gradient at (2, 0) should NOT be a local maximum")
		}
	})
}
