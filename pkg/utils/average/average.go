package average

import (
	"sync"
	"time"
)

// JobAverageInt average execution time of a thread
type JobAverageInt struct {
	Values []float64
	Count  int

	lock sync.Mutex
	time time.Time
}

// New return a new AverageInt obj
func New() *JobAverageInt {
	return &JobAverageInt{
		Values: make([]float64, 0),
		Count:  0,
		lock:   sync.Mutex{},
	}
}

// StartTimer starts the timer for a background proccess
func (a *JobAverageInt) StartTimer() {
	a.time = time.Now()
}

// EndTimer ends the timer for a background proccess
func (a *JobAverageInt) EndTimer() {
	duration := time.Since(a.time)
	a.add(duration.Seconds())
}

// GetAverage  reurn the average exection time
func (a *JobAverageInt) GetAverage() float64 {
	total := float64(0)
	for _, value := range a.Values {
		total += value
	}

	// Block divide by zero errors
	if total == 0 || a.Count == 0 {
		return 0
	}

	return float64(total) / float64(a.Count)
}

func (a *JobAverageInt) add(newInt float64) {
	a.lock.Lock()
	a.Values = append(a.Values, newInt)
	a.Count++
	a.lock.Unlock()
}
