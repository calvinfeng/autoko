package main

type MathError struct {
	Message string
}

func (m MathError) Error() string {
	return m.Message
}
