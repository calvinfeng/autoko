package autokeepout

import (
	"math/rand"
	"testing"
)

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

func BenchmarkGaussianMask(b *testing.B) {
	mat := randomMat(100, 100)

	b.Run("RegularGaussianMask", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			GaussianMasking(mat)
		}
	})

	b.Run("ParallelGaussianMask with 1 go-routine", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ParallelGaussianMasking(mat, 1)
		}
	})

	b.Run("ParallelGaussianMask with 2 go-routines", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ParallelGaussianMasking(mat, 2)
		}
	})

	b.Run("ParallelGaussianMask with 4 go-routines", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ParallelGaussianMasking(mat, 4)
		}
	})

	b.Run("ParallelGaussianMask with 32 go-routines", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ParallelGaussianMasking(mat, 32)
		}
	})
}

func BenchmarkConvolve(b *testing.B) {
	mat := randomMat(10, 10)

	for i := 0; i < b.N; i++ {
		convolve(mat, 5, 5, GaussianKernel)
	}
}
