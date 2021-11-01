package s3

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func signalS3CacheLoaded(s3CacheLoadedSignal chan bool) {
	time.Sleep(1 * time.Second)
	close(s3CacheLoadedSignal)
}

func waitForSignal(wg *sync.WaitGroup, s3CacheLoadedSignal chan bool) bool {
	for i := 0; i < 10; i++ {
		fmt.Println("waiting for s3CacheLoadedSignal")
		select {
		case <-s3CacheLoadedSignal:
			fmt.Println("s3CacheLoadedSignal received")
			wg.Done()
			return true
		default:
		}

		time.Sleep(500 * time.Millisecond)
	}
	wg.Done()
	return false
}

func TestSignal(t *testing.T) {
	s3CacheLoadedSignal := make(chan bool)
	wg := sync.WaitGroup{}
	go signalS3CacheLoaded(s3CacheLoadedSignal)
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go waitForSignal(&wg, s3CacheLoadedSignal)
	}

	wg.Wait()
}
