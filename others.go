package autokeepout

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
)

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
