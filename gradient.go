package autokeepout

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

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

// GradientSubmask is part of a larger gradient mask that is applied to the whole image.
// TODO: Combine this with Submask using empty interface for values.
type GradientSubmask struct {
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
	NonMaximumSuppression(gradMask, 255)

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

func getGradientSubmask(mat [][]float64, n, startRow, endRow int, output chan *GradientSubmask) {
	rowSize := endRow - startRow
	values := make([][]*Gradient, rowSize)
	for i := 0; i < rowSize; i++ {
		colSize := len(mat[startRow+i])
		values[i] = make([]*Gradient, colSize)
		for j := 0; j < colSize; j++ {
			values[i][j] = computeGradient(mat, startRow+i, j)
		}
	}

	output <- &GradientSubmask{
		Order:    n,
		StartRow: startRow,
		Values:   values,
	}
}

func ParallelGradientMask(mat [][]float64) [][]*Gradient {
	numOfRoutines := 32
	rowsPerRoutine := len(mat) / numOfRoutines
	outputChan := make(chan *GradientSubmask, numOfRoutines)

	n := 0
	for n < numOfRoutines-1 {
		go getGradientSubmask(mat, n, n*rowsPerRoutine, (n+1)*rowsPerRoutine, outputChan)
		n++
	}
	go getGradientSubmask(mat, n, n*rowsPerRoutine, len(mat), outputChan)

	n = 0
	partialMasks := make([]*GradientSubmask, numOfRoutines)
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

// NonMaximumSuppression looks at each gradient in the matrix and identifies local maxima. The
// reason why it is called non-maximum suppression is that it normally sets the gradients to zero if
// they are not local maxima.
func NonMaximumSuppression(mask [][]*Gradient, threshold float64) {
	for i := 0; i < len(mask); i++ {
		for j := 0; j < len(mask[i]); j++ {
			var forward, backward *Coordinate
			switch mask[i][j].Dir {
			case E:
				forward = &Coordinate{i, j + 1}
				backward = &Coordinate{i, j - 1}
			case NE:
				forward = &Coordinate{i - 1, j + 1}
				backward = &Coordinate{i + 1, j - 1}
			case N:
				forward = &Coordinate{i - 1, j}
				backward = &Coordinate{i + 1, j}
			case NW:
				forward = &Coordinate{i - 1, j - 1}
				backward = &Coordinate{i + 1, j + 1}
			case W:
				forward = &Coordinate{i, j - 1}
				backward = &Coordinate{i, j + 1}
			case SW:
				forward = &Coordinate{i + 1, j - 1}
				backward = &Coordinate{i - 1, j + 1}
			case S:
				forward = &Coordinate{i + 1, j}
				backward = &Coordinate{i - 1, j}
			case SE:
				forward = &Coordinate{i + 1, j + 1}
				backward = &Coordinate{i - 1, j - 1}
			default:
				forward = &Coordinate{i, j}
				backward = &Coordinate{i, j}
			}

			numRow, numCol := len(mask), len(mask[i])
			if forward.IsInBound(numRow, numCol) && backward.IsInBound(numRow, numCol) {
				mask[i][j].IsLocalMax = mask[forward.I][forward.J].magnitude() < mask[i][j].magnitude() &&
					mask[backward.I][backward.J].magnitude() < mask[i][j].magnitude() &&
					mask[i][j].magnitude() > threshold
			}
		}
	}
}

// Convolve performs convolution on a given location of a matrix. The operation assumes zero padding
// i.e. the contribution from out of bound region is zero.
func convolve(mat [][]float64, y, x, kernelSize int, kernel [][]float64) (sum float64) {
	if kernelSize%2 != 1 {
		panic("kernel size must be an odd integer")
	}

	offset := (kernelSize - 1) / 2

	for i := 0; i < kernelSize; i++ {
		if y+i-offset < 0 || len(mat) <= y+i-offset {
			continue
		}

		for j := 0; j < kernelSize; j++ {
			if x+j-offset < 0 || len(mat[i]) <= x+j-offset {
				continue
			}

			sum += kernel[i][j] * mat[y+i-offset][x+j-offset]
		}
	}

	return sum
}

func computeGradient(mat [][]float64, y, x int) *Gradient {
	grad := &Gradient{
		X: convolve(mat, y, x, 3, Gx),
		Y: convolve(mat, y, x, 3, Gy),
	}

	grad.setDirection()
	return grad
}

// Gradient is a vector which has vertical and horizontal component. It also contains a directional component that is
// quantized to be one of the eight possible choices (N, NE, E, SE, S, SW, W, NW).
type Gradient struct {
	Y          float64
	X          float64
	Dir        string
	IsLocalMax bool
	ClusterID  int
}

func (g *Gradient) magnitude() float64 {
	return math.Sqrt(g.X*g.X + g.Y*g.Y)
}

func (g *Gradient) setDirection() {
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
