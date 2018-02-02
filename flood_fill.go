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

	startCoord := &Coordinate{0, 0}
	distance := 5
	sourceValue := grid[startCoord.I][startCoord.J]
	targetValue := 255.0
	breadthFirstFloodFill(startCoord, distance, grid, mask, visitRecord, sourceValue, targetValue)

	return mask
}

func breadthFirstFloodFill(c *Coordinate, dist int, grid, mask [][]float64, visitRecord [][]bool, srcVal, targetVal float64) {
	queue := []*Coordinate{c}
	// Once it goes into the queue, mark it as visited and modify the mask right away.
	mask[c.I][c.J] = targetVal
	visitRecord[c.I][c.J] = true
	for len(queue) > 0 {
		coord := queue[0]
		queue = queue[1:]
		// Expand to neighboring pixels only if the current pixel on the grid matches the source value
		if grid[coord.I][coord.J] == srcVal {
			for i := coord.I - dist; i <= coord.I+dist; i += 1 {
				for j := coord.J - dist; j <= coord.J+dist; j += 1 {
					if i < 0 || i >= len(grid) {
						continue
					}

					if j < 0 || j >= len(grid[i]) {
						continue
					}

					if visitRecord[i][j] {
						continue
					}

					if i == coord.I && j == coord.J {
						continue
					}

					queue = append(queue, &Coordinate{i, j})
					mask[i][j] = targetVal
					visitRecord[i][j] = true
				}
			}
		}
	}
}
