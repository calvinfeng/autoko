package autokeepout

import (
	"image/color"
	"math/rand"
)

// RGBTo8BitGrayScaleIntensity converts RGB color into gray scale value. Color returns alpha
// pre-multiplied values, and thus the maximum value on this gray scale is 65535, not 255. We wish
// to convert the pre-multiplied value to 8-bit scale, so we divide the value by 257.
func RGBTo8BitGrayScaleIntensity(c color.Color) float64 {
	r, g, b, _ := c.RGBA()
	return (0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)) / 257
}

func randomMat(row, col int) [][]float64 {
	mat := make([][]float64, row)
	for i := 0; i < row; i++ {
		mat[i] = make([]float64, col)
		for j := 0; j < col; j++ {
			mat[i][j] = rand.Float64()
		}
	}

	return mat
}

func onesMat(row, col int) [][]float64 {
	mat := make([][]float64, row)
	for i := 0; i < row; i++ {
		mat[i] = make([]float64, col)
		for j := 0; j < col; j++ {
			mat[i][j] = 1.0
		}
	}

	return mat
}
