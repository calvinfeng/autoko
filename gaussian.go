package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
)

var GaussianMask = [][]float64{
	{2.0, 4.0, 5.0, 4.0, 2.0},
	{4.0, 9.0, 12.0, 9.0, 4.0},
	{5.0, 12.0, 15.0, 12.0, 5.0},
	{4.0, 9.0, 12.0, 9.0, 4.0},
	{2.0, 4.0, 5.0, 4.0, 2.0},
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

func ApplyPartialGaussianMask(grid [][]float64, startRow, rowsPerRoutine int, output chan [][]float64) {
	partialMask := make([][]float64, rowsPerRoutine)
	for i := 0; i < rowsPerRoutine; i += 1 {
		partialMask[i] = make([]float64, len(grid[i+startRow]))
		for j := 0; j < len(grid[i+startRow]); j += 1 {
			partialMask[i][j] = Convolve(i+startRow, j, grid, GaussianMask)
		}
	}

	output <- partialMask
}

func ApplyGaussianMaskOptimized(grid [][]float64) [][]float64 {
	numOfRoutines := 32
	rowsPerRoutine := len(grid) / numOfRoutines
	outputChan := make(chan [][]float64, numOfRoutines)
	startRow := 0

	for n := 1; n < numOfRoutines; n += 1 {
		go ApplyPartialGaussianMask(grid, startRow, rowsPerRoutine, outputChan)
		startRow += rowsPerRoutine
	}
	go ApplyPartialGaussianMask(grid, startRow, len(grid)-startRow, outputChan)

	mask := make([][]float64, len(grid))
	for partialResult := range outputChan {
		mask = append(mask, partialResult...)
		if len(mask) > len(grid) {
			break
		}
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

	maskedGrid := ApplyGaussianMaskOptimized(pixelGrid)

	newImage := image.NewGray(img.Bounds())
	for y := 0; y < len(maskedGrid); y += 1 {
		for x := 0; x < len(maskedGrid[y]); x += 1 {
			val := maskedGrid[y][x]
			if val < 0.0 {
				val = 0.0
			}

			newImage.Set(x, y, color.Gray{uint8(val)})
		}
	}

	//newImage := image.NewGray(img.Bounds())
	//for y := minPoint.Y; y < maxPoint.Y; y += 1 {
	//	for x := minPoint.X; x < maxPoint.X; x += 1 {
	//		val := maskedGrid[y][x]
	//		if val < 0.0 {
	//			val = 0.0
	//		}
	//		newImage.Set(x, y, color.Gray{uint8(val)})
	//	}
	//}

	outputFile, fileErr := os.Create(fmt.Sprintf("%s/%s_gaussian_blur.png", outputDir, imageName))
	if fileErr != nil {
		fmt.Println("Cannot create image")
	} else {
		png.Encode(outputFile, newImage)
		outputFile.Close()
	}
}
