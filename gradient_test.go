package autokeepout

import "testing"

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
			setDirection(g)

			if g.Dir != expected {
				t.Errorf("graident direction is not %s", expected)
			}
		}
	})

	t.Run("DirectionForZeroGrad", func(t *testing.T) {
		g := &Gradient{}
		setDirection(g)

		if g.Dir != "" {
			t.Error("gradient direction should be zero valued")
		}
	})

	t.Run("TrickyDirections", func(t *testing.T) {
		var g *Gradient

		g = &Gradient{X: -0.174, Y: 0.985}
		setDirection(g)

		if g.Dir != N {
			t.Errorf("gradient direction should be %s", N)
		}

		g = &Gradient{X: -0.731, Y: -0.682}
		setDirection(g)

		if g.Dir != SW {
			t.Errorf("gradient direction should be %s", SW)
		}

		g = &Gradient{X: 0.961, Y: -0.276}
		setDirection(g)

		if g.Dir != E {
			t.Errorf("gradient direction should be %s", E)
		}
	})
}
