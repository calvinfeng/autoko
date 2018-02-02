package main

type Coordinate struct {
	I int
	J int
}

func (c *Coordinate) IsOutOfBound(maxRow, maxCol int) bool {
	if c.I < 0 || c.I >= maxRow {
		return true
	}

	if c.J < 0 || c.J >= maxCol {
		return true
	}

	return false
}
