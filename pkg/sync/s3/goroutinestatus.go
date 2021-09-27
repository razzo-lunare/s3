package s3

import (
	"time"
)

// GoRoutineStatus is a helper to track when a goroutine that
// recursively adds more times to the queue finishes. This is done
// by keeping track of a channel and the state for goroutines
// to check if they are finished. Each goroutine must update
// their state when it starts and finish a task.
type GoRoutineStatus struct {
	channel chan string
	state   []int
}

// NewGoRoutineStatus returns a new GoRoutineStatus
func NewGoRoutineStatus(numberOfGoRoutines int, s3Prefixes chan string) *GoRoutineStatus {
	return &GoRoutineStatus{
		channel: s3Prefixes,
		state:   make([]int, numberOfGoRoutines),
	}
}

// Wait for the current go routines to finish
// TODO: This could be cpu intensive.
func (g *GoRoutineStatus) Wait() {
	// Verify no jobs are still running
	for {
		if g.isAllDone() {
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

// StartTask is used inside a goroutine to signify it started work
func (g *GoRoutineStatus) StartTask(goRoutineID int) {
	g.state[goRoutineID] = 1
}

// FinishTask is used inside a goroutine to signify it finished work
func (g *GoRoutineStatus) FinishTask(goRoutineID int) {
	g.state[goRoutineID] = 0
}

func (g *GoRoutineStatus) isAllDone() bool {
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
