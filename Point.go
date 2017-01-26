package gomoku_AI

import (
	"container/list"
	"fmt"
)

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

func (p Point) String() string {
	return fmt.Sprintf("(%d,%d)", p.X, p.Y)
}

type Points struct {
	*list.List
}

func NewPoints() *Points {
	return &Points{list.New()}
}

func (ps *Points) Each(fc func(*list.Element)) {
	for e := ps.Front(); e != nil; e = e.Next() {
		fc(e)
	}
}

func (ps Points) String() string {
	s := ""
	ps.Each(func(e *list.Element) {
		s = s + e.Value.(*Point).String() + ` `
	})
	return s
}
