package annotate

import "testing"

func BenchmarkFloodFill(b *testing.B) {
	m := randomMat(1000, 1000)

	b.Run("FloodFillFromTopLeftCorner", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			FloodFillFromTopLeftCorner(m, 5, 0.10)
		}
	})
}
