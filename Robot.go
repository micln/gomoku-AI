package gomoku_AI

import "log"

type Robot struct {
	*Board
}

func NewRobot() *Robot {
	return &Robot{
		Board: NewBoard(),
	}
}

//	当人类下 x,y 时，Robot要走哪里
func (r *Robot) HumanGo(x, y int) *Point {

	TimeStart(`HumanGo`)
	defer TimerEnd(`HumanGo`)

	log.Printf("Human %d,%d\n", x, y)

	r.Board.GoChess(NewPoint(x, y), C_Player)
	p := r.Board.MaxMin(4)
	r.Board.GoChess(p, C_Robot)

	log.Printf("Robot %d,%d\n", p.X, p.Y)

	return p
}
