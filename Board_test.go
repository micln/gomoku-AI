package gomoku_AI_test

import (
	"testing"

	"docker/pkg/testutil/assert"

	ai "github.com/micln/gomoku-AI"
)

func TestBoard_HasChessInXY(t *testing.T) {
	b := ai.NewBoard()
	b.GoChess(ai.NewPoint(6, 7), ai.C_Robot)
	assert.Equal(t, b.HasXY(6, 7), true)
}

func TestBoard_HasXYIs(t *testing.T) {
	b := ai.NewBoard()
	b.GoChess(ai.NewPoint(6, 7), ai.C_Robot)
	assert.Equal(t, b.HasXYIs(6, 7, ai.C_Robot), true)
}

func TestBoard_Evaluation(t *testing.T) {

	//	固定棋局
	//b := ai.NewBoard()
	//b.FillMap(`111`)
	//assert.Equal(t, b.Evaluation(), 1000)

	//	走几步
	testGoSteps(t, [][]interface{}{
		{7, 7, ai.C_Robot, 10},
		{6, 7, ai.C_Player, 0},
		{6, 8, ai.C_Robot, 90},
		{5, 7, ai.C_Player, 0},
		{5, 9, ai.C_Robot, 900},
		{4, 7, ai.C_Player, 0},
		{4, 10, ai.C_Robot, 9000},
	})

	testGoSteps(t, [][]interface{}{
		{7, 7, ai.C_Robot, 10},
		{7, 6, ai.C_Player, 0},
		{8, 7, ai.C_Robot, 90},
		{9, 8, ai.C_Robot, 90},
	})
}

func testGoSteps(t *testing.T, steps [][]interface{}) {
	b := ai.NewBoard()

	for si := range steps {
		step := steps[si]
		b.GoChess(
			ai.NewPoint(step[0].(int), step[1].(int)),
			step[2].(ai.ChessType),
		)
		assert.Equal(t, step[3].(int), b.Evaluation())
	}
}
