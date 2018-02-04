package main

import "fmt"

// SimpleNearestNeighborClustering performs clustering based on concept of connected component. This function will only
// look at local maximum gradients with magnitude greater than 255. Two selected gradients are considered neighbors if
// they are within a certain range.
func SimpleNearestNeighborClustering(gradGrid [][]*Gradient, neighborRange int) {
	visitRecord := make([][]bool, len(gradGrid))
	for i := 0; i < len(visitRecord); i += 1 {
		visitRecord[i] = make([]bool, len(gradGrid[i]))
	}

	clusterID := 1
	for i := 0; i < len(gradGrid); i += 1 {
		for j := 0; j < len(gradGrid[i]); j += 1 {
			if visitRecord[i][j] {
				continue
			}

			if gradGrid[i][j].IsLocalMax {
				gradGrid[i][j].ClusterID = clusterID
				visitRecord[i][j] = true
				depthFirstNeighborClusterLabel(i, j, clusterID, neighborRange, gradGrid, visitRecord)
				fmt.Println("Completed labeling group", clusterID)
				clusterID += 1
			}
		}
	}
}

func depthFirstNeighborClusterLabel(y, x, id, neighborRange int, gradGrid [][]*Gradient, visitRecord [][]bool) {
	for i := y - neighborRange; i <= y+neighborRange; i += 1 {
		for j := x - neighborRange; j <= x+neighborRange; j += 1 {
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
