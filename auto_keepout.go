package autokeepout

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
)

var Colors = []color.NRGBA{
	{255, 0, 0, 255},
	{255, 125, 0, 255},
	{255, 255, 0, 255},
	{125, 255, 0, 255},
	{0, 255, 0, 255},
	{0, 255, 125, 255},
	{0, 255, 255, 255},
	{0, 125, 255, 255},
	{0, 0, 255, 255},
	{125, 0, 255, 255},
	{255, 0, 255, 255},
	{255, 0, 125, 255},
}

func CreateAutoKeepoutImage(outputDir string, imageName string, img image.Image) {
	maxPoint := img.Bounds().Max
	minPoint := img.Bounds().Min

	pixelGrid := make([][]float64, maxPoint.Y)
	for i := minPoint.Y; i < maxPoint.Y; i += 1 {
		pixelGrid[i] = make([]float64, maxPoint.X)
		for j := minPoint.X; j < maxPoint.X; j += 1 {
			pixelGrid[i][j] = RGBTo8BitGrayScaleIntensity(img.At(j, i))
		}
	}

	wallRemovedMask := FloodFillFromTopLeftCorner(pixelGrid)
	gaussMask := ParallelGaussianMasking(wallRemovedMask, 32)
	gradMask := ParallelGradientMasking(gaussMask)
	MaximumSuppression(gradMask)
	SimpleNearestNeighborClustering(gradMask, 10)
	hullMask := ConvexHullMasking(gradMask)

	newImage := image.NewNRGBA(img.Bounds())
	for y := minPoint.Y; y < maxPoint.Y; y += 1 {
		for x := minPoint.X; x < maxPoint.X; x += 1 {
			grad := gradMask[y][x]
			if grad.IsLocalMax {
				newImage.Set(x, y, Colors[grad.ClusterID%len(Colors)])
			} else {
				val := uint8(wallRemovedMask[y][x])
				if val < 0 {
					val = 0
				}
				newImage.Set(x, y, color.NRGBA{val, val, val, 255})
			}
		}
	}

	for i := range hullMask {
		for j := range hullMask[i] {
			newImage.Set(j, i, color.NRGBA{255, 0, 0, 255})
		}
	}

	outputFile, fileErr := os.Create(fmt.Sprintf("%s/%s_auto_keepout.png", outputDir, imageName))
	if fileErr != nil {
		fmt.Println("Cannot create image")
	} else {
		png.Encode(outputFile, newImage)
		outputFile.Close()
	}
}
