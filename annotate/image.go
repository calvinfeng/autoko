package annotate

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
func CreateFloodFillImage(outputDir, imageName string, img image.Image) {
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

	wallRemovedMask := FloodFillFromTopLeftCorner(pixelGrid, 5, 0.10)
	gaussMask := ParallelGaussianMask(wallRemovedMask, 4)

	newImage := image.NewGray(img.Bounds())
	for y := minPoint.Y; y < maxPoint.Y; y++ {
		for x := minPoint.X; x < maxPoint.X; x++ {
			val := gaussMask[y][x]
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

	wallRemovedMask := FloodFillFromTopLeftCorner(pixelGrid, 5, 0.10)
	gaussMask := ParallelGaussianMask(wallRemovedMask, 32)
	gradMask := ParallelGradientMask(gaussMask, 32)
	NonMaximumSuppression(gradMask, 255)

	newImage := image.NewNRGBA(img.Bounds())
	for y := minPoint.Y; y < maxPoint.Y; y++ {
		for x := minPoint.X; x < maxPoint.X; x++ {
			grad := gradMask[y][x]
			if grad.IsLocalMax {
				newImage.Set(x, y, color.NRGBA{255, 0, 0, 255})
			} else {
				newImage.Set(x, y, color.Gray{uint8(gaussMask[y][x])})
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

// CreateClusteringImage takes an image and performs the nearest neighbor clustering algorithm to it.
// The output is an image with different clusters where each cluster is an obstacle.
func CreateClusteringImage(outputDir, imageName string, img image.Image) {
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

	newImage := image.NewNRGBA(img.Bounds())
	for y := minPoint.Y; y < maxPoint.Y; y++ {
		for x := minPoint.X; x < maxPoint.X; x++ {
			grad := gradMask[y][x]
			if grad.IsLocalMax {
				newImage.Set(x, y, Colors[grad.ClusterID%len(Colors)])
			} else {
				val := uint8(gaussMask[y][x])
				if val < 0 {
					val = 0
				}
				newImage.Set(x, y, color.NRGBA{val, val, val, 255})
			}
		}
	}

	outputFile, fileErr := os.Create(fmt.Sprintf("%s/%s_clustering.png", outputDir, imageName))
	if fileErr != nil {
		fmt.Println("Cannot create image")
	} else {
		png.Encode(outputFile, newImage)
		outputFile.Close()
	}
}

// CreateConvexHullImage takes an image and performs the whole set of auto keepout algorithm to it.
// The output is an image with obstacle groupings. The red dots represent the convex hull corners of
// a keepout polygon.
func CreateConvexHullImage(outputDir, imageName string, img image.Image) {
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

	newImage := image.NewNRGBA(img.Bounds())
	for y := minPoint.Y; y < maxPoint.Y; y++ {
		for x := minPoint.X; x < maxPoint.X; x++ {
			grad := gradMask[y][x]
			if grad.IsLocalMax {
				newImage.Set(x, y, Colors[grad.ClusterID%len(Colors)])
			} else {
				val := uint8(gaussMask[y][x])
				if val < 0 {
					val = 0
				}
				newImage.Set(x, y, color.NRGBA{val, val, val, 255})
			}
		}
	}

	radius := 2

	hullMask := ConvexHullMasking(gradMask)
	for i := range hullMask {
		for j := range hullMask[i] {
			for y := i - radius; y < i+radius; y++ {
				for x := j - radius; x < j+radius; x++ {
					newImage.Set(x, y, color.NRGBA{255, 0, 0, 255})
				}
			}
		}
	}

	outputFile, fileErr := os.Create(fmt.Sprintf("%s/%s_convex_hull.png", outputDir, imageName))
	if fileErr != nil {
		fmt.Println("Cannot create image")
	} else {
		png.Encode(outputFile, newImage)
		outputFile.Close()
	}
}

func CreateSubtractMeanImage(outputDir string, imageName string, img image.Image) {
	maxPoint := img.Bounds().Max
	minPoint := img.Bounds().Min

	mean := 0.0
	for y := minPoint.Y; y < maxPoint.Y; y++ {
		for x := minPoint.X; x < maxPoint.X; x++ {
			mean += RGBTo8BitGrayScaleIntensity(img.At(x, y))
		}
	}
	mean = mean / float64(maxPoint.Y*maxPoint.X)

	newImage := image.NewGray(img.Bounds())
	for y := minPoint.Y; y < maxPoint.Y; y++ {
		for x := minPoint.X; x < maxPoint.X; x++ {
			val := RGBTo8BitGrayScaleIntensity(img.At(x, y)) - mean
			if val < 0.0 {
				val = 0.0
			}

			newImage.Set(x, y, color.Gray{uint8(val)})
		}
	}

	outputFile, fileErr := os.Create(fmt.Sprintf("%s/%s_subtracted_mean.png", outputDir, imageName))
	if fileErr != nil {
		fmt.Println("Cannot create image")
	} else {
		png.Encode(outputFile, newImage)
		outputFile.Close()
	}
}

func CreateColorfulImage(outputDir string, imageName string, img image.Image) {
	maxPoint := img.Bounds().Max
	minPoint := img.Bounds().Min

	newImage := image.NewRGBA(img.Bounds())
	for y := minPoint.Y; y < maxPoint.Y; y++ {
		for x := minPoint.X; x < maxPoint.X; x++ {
			val := RGBTo8BitGrayScaleIntensity(img.At(x, y))
			if val > 100 {
				newImage.Set(x, y, color.RGBA{0, 0, 0, 255})
			} else {
				newImage.Set(x, y, color.RGBA{255, 0, 0, 255})
			}
		}
	}

	outputFile, fileErr := os.Create(fmt.Sprintf("%s/%s_color.png", outputDir, imageName))
	if fileErr != nil {
		fmt.Println("Cannot create image")
	} else {
		png.Encode(outputFile, newImage)
		outputFile.Close()
	}
}
