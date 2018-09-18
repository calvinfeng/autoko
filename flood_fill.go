package autokeepout

// FloodFillFromTopLeftCorner uses breadth first approach to flood fill an image to get rid of
// exterior wall.
func FloodFillFromTopLeftCorner(mat [][]float64, neighborDist int) [][]float64 {
	// Instantiate a mask that is an identical copy of the original mat
	mask := make([][]float64, len(mat))
	visitRecord := make([][]bool, len(mat))
	for i := 0; i < len(mat); i++ {
		visitRecord[i] = make([]bool, len(mat[i]))
		mask[i] = make([]float64, len(mat[i]))
		copy(mask[i], mat[i])
	}

	srcCoord := &Coordinate{0, 0}
	sourceValue := mat[srcCoord.I][srcCoord.J]
	targetValue := 255.0
	breadthFirst(srcCoord, neighborDist, mat, mask, visitRecord, sourceValue, targetValue)

	return mask
}

func breadthFirst(c *Coordinate, dist int, mat, mask [][]float64, visitRecord [][]bool, srcVal, tgtVal float64) {
	queue := []*Coordinate{c}
	mask[c.I][c.J] = tgtVal
	visitRecord[c.I][c.J] = true
	for len(queue) > 0 {
		coord := queue[0]
		queue = queue[1:]

		// Expand to neighboring pixels only if the current pixel on the mat matches the source
		// value.
		if mat[coord.I][coord.J] == srcVal {
			for i := coord.I - dist; i <= coord.I+dist; i++ {
				for j := coord.J - dist; j <= coord.J+dist; j++ {
					if i < 0 || i >= len(mat) {
						continue
					}

					if j < 0 || j >= len(mat[i]) {
						continue
					}

					if visitRecord[i][j] {
						continue
					}

					if i == coord.I && j == coord.J {
						continue
					}

					queue = append(queue, &Coordinate{i, j})
					mask[i][j] = tgtVal
					visitRecord[i][j] = true
				}
			}
		}
	}
}
