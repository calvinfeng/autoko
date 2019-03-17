package annotate

type Coordinate struct {
	I int
	J int
}

// IsInBound indicates whether the coordinate is in bound.
func (c *Coordinate) IsInBound(maxRow, maxCol int) bool {
	return (c.I >= 0 && c.I < maxRow) && (c.J >= 0 && c.J < maxCol)
}

// FloodFillVal is the value for flood filling, if it is 255.0 then it means the image gets flood
// filled with white.
const FloodFillVal = 255.0

// FloodFillFromTopLeftCorner uses breadth first approach to flood fill an image to get rid of
// exterior wall.
func FloodFillFromTopLeftCorner(mat [][]float64, neighborDist int, tolerance float64) [][]float64 {
	// Instantiate a mask that is an identical copy of the original mat
	mask := make([][]float64, len(mat))
	visitRecord := make([][]bool, len(mat))
	for i := 0; i < len(mat); i++ {
		visitRecord[i] = make([]bool, len(mat[i]))
		mask[i] = make([]float64, len(mat[i]))
		copy(mask[i], mat[i])
	}

	srcVal := mat[0][0]
	queue := []*Coordinate{&Coordinate{0, 0}}
	for len(queue) > 0 {
		c := queue[0]
		queue = queue[1:]
		if srcVal*(1.0-tolerance) <= mat[c.I][c.J] && mat[c.I][c.J] <= srcVal*(1.0+tolerance) {
			for i := c.I - neighborDist; i <= c.I+neighborDist; i++ {
				for j := c.J - neighborDist; j <= c.J+neighborDist; j++ {
					if i < 0 || i >= len(mat) {
						continue
					}

					if j < 0 || j >= len(mat[i]) {
						continue
					}

					if visitRecord[i][j] {
						continue
					}

					mask[i][j] = FloodFillVal
					visitRecord[i][j] = true
					queue = append(queue, &Coordinate{i, j})
				}
			}
		}
	}

	return mask
}
