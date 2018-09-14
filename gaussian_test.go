package autokeepout

import (
	"math/rand"
	"testing"
)

func BenchmarkGaussianMask(b *testing.B) {
	grid := make([][]float64, 1000)
	for i := 0; i < len(grid); i++ {
		grid[i] = make([]float64, 1000)
		for j := 0; j < len(grid[i]); j++ {
			grid[i][j] = rand.Float64()
		}
	}

	b.Run("RegularGaussianMask", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			GaussianMasking(grid)
		}
	})

	b.Run("8ParallelGaussianMask", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ParallelGaussianMasking(grid, 8)
		}
	})

	b.Run("32ParallelGaussianMask", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ParallelGaussianMasking(grid, 32)
		}
	})

	b.Run("128ParallelGaussianMask", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ParallelGaussianMasking(grid, 128)
		}
	})
}
