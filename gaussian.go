package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"time"
)

var GaussianMask = [][]float64{
	{2.0, 4.0, 5.0, 4.0, 2.0},
	{4.0, 9.0, 12.0, 9.0, 4.0},
	{5.0, 12.0, 15.0, 12.0, 5.0},
	{4.0, 9.0, 12.0, 9.0, 4.0},
	{2.0, 4.0, 5.0, 4.0, 2.0},
}

type PartialMask struct {
	Order    int
	Values   [][]float64
	StartRow int
}

// Convolve will assume zero padding
func Convolve(y, x int, grid [][]float64, kernel [][]float64) float64 {
	kernelNorm, sum := 0.0, 0.0
	for i := 0; i < 5; i += 1 {
		for j := 0; j < 5; j += 1 {
			// Check if it is out of bound
			outOfBound := false
			if 0 > i+y-2 || len(grid) <= i+y-2 {
				outOfBound = true
			}

			if 0 > j+x-2 || len(grid[i]) <= j+x-2 {
				outOfBound = true
			}

			if !outOfBound {
				kernelNorm += kernel[i][j]
				sum += grid[i+y-2][j+x-2] * kernel[i][j]
			}
		}
	}

	return sum / kernelNorm
}

func ApplyGaussianMask(grid [][]float64) [][]float64 {
	maskedGrid := make([][]float64, len(grid))
	for i := 0; i < len(grid); i += 1 {
		maskedGrid[i] = make([]float64, len(grid[i]))
		for j := 0; j < len(grid[i]); j += 1 {
			maskedGrid[i][j] = Convolve(i, j, grid, GaussianMask)
		}
	}

	return maskedGrid
}

func ApplyPartialGaussianMask(grid [][]float64, n, startRow, endRow int, output chan *PartialMask) {
	rowSize := endRow - startRow
	values := make([][]float64, rowSize)
	for i := 0; i < rowSize; i += 1 {
		colSize := len(grid[startRow+i])
		values[i] = make([]float64, colSize)
		for j := 0; j < colSize; j += 1 {
			values[i][j] = Convolve(startRow+i, j, grid, GaussianMask)
		}
	}

	output <- &PartialMask{
		Order:    n,
		Values:   values,
		StartRow: startRow,
	}
}

func ApplyGaussianMaskOptimized(grid [][]float64) [][]float64 {
	numOfRoutines := 32
	rowsPerRoutine := len(grid) / numOfRoutines
	outputChan := make(chan *PartialMask, numOfRoutines)

	n := 0
	for n < numOfRoutines-1 {
		go ApplyPartialGaussianMask(grid, n, n*rowsPerRoutine, (n+1)*rowsPerRoutine, outputChan)
		n += 1
	}
	go ApplyPartialGaussianMask(grid, n, n*rowsPerRoutine, len(grid), outputChan)

	n = 0
	partialMasks := make([]*PartialMask, numOfRoutines)
	for partialMask := range outputChan {
		partialMasks[partialMask.Order] = partialMask
		n += 1
		if n == numOfRoutines {
			break
		}
	}

	mask := [][]float64{}
	for _, partialMask := range partialMasks {
		mask = append(mask, partialMask.Values...)
	}

	return mask
}

func CreateGaussianBlurImage(outputDir string, imageName string, img image.Image) {
	maxPoint := img.Bounds().Max
	minPoint := img.Bounds().Min

	pixelGrid := make([][]float64, maxPoint.Y)
	for i := minPoint.Y; i < maxPoint.Y; i += 1 {
		pixelGrid[i] = make([]float64, maxPoint.X)
		for j := minPoint.X; j < maxPoint.X; j += 1 {
			pixelGrid[i][j] = RGBTo8BitGrayScale(img.At(j, i))
		}
	}

	start := time.Now()
	maskedGrid := ApplyGaussianMaskOptimized(pixelGrid)
	end := time.Now()
	fmt.Printf("Algorithm took %v time\n", end.Sub(start))

	newImage := image.NewGray(img.Bounds())
	for y := minPoint.Y; y < maxPoint.Y; y += 1 {
		for x := minPoint.X; x < maxPoint.X; x += 1 {
			val := maskedGrid[y][x]
			if val < 0.0 {
				val = 0.0
			}
			newImage.Set(x, y, color.Gray{uint8(val)})
		}
	}

	outputFile, fileErr := os.Create(fmt.Sprintf("%s/%s_gaussian_blur.png", outputDir, imageName))
	if fileErr != nil {
		fmt.Println("Cannot create image")
	} else {
		png.Encode(outputFile, newImage)
		outputFile.Close()
	}
}
