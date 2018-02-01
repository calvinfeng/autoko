package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

// Gradient is a vector which has vertical and horizontal component. It also contains a directional component that is
// quantized to be one of the eight possible choices (N, NE, E, SE, S, SW, W, NW).
type Gradient struct {
	Y          float64
	X          float64
	Dir        string
	IsLocalMax bool
}

func (g *Gradient) Magnitude() float64 {
	return math.Sqrt(g.X*g.X + g.Y*g.Y)
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

// Gy is the Sobel operator in the y-direction
var Gy = [][]float64{
	{1.0, 2.0, 1.0},
	{0.0, 0.0, 0.0},
	{-1.0, -2.0, -1.0},
}

type PartialGradientMask struct {
	Order    int
	StartRow int
	Values   [][]*Gradient
}

func CreateEdgeDetectionImage(outputDir string, imageName string, img image.Image) {
	maxPoint := img.Bounds().Max
	minPoint := img.Bounds().Min

	pixelGrid := make([][]float64, maxPoint.Y)
	for i := minPoint.Y; i < maxPoint.Y; i += 1 {
		pixelGrid[i] = make([]float64, maxPoint.X)
		for j := minPoint.X; j < maxPoint.X; j += 1 {
			pixelGrid[i][j] = RGBTo8BitGrayScaleIntensity(img.At(j, i))
		}
	}

	gaussMask := GetGaussianMaskOptimized(pixelGrid)
	gradMask := GetGradientMaskOptimized(gaussMask)
	MaximumSuppression(gradMask)

	newImage := image.NewNRGBA(img.Bounds())
	for y := minPoint.Y; y < maxPoint.Y; y += 1 {
		for x := minPoint.X; x < maxPoint.X; x += 1 {
			grad := gradMask[y][x]
			if grad.IsLocalMax {
				newImage.Set(x, y, color.NRGBA{255, 0, 0, 255})
			} else {
				newImage.Set(x, y, img.At(x, y))
			}
		}
	}

	outputFile, fileErr := os.Create(fmt.Sprintf("%s/%s_edge_detection.png", outputDir, imageName))
	if fileErr != nil {
		fmt.Println("Cannot create image")
	} else {
		png.Encode(outputFile, newImage)
		outputFile.Close()
	}
}

func GetGradientMask(grid [][]float64) [][]*Gradient {
	mask := make([][]*Gradient, len(grid))
	for i := 0; i < len(grid); i += 1 {
		mask[i] = make([]*Gradient, len(grid[i]))
		for j := 0; j < len(grid[i]); j += 1 {
			mask[i][j] = computeGradient(grid, i, j)
		}
	}

	return mask
}

func getPartialGradientMask(grid [][]float64, n, startRow, endRow int, output chan *PartialGradientMask) {
	rowSize := endRow - startRow
	values := make([][]*Gradient, rowSize)
	for i := 0; i < rowSize; i += 1 {
		colSize := len(grid[startRow+i])
		values[i] = make([]*Gradient, colSize)
		for j := 0; j < colSize; j += 1 {
			values[i][j] = computeGradient(grid, startRow+i, j)
		}
	}

	output <- &PartialGradientMask{
		Order:    n,
		StartRow: startRow,
		Values:   values,
	}
}

func GetGradientMaskOptimized(grid [][]float64) [][]*Gradient {
	numOfRoutines := 32
	rowsPerRoutine := len(grid) / numOfRoutines
	outputChan := make(chan *PartialGradientMask, numOfRoutines)

	n := 0
	for n < numOfRoutines-1 {
		go getPartialGradientMask(grid, n, n*rowsPerRoutine, (n+1)*rowsPerRoutine, outputChan)
		n += 1
	}
	go getPartialGradientMask(grid, n, n*rowsPerRoutine, len(grid), outputChan)

	n = 0
	partialMasks := make([]*PartialGradientMask, numOfRoutines)
	for partialMask := range outputChan {
		partialMasks[partialMask.Order] = partialMask
		n += 1
		if n == numOfRoutines {
			break
		}
	}

	mask := [][]*Gradient{}
	for _, partialMask := range partialMasks {
		mask = append(mask, partialMask.Values...)
	}

	return mask
}

func MaximumSuppression(mask [][]*Gradient) {
	for i := 0; i < len(mask); i += 1 {
		for j := 0; j < len(mask[i]); j += 1 {
			var forwardPos, backwardPos *Position
			switch mask[i][j].Dir {
			case E:
				forwardPos = &Position{i, j + 1}
				backwardPos = &Position{i, j - 1}
			case NE:
				forwardPos = &Position{i - 1, j + 1}
				backwardPos = &Position{i + 1, j - 1}
			case N:
				forwardPos = &Position{i - 1, j}
				backwardPos = &Position{i + 1, j}
			case NW:
				forwardPos = &Position{i - 1, j - 1}
				backwardPos = &Position{i + 1, j + 1}
			case W:
				forwardPos = &Position{i, j - 1}
				backwardPos = &Position{i, j + 1}
			case SW:
				forwardPos = &Position{i + 1, j - 1}
				backwardPos = &Position{i - 1, j + 1}
			case S:
				forwardPos = &Position{i + 1, j}
				backwardPos = &Position{i - 1, j}
			case SE:
				forwardPos = &Position{i + 1, j + 1}
				backwardPos = &Position{i - 1, j - 1}
			default:
				forwardPos = &Position{i, j}
				backwardPos = &Position{i, j}
			}

			if forwardPos.IsOutOfBound(len(mask), len(mask[i])) || backwardPos.IsOutOfBound(len(mask), len(mask[i])) {
				mask[i][j].IsLocalMax = false
			} else {
				mask[i][j].IsLocalMax = mask[forwardPos.I][forwardPos.J].Magnitude() < mask[i][j].Magnitude() &&
					mask[backwardPos.I][backwardPos.J].Magnitude() < mask[i][j].Magnitude()
			}
		}
	}
}

func computeGradient(grid [][]float64, y, x int) *Gradient {
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
				grad.Y += Gy[i][j] * grid[i+y-1][j+x-1]
				grad.X += Gx[i][j] * grid[i+y-1][j+x-1]
			}
		}
	}

	// When X is zero, we have a division by zero case here. Sometimes gradient can be zero when there is absolutely no
	// change within the local region.
	if grad.X == 0.0 {
		if grad.Y > 0.0 {
			grad.Dir = N
		} else if grad.Y < 0.0 {
			grad.Dir = S
		}

		return grad
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
	}

	return grad
}
