package s3

import (
	"time"
)

type GoRoutineStatus struct {
	channel chan string
	state   []int
}

func NewGoRoutineStatus(numberOfGoRoutines int, s3Prefixes chan string) *GoRoutineStatus {
	return &GoRoutineStatus{
		channel: s3Prefixes,
		state:   make([]int, numberOfGoRoutines),
	}
}

// TODO: This could be cpu intensive.
func (g *GoRoutineStatus) Wait() {
	// Verify no jobs are still running
	for {
		if g.IsAllDone() {
			return
		}
		// Vary the sleep time depending how how close we are to finishing
		if len(g.channel) > 50 {
			time.Sleep(500 * time.Millisecond)
		} else {
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (g *GoRoutineStatus) SetStateRunning(goRoutineID int) {
	g.state[goRoutineID] = 1
}

func (g *GoRoutineStatus) SetStateDone(goRoutineID int) {
	g.state[goRoutineID] = 0
}

func (g *GoRoutineStatus) IsAllDone() bool {
	// If there is an item in the channel we are
	if len(g.channel) != 0 {
		return false
	}

	for _, state := range g.state {
		if state == 1 {
			return false
		}
	}

	return true
}
