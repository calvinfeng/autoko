package annotate

import (
	"fmt"
	"math"
)

// ConvexHullMasking returns a boolean map with [i][j] as keys. The boolean value indicates whether
// point at i, j is a polygon corner.
func ConvexHullMasking(grads [][]*Gradient) map[int]map[int]bool {
	hullMask := make(map[int]map[int]bool)

	clusters := make(map[int][]*Point)
	for i := 0; i < len(grads); i++ {
		for j := 0; j < len(grads[i]); j++ {
			if !grads[i][j].IsLocalMax {
				continue
			}

			if _, ok := clusters[grads[i][j].ClusterID]; !ok {
				clusters[grads[i][j].ClusterID] = []*Point{}
			}

			clusters[grads[i][j].ClusterID] = append(clusters[grads[i][j].ClusterID], &Point{false, i, j})
		}
	}

	for id := range clusters {
		LabelHullVertices(clusters[id])

		for _, point := range clusters[id] {
			if !point.IsHullVertex {
				continue
			}

			if _, ok := hullMask[point.Y]; !ok {
				hullMask[point.Y] = make(map[int]bool)
			}

			hullMask[point.Y][point.X] = true
		}
	}

	return hullMask
}

type Point struct {
	IsHullVertex bool
	Y, X         int
}

func (p *Point) String() string {
	return fmt.Sprintf("(Y:%v, X:%v)", p.Y, p.X)
}

func LabelHullVertices(points []*Point) {
	if len(points) == 0 {
		return
	}

	// Begin with finding the top-most point and bottom-most point, label them as vertices.
	high, low := findHighPointAndLowPoint(points)
	high.IsHullVertex, low.IsHullVertex = true, true

	// Draw a line with high and low points, and group the points into left/right using the line as
	// a divisor.
	leftGroup, rightGroup := []*Point{}, []*Point{}
	for _, point := range points {
		if point.IsHullVertex {
			// Ignore points that have already been labeled as Hull vertex
			continue
		}

		crossProduct := crossProduct(high, low, point)
		// Using right-hand rule, points that lie on the right-hand side of the division line has a
		// positive cross product value, and points that lie on the left-hand side has a negative
		// cross product.
		if crossProduct > 0 {
			rightGroup = append(rightGroup, point)
		} else {
			leftGroup = append(leftGroup, point)
		}
	}

	recursiveLabeling(leftGroup, high, low)
	recursiveLabeling(rightGroup, low, high)
}

func recursiveLabeling(points []*Point, source, target *Point) {
	if len(points) == 0 {
		return
	}

	farthestAwayPoint := points[0]
	maxDist := perpendicularDistanceFromLine(source, target, farthestAwayPoint)
	for _, point := range points {
		dist := perpendicularDistanceFromLine(source, target, point)
		if dist > maxDist {
			farthestAwayPoint = point
			maxDist = dist
		}
	}

	farthestAwayPoint.IsHullVertex = true

	// Using source, target, and farthestAwayPoint to create a triangle. Any point that lies inside
	// this triangle can be discarded because they are for sure not the hull vertices. There will be
	// two candidate set.
	leftCandidates, rightCandidates := []*Point{}, []*Point{}
	for _, point := range points {
		if point.IsHullVertex {
			continue
		}

		leftCrossProduct := crossProduct(target, farthestAwayPoint, point)
		if leftCrossProduct > 0 {
			leftCandidates = append(leftCandidates, point)
			continue
		}

		rightCrossProduct := crossProduct(farthestAwayPoint, source, point)
		if rightCrossProduct > 0 {
			rightCandidates = append(rightCandidates, point)
			continue
		}
	}

	recursiveLabeling(leftCandidates, farthestAwayPoint, target)
	recursiveLabeling(rightCandidates, source, farthestAwayPoint)
}

// findHighPointAndLowPoint returns two points that are highest and lowest in a set of points.
// Highest indicates that it has the minimum y-value because grid coordinates are upside-down from
// cartesian coordinate. Lowest indicates that it has the maximum y-value.
func findHighPointAndLowPoint(points []*Point) (high, low *Point) {
	if len(points) == 0 {
		return nil, nil
	}

	high, low = points[0], points[0]
	for _, point := range points {
		if point.Y < high.Y {
			high = point
		} else if point.Y > low.Y {
			low = point
		}
	}

	return high, low
}

// Source and target points form a line. perpendicularDistanceFromLine returns the distance from the
// point to the line that was formed by source and target.
func perpendicularDistanceFromLine(source, target, point *Point) float64 {
	temp := (target.Y - source.Y) * point.X
	temp -= (target.X - source.X) * point.Y
	temp += target.X * source.Y
	temp -= target.Y * source.X
	dist := math.Abs(float64(temp))
	norm := math.Sqrt(float64((target.Y-source.Y)*(target.Y-source.Y) + (target.X-source.X)*(target.X-source.X)))

	return dist / norm
}

func crossProduct(source, target, point *Point) int {
	// Create a vector that goes from source to target
	v1X := target.X - source.X
	v1Y := target.Y - source.Y

	// Create a second vector that goes from source to point
	v2X := point.X - source.X
	v2Y := point.Y - source.Y

	// Perform vector cross product
	return (v1X * v2Y) - (v1Y * v2X)
}
