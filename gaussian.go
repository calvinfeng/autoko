package autokeepout

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
)

// Kernel attributes, kernel size should always be odd and offset is the always kernel size minus
// one divide by two.
const (
	KernelSize = 5
	Offset     = (KernelSize - 1) / 2
)

// GaussKernel is used for applying Gaussian blur to an image.
var GaussKernel = [][]float64{
	{2.0, 4.0, 5.0, 4.0, 2.0},
	{4.0, 9.0, 12.0, 9.0, 4.0},
	{5.0, 12.0, 15.0, 12.0, 5.0},
	{4.0, 9.0, 12.0, 9.0, 4.0},
	{2.0, 4.0, 5.0, 4.0, 2.0},
}

func init() {
	var norm float64
	for i := 0; i < len(GaussKernel); i++ {
		for j := 0; j < len(GaussKernel[i]); j++ {
			norm += GaussKernel[i][j]
		}
	}

	for i := 0; i < len(GaussKernel); i++ {
		for j := 0; j < len(GaussKernel[i]); j++ {
			GaussKernel[i][j] /= norm
		}
	}
}

// Submask is part of a larger mask that is applied to the whole image.
type Submask struct {
	Order    int
	StartRow int
	Values   [][]float64
}

// GaussFilter assumes zero padding, i.e. if it is being operated on the corner of a mat,
// everything that is out of bound is assumed to be zero valued.
func GaussFilter(mat [][]float64, y, x int) float64 {
	return convolve(mat, y, x, 5, GaussKernel)
}

// GaussianMask applies Gaussian blur to an image matrix.
func GaussianMask(mat [][]float64) [][]float64 {
	maskedGrid := make([][]float64, len(mat))
	for i := 0; i < len(mat); i++ {
		maskedGrid[i] = make([]float64, len(mat[i]))
		for j := 0; j < len(mat[i]); j++ {
			maskedGrid[i][j] = GaussFilter(mat, i, j)
		}
	}

	return maskedGrid
}

// getGaussSubmask is called in the optimized version of gaussian masking. It is called in
// multiple go routines to achieve parallel convolution operations.
func getGaussSubmask(mat [][]float64, n, startRow, endRow int, output chan *Submask) {
	rowSize := endRow - startRow
	values := make([][]float64, rowSize)
	for i := 0; i < rowSize; i++ {
		colSize := len(mat[startRow+i])
		values[i] = make([]float64, colSize)
		for j := 0; j < colSize; j++ {
			values[i][j] = GaussFilter(mat, startRow+i, j)
		}
	}

	output <- &Submask{
		Order:    n,
		StartRow: startRow,
		Values:   values,
	}
}

// ParallelGaussianMask applies Gaussian blur to an image matrix using multiple subroutines to
// achieve parallelism.
func ParallelGaussianMask(mat [][]float64, numRoutines int) [][]float64 {
	rowsPerRoutine := len(mat) / numRoutines
	outputChan := make(chan *Submask, numRoutines)

	n := 0
	for n < numRoutines-1 {
		go getGaussSubmask(mat, n, n*rowsPerRoutine, (n+1)*rowsPerRoutine, outputChan)
		n++
	}

	go getGaussSubmask(mat, n, n*rowsPerRoutine, len(mat), outputChan)

	n = 0
	submasks := make([]*Submask, numRoutines)
	for submask := range outputChan {
		submasks[submask.Order] = submask
		n++

		if n == numRoutines {
			break
		}
	}

	mask := [][]float64{}
	for _, submask := range submasks {
		mask = append(mask, submask.Values...)
	}

	return mask
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
