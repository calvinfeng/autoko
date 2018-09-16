package autokeepout

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
	ClusterID  int
}

func (g *Gradient) Magnitude() float64 {
	return math.Sqrt(g.X*g.X + g.Y*g.Y)
}

// Gradient directions
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

type GradSubmask struct {
	Order    int
	StartRow int
	Values   [][]*Gradient
}

func CreateEdgeDetectionImage(outputDir string, imageName string, img image.Image) {
	maxPoint := img.Bounds().Max
	minPoint := img.Bounds().Min

	pixelGrid := make([][]float64, maxPoint.Y)
	for i := minPoint.Y; i < maxPoint.Y; i++ {
		pixelGrid[i] = make([]float64, maxPoint.X)
		for j := minPoint.X; j < maxPoint.X; j++ {
			pixelGrid[i][j] = RGBTo8BitGrayScaleIntensity(img.At(j, i))
		}
	}

	gaussMask := ParallelGaussianMask(pixelGrid, 32)
	gradMask := ParallelGradientMask(gaussMask)
	MaximumSuppression(gradMask)

	newImage := image.NewNRGBA(img.Bounds())
	for y := minPoint.Y; y < maxPoint.Y; y++ {
		for x := minPoint.X; x < maxPoint.X; x++ {
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

// GradientMask takes an image matrix and returns a matrix of gradients.
func GradientMask(mat [][]float64) [][]*Gradient {
	mask := make([][]*Gradient, len(mat))
	for i := 0; i < len(mat); i++ {
		mask[i] = make([]*Gradient, len(mat[i]))
		for j := 0; j < len(mat[i]); j++ {
			mask[i][j] = computeGradient(mat, i, j)
		}
	}

	return mask
}

func getGradSubmask(mat [][]float64, n, startRow, endRow int, output chan *GradSubmask) {
	rowSize := endRow - startRow
	values := make([][]*Gradient, rowSize)
	for i := 0; i < rowSize; i++ {
		colSize := len(mat[startRow+i])
		values[i] = make([]*Gradient, colSize)
		for j := 0; j < colSize; j++ {
			values[i][j] = computeGradient(mat, startRow+i, j)
		}
	}

	output <- &GradSubmask{
		Order:    n,
		StartRow: startRow,
		Values:   values,
	}
}

func ParallelGradientMask(mat [][]float64) [][]*Gradient {
	numOfRoutines := 32
	rowsPerRoutine := len(mat) / numOfRoutines
	outputChan := make(chan *GradSubmask, numOfRoutines)

	n := 0
	for n < numOfRoutines-1 {
		go getGradSubmask(mat, n, n*rowsPerRoutine, (n+1)*rowsPerRoutine, outputChan)
		n++
	}
	go getGradSubmask(mat, n, n*rowsPerRoutine, len(mat), outputChan)

	n = 0
	partialMasks := make([]*GradSubmask, numOfRoutines)
	for partialMask := range outputChan {
		partialMasks[partialMask.Order] = partialMask
		n++
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
	for i := 0; i < len(mask); i++ {
		for j := 0; j < len(mask[i]); j++ {
			var forwardCoord, backwardCoord *Coordinate
			switch mask[i][j].Dir {
			case E:
				forwardCoord = &Coordinate{i, j + 1}
				backwardCoord = &Coordinate{i, j - 1}
			case NE:
				forwardCoord = &Coordinate{i - 1, j + 1}
				backwardCoord = &Coordinate{i + 1, j - 1}
			case N:
				forwardCoord = &Coordinate{i - 1, j}
				backwardCoord = &Coordinate{i + 1, j}
			case NW:
				forwardCoord = &Coordinate{i - 1, j - 1}
				backwardCoord = &Coordinate{i + 1, j + 1}
			case W:
				forwardCoord = &Coordinate{i, j - 1}
				backwardCoord = &Coordinate{i, j + 1}
			case SW:
				forwardCoord = &Coordinate{i + 1, j - 1}
				backwardCoord = &Coordinate{i - 1, j + 1}
			case S:
				forwardCoord = &Coordinate{i + 1, j}
				backwardCoord = &Coordinate{i - 1, j}
			case SE:
				forwardCoord = &Coordinate{i + 1, j + 1}
				backwardCoord = &Coordinate{i - 1, j - 1}
			default:
				forwardCoord = &Coordinate{i, j}
				backwardCoord = &Coordinate{i, j}
			}

			if forwardCoord.IsOutOfBound(len(mask), len(mask[i])) || backwardCoord.IsOutOfBound(len(mask), len(mask[i])) {
				mask[i][j].IsLocalMax = false
			} else {
				isMagLocalMax := mask[forwardCoord.I][forwardCoord.J].Magnitude() < mask[i][j].Magnitude() &&
					mask[backwardCoord.I][backwardCoord.J].Magnitude() < mask[i][j].Magnitude()
				mask[i][j].IsLocalMax = isMagLocalMax && mask[i][j].Magnitude() > 255
			}
		}
	}
}

func computeGradient(mat [][]float64, y, x int) *Gradient {
	grad := &Gradient{
		X: 0.0,
		Y: 0.0,
	}

	// When we perform Sobel convolution on the pixel mat, we assume zero padding if it goes out of bound.
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			outOfBound := false
			if 0 > i+y-1 || len(mat) <= i+y-1 {
				outOfBound = true
			}

			if 0 > j+x-1 || len(mat[i]) <= j+x-1 {
				outOfBound = true
			}

			if !outOfBound {
				grad.Y += Gy[i][j] * mat[i+y-1][j+x-1]
				grad.X += Gx[i][j] * mat[i+y-1][j+x-1]
			}
		}
	}

	setDirection(grad)
	return grad
}

func setDirection(g *Gradient) {
	if g.X == 0.0 {
		if g.Y > 0.0 {
			g.Dir = N
		} else if g.Y < 0.0 {
			g.Dir = S
		}

		return
	}

	angle := math.Atan2(g.Y, g.X)

	var quadrant int
	if g.X > 0.0 && g.Y >= 0.0 {
		quadrant = 1
	} else if g.X < 0.0 && g.Y >= 0.0 {
		quadrant = 2
	} else if g.X < 0.0 && g.Y < 0.0 {
		quadrant = 3
	} else if g.X > 0.0 && g.Y < 0.0 {
		quadrant = 4
	}

	switch quadrant {
	case 1:
		if 0 <= angle && angle < math.Pi/8 {
			g.Dir = E
		} else if math.Pi/8 <= angle && angle < 3*math.Pi/8 {
			g.Dir = NE
		} else {
			g.Dir = N
		}
	case 2:
		if math.Pi/2 <= angle && angle < 5*math.Pi/8 {
			g.Dir = N
		} else if 5*math.Pi/8 < angle && angle < 7*math.Pi/8 {
			g.Dir = NW
		} else {
			g.Dir = W
		}
	case 3:
		angle += 2 * math.Pi
		if math.Pi <= angle && angle < 9*math.Pi/8 {
			g.Dir = W
		} else if 9*math.Pi/8 <= angle && angle < 11*math.Pi/8 {
			g.Dir = SW
		} else {
			g.Dir = S
		}
	case 4:
		angle += 2 * math.Pi
		if 1.5*math.Pi <= angle && angle < 13*math.Pi/8 {
			g.Dir = S
		} else if 13*math.Pi/8 <= angle && angle < 15*math.Pi/8 {
			g.Dir = SE
		} else {
			g.Dir = E
		}
	}
}
