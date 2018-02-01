package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
)

func CreateFloodFillImage(outputDir string, imageName string, img image.Image) {
	maxPoint := img.Bounds().Max
	minPoint := img.Bounds().Min

	pixelGrid := make([][]float64, maxPoint.Y)
	for i := minPoint.Y; i < maxPoint.Y; i += 1 {
		pixelGrid[i] = make([]float64, maxPoint.X)
		for j := minPoint.X; j < maxPoint.X; j += 1 {
			pixelGrid[i][j] = RGBTo8BitGrayScaleIntensity(img.At(j, i))
		}
	}

	maskedGrid := FloodFillFromTopLeftCorner(pixelGrid)

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

	outputFile, fileErr := os.Create(fmt.Sprintf("%s/%s_flood_fill.png", outputDir, imageName))
	if fileErr != nil {
		fmt.Println("Cannot create image")
	} else {
		png.Encode(outputFile, newImage)
		outputFile.Close()
	}
}

func FloodFillFromTopLeftCorner(grid [][]float64) [][]float64 {
	// Instantiate a mask that is an identical copy of the original grid
	mask := make([][]float64, len(grid))
	visitRecord := make([][]bool, len(grid))
	for i := 0; i < len(grid); i += 1 {
		visitRecord[i] = make([]bool, len(grid[i]))
		mask[i] = make([]float64, len(grid[i]))
		copy(mask[i], grid[i])
	}

	// Start flood fill from top-left corner
	startRow, startCol, distance := 0, 0, 5
	sourceValue := grid[startRow][startCol]
	targetValue := 255.0
	floodFill(startRow, startCol, distance, grid, mask, visitRecord, sourceValue, targetValue)

	return mask
}

func floodFill(y, x, distance int, grid, mask [][]float64, visitRecord [][]bool, sourceValue, targetValue float64) {
	visitRecord[y][x] = true
	mask[y][x] = targetValue

	// Expand to neighboring pixels only if the source pixel matches the source value
	if grid[y][x] == sourceValue {
		for i := y - distance; i <= y+distance; i += 1 {
			for j := x - distance; j <= x+distance; j += 1 {
				if i < 0 || i >= len(grid) {
					continue
				}

				if j < 0 || j >= len(grid[i]) {
					continue
				}

				if visitRecord[i][j] {
					continue
				}

				if i == y && j == x {
					continue
				}

				floodFill(i, j, distance, grid, mask, visitRecord, grid[y][x], targetValue)
			}
		}
	}
}
