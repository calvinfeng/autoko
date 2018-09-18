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

// CreateFloodFillImage takes an image and applies flood fill to it. The output is an image that has
// exterior wall dissolved.
func CreateFloodFillImage(outputDir string, imageName string, img image.Image) {
	maxPoint := img.Bounds().Max
	minPoint := img.Bounds().Min

	pixelGrid := make([][]float64, maxPoint.Y)
	for i := minPoint.Y; i < maxPoint.Y; i++ {
		pixelGrid[i] = make([]float64, maxPoint.X)
		for j := minPoint.X; j < maxPoint.X; j++ {
			pixelGrid[i][j] = RGBTo8BitGrayScaleIntensity(img.At(j, i))
		}
	}

	maskedGrid := FloodFillFromTopLeftCorner(pixelGrid, 5, 0.15)

	newImage := image.NewGray(img.Bounds())
	for y := minPoint.Y; y < maxPoint.Y; y++ {
		for x := minPoint.X; x < maxPoint.X; x++ {
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

// CreateGaussianBlurImage takes an image and applies Gaussian blur to it. It outputs a blurred
// image.
func CreateGaussianBlurImage(outputDir, imageName string, img image.Image) {
	maxPoint := img.Bounds().Max
	minPoint := img.Bounds().Min

	pixelGrid := make([][]float64, maxPoint.Y)
	for i := minPoint.Y; i < maxPoint.Y; i++ {
		pixelGrid[i] = make([]float64, maxPoint.X)
		for j := minPoint.X; j < maxPoint.X; j++ {
			pixelGrid[i][j] = RGBTo8BitGrayScaleIntensity(img.At(j, i))
		}
	}

	maskedGrid := ParallelGaussianMask(pixelGrid, 32)

	newImage := image.NewGray(img.Bounds())
	for y := minPoint.Y; y < maxPoint.Y; y++ {
		for x := minPoint.X; x < maxPoint.X; x++ {
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

// CreateEdgeDetectionImage takes an image and applies Canny's edge detection algorithm to it. It
// outputs an image that has edge highlighted.
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
	gradMask := ParallelGradientMask(gaussMask, 32)
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

// CreateAutoKeepoutImage takes an image and performs the whole set of auto keepout algorithm to it.
// The output is an image with obstacle groupings. The red dots represent the corners of a keepout
// polygon.
func CreateAutoKeepoutImage(outputDir string, imageName string, img image.Image) {
	maxPoint := img.Bounds().Max
	minPoint := img.Bounds().Min

	pixelGrid := make([][]float64, maxPoint.Y)
	for i := minPoint.Y; i < maxPoint.Y; i++ {
		pixelGrid[i] = make([]float64, maxPoint.X)
		for j := minPoint.X; j < maxPoint.X; j++ {
			pixelGrid[i][j] = RGBTo8BitGrayScaleIntensity(img.At(j, i))
		}
	}

	wallRemovedMask := FloodFillFromTopLeftCorner(pixelGrid, 5, 0.10)
	gaussMask := ParallelGaussianMask(wallRemovedMask, 4)
	gradMask := ParallelGradientMask(gaussMask, 4)
	NonMaximumSuppression(gradMask, 255)
	SimpleNearestNeighborClustering(gradMask, 10)
	hullMask := ConvexHullMasking(gradMask)

	newImage := image.NewNRGBA(img.Bounds())
	for y := minPoint.Y; y < maxPoint.Y; y++ {
		for x := minPoint.X; x < maxPoint.X; x++ {
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
