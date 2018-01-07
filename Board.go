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
	ErrNoHistory = errors.New(`empty history`)
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

func (b *Board) Each(f func(x, y int)) {
	for i := 0; i < bSize; i++ {
		for j := 0; j < bSize; j++ {
			f(i, j)
		}
	}
}

//	生成可用路径
// 已落的子周围1个距离内，都能走
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

	var results []*Point
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

// 当前局势评分
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
// 计算当前局势对 c 的分值
// 算法：c现在最长的连子长度 * 权重
func (b *Board) scoreFor(c ChessType) int {

	dx := []int{0, 1, 1, -1}
	dy := []int{1, 0, 1, 1}

	maxChess := 0

	for i := 0; i < b.size; i++ {
		for j := 0; j < b.size; j++ {
			if b.HasXYIs(i, j, c) {
				for k := 0; k < len(dx); k++ {
					maxLenForIJ := 1
					for r := 1; r < 5; r++ {
						if !b.HasXYIs(i+dx[k]*r, j+dy[k]*r, c) {
							break
						}
						maxLenForIJ++
					}
					if maxLenForIJ > maxChess {
						maxChess = maxLenForIJ
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
// dep 表示 ai 能往后算几步，如果为0，表示不思考，这一步能落子的地方都能走
func (b *Board) BestPoint(dep int) *Point {
	found := MIN
	results := NewPoints()
	points := b.genAvailablePoints()

	lth := len(points)

	ch := make(chan *Point, lth)
	pv := [400]int{}

	for idx := range points {
		go func(p *Point) {
			nb := b.Copy()
			nb.GoChess(p, C_Robot)

			pp, v := nb.playerDfs(dep)
			fmt.Printf("IF AI take %v->%v, then player take %v->%v\n", p, nb.Evaluation(), pp, v)

			pv[pp.Hash] = v
			ch <- pp
		}(points[idx])
	}

	var p *Point
	for i := 0; i < lth; i++ {
		p = <-ch
		v := pv[p.Hash]

		if v < found {
			continue
		}

		if v > found {
			found = v
			results.Init()
		}

		results.PushBack(p)
	}

	log.Printf("AI 一共有 %d points: %v\n", lth, points)
	log.Printf("其中，%d bests(%v): %v\n", results.Len(), found, results)

	if results.Len() == 0 {
		panic("no way")
	}

	result := results.Front().Value.(*Point)
	log.Println("最终落子：", result)

	return result
}

//	AI 在当前局势下能拿到的最高分
//	求极大值
func (b *Board) aiDfs(dep int) (rp *Point, v int) {
	v = b.Evaluation()
	if dep <= 0 {
		return
	}

	//	分越高，robot越喜欢
	best := MIN
	points := b.genAvailablePoints()

	nb := b.Copy()

	for idx := range points {
		p := points[idx]
		nb.GoChess(p, C_Robot)
		_, v = nb.playerDfs(dep - 1)

		if v > best {
			rp = p
			best = v
		}
	}
	return
}

//	Player 走最优解后，当前局势的评分
//	极小值算法
//	返回值：player 落子位置和分值
func (b *Board) playerDfs(dep int) (rp *Point, v int) {
	v = b.Evaluation()
	if dep <= 0 {
		return
	}

	//	对玩家来说，棋局打分越低越好，每次都挑最低分
	found := MAX
	points := b.genAvailablePoints()
	lth := len(points)
	ch := make(chan *Point, lth)
	chv := sync.Map{}

	for idx := range points {
		go func(p *Point) {
			nb := b.Copy()
			nb.GoChess(p, C_Player)
			//defer nb.GoBack()

			_, v = nb.aiDfs(dep - 1)

			chv.Store(p.Hash, v)
			ch <- p
		}(points[idx])
	}

	var p *Point
	for i := 0; i < lth; i++ {
		p = <-ch
		vi, _ := chv.Load(p.Hash)
		v := vi.(int)

		if v < found {
			rp = p
			found = v
		}
	}

	return
}

//	当前棋子在每个位置落子后当前局势分
func (b *Board) CalcScoreMaps(c ChessType) [bSize][bSize]int {
	maps := [bSize][bSize]int{}

	b.Each(func(x, y int) {
		maps[x][y] = 1 << 30;
	})

	ps := b.genAvailablePoints()
	for _, p := range ps {
		i := p.X
		j := p.Y

		b.GoXY(i, j, c)
		maps[i][j] = b.Evaluation()
		b.GoBack()
	}

	return maps
}
