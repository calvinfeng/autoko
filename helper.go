package main

type Position struct {
	I int
	J int
}

func (p *Position) IsOutOfBound(maxRow, maxCol int) bool {
	if p.I < 0 || p.I >= maxRow {
		return true
	}

	if p.J < 0 || p.J >= maxCol {
		return true
	}

	return false
}
