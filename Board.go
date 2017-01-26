package gomoku_AI

import (
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
	ErrNoHistory = errors.New(`Empty history`)
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

//	主要用来并发搜索，不记 history了暂时
func (b *Board) Copy() *Board {
	return &Board{
		mp:   b.mp,
		size: b.size,
		//history: b.history,
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

func (b *Board) FillMap(m string) {

}

func isValidPoint(x, y int) bool {
	return x >= 0 && x < bSize && y >= 0 && y < bSize
}

func (b *Board) HasXYIs(x, y int, c ChessType) bool {
	return b.HasXY(x, y) && b.mp[x][y] == c
}

func (b *Board) HasXY(x, y int) bool {
	if isValidPoint(x, y) {
		return b.mp[x][y] != 0
	}
	return false
}

func (b *Board) HasPoint(p *Point) bool {
	return b.HasXY(p.X, p.Y)
}

//	生成可用路径
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
					if !b.HasXY(xx, yy) {
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
	if b.HasPoint(p) {
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
		panic(ErrNoHistory)
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

/**

 */
func (b *Board) scoreForV2(c ChessType) (result int) {

	return
}

/**
老算法：连着的几个子会重复计算几次
*/
func (b *Board) scoreFor(c ChessType) int {

	dx := []int{0, 1, 1, -1}
	dy := []int{1, 0, 1, 1}

	maxChess := 0

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
					if max > maxChess {
						maxChess = max
					}
				}
			}
		}
	}
	scores := []int{0, 10, 100, 1000, 10000, 10000}

	return scores[maxChess]
}

//	AI 挑一个最高分的位置
//	极大值极小值算法
//	dep 最小为1
func (b *Board) BestStep(dep int) *Point {
	best := MIN
	results := NewPoints()
	points := b.genAvailablePoints()

	l := sync.Mutex{}
	wg := sync.WaitGroup{}
	wg.Add(len(points))

	for idx := range points {
		go func(p *Point) {
			defer wg.Done()

			nb := b.Copy()
			nb.GoChess(p, C_Robot)

			v := nb.playerDfs(dep - 1)

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

		}(points[idx])
	}

	wg.Wait()
	log.Printf("AI %d paths: %v\n", len(points), points)
	log.Printf("%d bests(%v): %v\n", results.Len(), best, results)

	if results.Len() == 0 {
		panic("no way")
	}

	return results.Front().Value.(*Point)
}

//	AI 在当前局势下能拿到的最高分
//	求极大值
func (b *Board) aiDfs(dep int) int {
	v := b.Evaluation()
	if dep <= 0 {
		return v
	}

	best := MIN
	l := sync.Mutex{}
	points := b.genAvailablePoints()

	wg := sync.WaitGroup{}
	wg.Add(len(points))

	for idx := range points {
		go func(p *Point) {
			defer wg.Done()

			nb := b.Copy()
			nb.GoChess(p, C_Robot)

			v := nb.playerDfs(dep - 1)

			l.Lock()
			defer l.Unlock()

			if v > best {
				best = v
			}

		}(points[idx])
	}

	wg.Wait()

	return best
}

//	Player 走最优解后，当前局势的评分
//	极小值
func (b *Board) playerDfs(dep int) int {
	v := b.Evaluation()
	if dep <= 0 {
		return v
	}

	best := MAX
	l := sync.Mutex{}
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

			l.Lock()
			defer l.Unlock()

			if v < best {
				best = v
			}

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
			if b.HasXY(i, j) {
				continue
			}

			b.GoXY(i, j, c)
			maps[i][j] = b.Evaluation()
			b.GoBack()
		}
	}

	return maps
}
