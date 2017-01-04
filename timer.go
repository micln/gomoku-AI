package gomoku_AI

import (
	"time"

	"log"
)

var start = make(map[string]time.Time)

func TimeStart(id string) {
	start[id] = time.Now()
}

func TimerEnd(id string) {
	diff := time.Now().Sub(start[id])

	log.Printf("Time[%s] %.3fms\n",id, float64(diff.Nanoseconds()) /1000000)
}
