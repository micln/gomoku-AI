package gomoku_AI

import (
	"container/list"
	"errors"
	"log"
	"sync"
)

type ChessType byte
type MapType [bSize][bSize]ChessType

const (
	bSize = 15
)

const (
	C_Robot ChessType = iota
	C_Player
)

var (
	ErrDuplicate = errors.New(`duplicate chess`)
)

type Board struct {
	mp   MapType
	size int

	history []*Step
}

type Step struct {
	*Point
	Chess ChessType
}

func NewBoard() *Board {
	return &Board{
		size: bSize,
	}
}

//	用来并发处理
func (b *Board) copy() *Board {
	return &Board{
		mp:      b.mp,
		size:    b.size,
		history: b.history,
	}
}

func (b *Board) Size() int {
	return b.size
}

func (b *Board) Map() MapType {
	return b.mp
}

func (b *Board) History() []*Step {
	return b.history
}

func isValidPoint(x, y int) bool {
	return x >= 0 && x < bSize && y >= 0 && y < bSize
}

func (b *Board) HasXYIs(x, y int, c ChessType) bool {
	return b.HasChessInXY(x, y) && b.mp[x][y] == c
}

func (b *Board) HasChessInXY(x, y int) bool {
	if isValidPoint(x, y) {
		return b.mp[x][y] != 0
	}
	return false
}

func (b *Board) HasChessInPoint(p *Point) bool {
	return b.HasChessInXY(p.X, p.Y)
}

func (b *Board) genAvailablePoints() []*Point {

	if len(b.history) == 0 {
		return []*Point{NewPoint(7, 7)}
	}

	resultMap := [bSize][bSize]bool{}

	width := 1

	for si := range b.history {
		p := b.history[si].Point
		for i := -width; i <= width; i++ {
			for j := -width; j <= width; j++ {
				xx := p.X + i
				yy := p.Y + j
				if isValidPoint(xx, yy) {
					if !b.HasChessInXY(xx, yy) {
						resultMap[xx][yy] = true
					}
				}
			}
		}
	}

	results := []*Point{}
	for i := 0; i < bSize; i++ {
		for j := 0; j < bSize; j++ {
			if resultMap[i][j] {
				results = append(results, NewPoint(i, j))
			}
		}
	}

	return results
}

//	落子
func (b *Board) GoChess(p *Point, chess ChessType) error {
	if b.HasChessInPoint(p) {
		panic(ErrDuplicate)
	}

	b.mp[p.X][p.Y] = chess

	b.history = append(b.history, &Step{
		p,
		chess,
	})

	return nil
}

//	撤回一步棋
func (b *Board) GoBack() bool {
	if len(b.history) == 0 {
		panic(`Unknow Goback`)
	}

	last := b.history[len(b.history)-1]
	b.mp[last.X][last.Y] = 0

	b.history = b.history[:len(b.history)-1]
	return true
}

//	当前局势评分
func (b *Board) Evaluation() int {
	return b.scoreFor(C_Robot) - b.scoreFor(C_Player)
}

func (b *Board) scoreFor(c ChessType) (result int) {

	scores := []int{0, 10, 100, 1000, 10000, 10000}

	dx := []int{0, 1, 1}
	dy := []int{1, 0, 1}

	for i := 0; i < b.size; i++ {
		for j := 0; j < b.size; j++ {
			if b.HasXYIs(i, j, c) {
				max := 1
				for k := 0; k < 3; k++ {
					if !b.HasXYIs(i+dx[k], j+dy[k], c) {
						break
					}
					max++
				}
				max = scores[max]
				result += max
			}
		}
	}

	return
}

//	AI 挑一个最高分的位置
func (b *Board) MaxMin(dep int) *Point {
	best := 0
	results := list.New()
	points := b.genAvailablePoints()

	wg := sync.WaitGroup{}
	wg.Add(len(points))

	for idx := range points {
		go func(p *Point) {
			defer wg.Done()

			nb := b.copy()
			nb.GoChess(p, C_Robot)
			defer nb.GoBack()

			v := nb.min(dep - 1)
			if v < best {
				return
			}

			if v > best {
				best = v
				results.Init()
			}

			results.PushBack(p)
		}(points[idx])
	}

	wg.Wait()
	log.Printf("AI (%d) paths: %v\n", len(points), points)
	log.Printf("AI (%d) bests\n", results.Len())

	if results.Len() == 0 {
		panic("no way")
	}

	return results.Front().Value.(*Point)
}

//	AI 在当前局势下能拿到的最高分
func (b *Board) maxAI(dep int) int {
	v := b.Evaluation()
	if dep <= 0 {
		return v
	}

	best := 0
	points := b.genAvailablePoints()

	wg := sync.WaitGroup{}
	wg.Add(len(points))

	for idx := range points {
		go func(p *Point) {
			defer wg.Done()

			nb := b.copy()
			nb.GoChess(p, C_Robot)
			defer nb.GoBack()

			v := nb.min(dep - 1)
			if v > best {
				best = v
			}
		}(points[idx])
	}

	wg.Wait()

	return best
}

//	Player 走最优解后，当前局势的评分
//	evaluation 最小
func (b *Board) min(dep int) int {
	v := b.Evaluation()
	if dep <= 0 {
		return v
	}

	best := 1 << 30
	points := b.genAvailablePoints()

	wg := sync.WaitGroup{}
	wg.Add(len(points))

	for idx := range points {
		go func(p *Point) {
			defer wg.Done()

			nb := b.copy()
			nb.GoChess(p, C_Player)
			defer nb.GoBack()

			v := nb.maxAI(dep - 1)
			if v < best {
				best = v
			}
		}(points[idx])
	}

	wg.Wait()

	return best
}
