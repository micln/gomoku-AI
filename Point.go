package gomoku_AI

import "fmt"

type Point struct {
	X int
	Y int

	Hash int
}

func NewPoint(x, y int) *Point {
	return &Point{
		X:    x,
		Y:    y,
		Hash: x*20 + y,
	}
}

func (p *Point) String() string {
	return fmt.Sprintf("(%d,%d)", p.X, p.Y)
}
