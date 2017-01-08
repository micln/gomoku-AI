package gomoku_AI

import (
	"container/list"
	"errors"
	"fmt"
	"log"
	"sync"
)

type ChessType byte
type MapType [bSize][bSize]ChessType

const (
	bSize = 15

	MAX = 1 << 30
	MIN = -MAX
)

const (
	C_Robot  ChessType = 1
	C_Player ChessType = 2
)

var (
	ErrDuplicate = errors.New(`duplicate chess`)
)

type Board struct {
	mp   MapType
	size int

	history []*Step
}

func NewBoard() *Board {
	return &Board{
		size: bSize,
	}
}

//	用来并发处理
func (b *Board) Copy() *Board {
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

//
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

func (b *Board) GoXY(x, y int, chess ChessType) error {
	return b.GoChess(NewPoint(x, y), chess)
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

	dx := []int{0, 1, 1, -1}
	dy := []int{1, 0, 1, 1}

	allMax := 0

	for i := 0; i < b.size; i++ {
		for j := 0; j < b.size; j++ {
			if b.HasXYIs(i, j, c) {
				for k := 0; k < len(dx); k++ {
					max := 1
					for r := 1; r < 5; r++ {
						if !b.HasXYIs(i+dx[k]*r, j+dy[k]*r, c) {
							break
						}
						max++
					}
					if max > allMax {
						allMax = max
					}
				}
			}
		}
	}
	scores := []int{0, 10, 100, 1000, 10000, 10000}

	result = scores[allMax]

	//fmt.Printf("c[%v]:%d\n", c, result)
	return
}

//	AI 挑一个最高分的位置
func (b *Board) BestStep(dep int) *Point {
	best := MIN
	results := list.New()
	points := b.genAvailablePoints()

	l := sync.Mutex{}
	tryPoint := func(v int, p *Point) {
		l.Lock()
		defer l.Unlock()

		fmt.Printf("In(%v) score:%d\n", p, v)

		if v < best {
			return
		}

		if v > best {
			best = v
			results.Init()
		}

		results.PushBack(p)
	}

	wg := sync.WaitGroup{}
	wg.Add(len(points))

	for idx := range points {
		go func(p *Point) {
			defer wg.Done()

			nb := b.Copy()
			nb.GoChess(p, C_Robot)
			defer nb.GoBack()

			v := nb.playerDfs(dep - 1)

			tryPoint(v, p)

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
func (b *Board) aiDfs(dep int) int {
	v := b.Evaluation()
	if dep <= 0 {
		return v
	}

	best := MIN
	l := sync.Mutex{}
	tryV := func(v int) {
		l.Lock()
		defer l.Unlock()

		if v > best {
			best = v
		}
	}
	points := b.genAvailablePoints()

	wg := sync.WaitGroup{}
	wg.Add(len(points))

	for idx := range points {
		go func(p *Point) {
			defer wg.Done()

			nb := b.Copy()
			nb.GoChess(p, C_Robot)
			defer nb.GoBack()

			v := nb.playerDfs(dep - 1)
			tryV(v)

		}(points[idx])
	}

	wg.Wait()

	return best
}

//	Player 走最优解后，当前局势的评分
//	evaluation 最小
func (b *Board) playerDfs(dep int) int {
	v := b.Evaluation()
	if dep <= 0 {
		return v
	}

	best := MAX
	l := sync.Mutex{}
	tryV := func(v int) {
		l.Lock()
		defer l.Unlock()

		if v < best {
			best = v
		}
	}
	points := b.genAvailablePoints()

	wg := sync.WaitGroup{}
	wg.Add(len(points))

	for idx := range points {
		go func(p *Point) {
			defer wg.Done()

			nb := b.Copy()
			nb.GoChess(p, C_Player)
			defer nb.GoBack()

			v := nb.aiDfs(dep - 1)
			tryV(v)

		}(points[idx])
	}

	wg.Wait()

	return best
}

//	当前棋子在每个位置落子后当前局势分
func (b *Board) CalcScoreMaps(c ChessType) [bSize][bSize]int {
	maps := [bSize][bSize]int{}

	for i := 0; i < b.size; i++ {
		for j := 0; j < b.size; j++ {
			if b.HasChessInXY(i, j) {
				continue
			}

			b.GoXY(i, j, c)
			maps[i][j] = b.Evaluation()

			b.GoBack()
		}
	}

	return maps
}
