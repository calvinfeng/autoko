package autokeepout

type Coordinate struct {
	I int
	J int
}

// IsInBound indicates whether the coordinate is in bound.
func (c *Coordinate) IsInBound(maxRow, maxCol int) bool {
	return (c.I >= 0 && c.I < maxRow) && (c.J >= 0 && c.J < maxCol)
}
