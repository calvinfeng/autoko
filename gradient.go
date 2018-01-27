package main

import (
	"math"
)

// Gradient is a vector which has vertical and horizontal component. It also contains a directional component that is
// quantized to be one of the eight possible choices (N, NE, E, SE, S, SW, W, NW).
type Gradient struct {
	Y   float64
	X   float64
	Dir string
}

const (
	N  = "NORTH"
	NE = "NORTH_EAST"
	E  = "EAST"
	SE = "SOUTH_EAST"
	S  = "SOUTH"
	SW = "SOUTH_WEST"
	W  = "WEST"
	NW = "NORTH_WEST"
)

// Gx is the Sobel operator in the x-direction
var Gx = [][]float64{
	{-1.0, 0.0, 1.0},
	{-2.0, 0.0, 2.0},
	{-1.0, 0.0, 1.0},
}

// Gy is the sobel operator in the y-direction
var Gy = [][]float64{
	{1.0, 2.0, 1.0},
	{0.0, 0.0, 0.0},
	{-1.0, -2.0, -1.0},
}

func GetImageGradient(grid [][]float64, y, x int) (*Gradient, error) {
	grad := &Gradient{
		X: 0.0,
		Y: 0.0,
	}

	// When we perform Sobel convolution on the pixel grid, we assume zero padding if it goes out of bound.
	for i := 0; i < 3; i += 1 {
		for j := 0; j < 3; j += 1 {
			outOfBound := false
			if 0 > i+y-1 || len(grid) <= i+y-1 {
				outOfBound = true
			}

			if 0 > j+x-1 || len(grid[i]) <= j+x-1 {
				outOfBound = true
			}

			if !outOfBound {
				grad.Y += Gy[i][j] * grid[i][j]
				grad.X += Gx[i][j] * grid[i][j]
			}
		}
	}

	// When X is zero, we have a division by zero case here.
	if grad.X == 0.0 {
		if grad.Y > 0.0 {
			grad.Dir = N
		} else if grad.Y < 0.0 {
			grad.Dir = S
		}

		return grad, nil
	}

	angle := math.Atan2(grad.Y, grad.X)

	var quadrant int
	if grad.X > 0.0 && grad.Y >= 0.0 {
		quadrant = 1
	} else if grad.X < 0.0 && grad.Y >= 0.0 {
		quadrant = 2
	} else if grad.X < 0.0 && grad.Y < 0.0 {
		quadrant = 3
	} else if grad.X > 0.0 && grad.Y < 0.0 {
		quadrant = 4
	}

	switch quadrant {
	case 1:
		if 0 <= angle && angle < math.Pi/8 {
			grad.Dir = E
		} else if math.Pi/8 <= angle && angle < 3*math.Pi/8 {
			grad.Dir = NE
		} else {
			grad.Dir = N
		}
	case 2:
		if math.Pi/2 <= angle && angle < 5*math.Pi/8 {
			grad.Dir = N
		} else if 5*math.Pi/8 < angle && angle < 7*math.Pi/8 {
			grad.Dir = NW
		} else {
			grad.Dir = W
		}
	case 3:
		angle += 2 * math.Pi
		if math.Pi <= angle && angle < 9*math.Pi/8 {
			grad.Dir = W
		} else if 9*math.Pi/8 <= angle && angle < 11*math.Pi/8 {
			grad.Dir = SW
		} else {
			grad.Dir = S
		}
	case 4:
		angle += 2 * math.Pi
		if 1.5*math.Pi <= angle && angle < 13*math.Pi/8 {
			grad.Dir = S
		} else if 13*math.Pi/8 <= angle && angle < 15*math.Pi/8 {
			grad.Dir = SE
		} else {
			grad.Dir = E
		}
	default:
		return nil, MathError{"Cannot find the appropriate quadrant to determine direction of this vector."}
	}

	return grad, nil
}
