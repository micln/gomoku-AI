package gomoku_AI_test

import (
	"testing"

	"docker/pkg/testutil/assert"

	"github.com/micln/gomoku-AI"
)

func TestBoard_HasChessInXY(t *testing.T) {
	b := gomoku_AI.NewBoard()
	b.GoChess(gomoku_AI.NewPoint(6, 7), gomoku_AI.C_Robot)
	assert.Equal(t, b.HasChessInXY(6, 7), true)
}

func TestBoard_HasXYIs(t *testing.T) {
	b := gomoku_AI.NewBoard()
	b.GoChess(gomoku_AI.NewPoint(6, 7), gomoku_AI.C_Robot)
	assert.Equal(t, b.HasXYIs(6, 7, gomoku_AI.C_Robot), true)
}

func TestBoard_Evaluation(t *testing.T) {
	testGoSteps(t, [][]interface{}{
		{7, 7, gomoku_AI.C_Robot, 10},
		{6, 7, gomoku_AI.C_Player, 0},
		{6, 8, gomoku_AI.C_Robot, 90},
		{5, 7, gomoku_AI.C_Player, 0},
		{5, 9, gomoku_AI.C_Robot, 900},
		{4, 7, gomoku_AI.C_Player, 0},
		{4, 10, gomoku_AI.C_Robot, 9000},
	})

	testGoSteps(t, [][]interface{}{
		{7, 7, gomoku_AI.C_Robot, 10},
		{7, 6, gomoku_AI.C_Player, 0},
		{8, 7, gomoku_AI.C_Robot, 90},
		{9, 8, gomoku_AI.C_Robot, 90},
	})
}

func testGoSteps(t *testing.T, steps [][]interface{}) {
	b := gomoku_AI.NewBoard()

	for si := range steps {
		step := steps[si]
		b.GoChess(
			gomoku_AI.NewPoint(step[0].(int), step[1].(int)),
			step[2].(gomoku_AI.ChessType),
		)
		assert.Equal(t, step[3].(int), b.Evaluation())
	}
}
