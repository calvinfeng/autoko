package autokeepout

import (
	"testing"
)

func BenchmarkGaussianMask(b *testing.B) {
	m := randomMat(1000, 1000)

	b.Run("RegularGaussianMask", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			GaussianMask(m)
		}
	})

	b.Run("ParallelGaussianMask with 1 go-routine", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ParallelGaussianMask(m, 1)
		}
	})

	b.Run("ParallelGaussianMask with 2 go-routines", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ParallelGaussianMask(m, 2)
		}
	})

	b.Run("ParallelGaussianMask with 4 go-routines", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ParallelGaussianMask(m, 4)
		}
	})

	b.Run("ParallelGaussianMask with 32 go-routines", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ParallelGaussianMask(m, 32)
		}
	})
}

func TestGaussFilter(t *testing.T) {
	img := [][]float64{
		{0, 0, 0, 0, 0},
		{0, 1, 1, 1, 0},
		{0, 1, 1, 1, 0},
		{0, 1, 1, 1, 0},
		{0, 0, 0, 0, 0},
	}

	t.Run("ApplyToCenter", func(t *testing.T) {
		result := gaussFilter(img, 2, 2)

		var expected float64
		for i := 0; i < KernelSize; i++ {
			for j := 0; j < KernelSize; j++ {
				expected += GaussKernel[i][j] * img[i][j]
			}
		}

		if result != expected {
			t.Errorf("incorrect Gaussian filter result %f", result)
		}
	})

	t.Run("ApplyToCorner", func(t *testing.T) {
		result := gaussFilter(img, 0, 0)

		var expected float64
		for i := 2; i < KernelSize; i++ {
			for j := 2; j < KernelSize; j++ {
				expected += GaussKernel[i][j] * img[i-2][j-2]
			}
		}

		if result != expected {
			t.Errorf("incorrect Gaussian filter result %f", result)
		}
	})
}
