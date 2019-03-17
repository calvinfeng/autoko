package annotate

// SimpleNearestNeighborClustering performs clustering based on concept of connected component. This function will only
// look at local maximum gradients with magnitude greater than 255. Two selected gradients are considered neighbors if
// they are within a certain range.
func SimpleNearestNeighborClustering(gradGrid [][]*Gradient, neighborRange int) {
	visitRecord := make([][]bool, len(gradGrid))
	for i := 0; i < len(visitRecord); i++ {
		visitRecord[i] = make([]bool, len(gradGrid[i]))
	}

	clusterID := 1
	for i := 0; i < len(gradGrid); i++ {
		for j := 0; j < len(gradGrid[i]); j++ {
			if visitRecord[i][j] {
				continue
			}

			if gradGrid[i][j].IsLocalMax {
				gradGrid[i][j].ClusterID = clusterID
				visitRecord[i][j] = true
				depthFirstNeighborClusterLabel(i, j, clusterID, neighborRange, gradGrid, visitRecord)
				clusterID++
			}
		}
	}
}

func depthFirstNeighborClusterLabel(y, x, id, neighborRange int, gradGrid [][]*Gradient, visitRecord [][]bool) {
	for i := y - neighborRange; i <= y+neighborRange; i++ {
		for j := x - neighborRange; j <= x+neighborRange; j++ {
			if i < 0 || i >= len(gradGrid) {
				continue
			}

			if j < 0 || j >= len(gradGrid[i]) {
				continue
			}

			if visitRecord[i][j] {
				continue
			}

			if gradGrid[i][j].IsLocalMax {
				gradGrid[i][j].ClusterID = id
				visitRecord[i][j] = true
				depthFirstNeighborClusterLabel(i, j, id, neighborRange, gradGrid, visitRecord)
			}
		}
	}
}
