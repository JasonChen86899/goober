package ring

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestLockFreeRing_Put(t *testing.T) {
	lFRing := NewLockFreeRing(1024)
	for i := 0; i < 1000; i++ {
		j := i
		go lFRing.Put(j)
	}

	w := sync.WaitGroup{}
	w.Add(1000)
	m := sync.Map{}
	for i := 0; i < 1000; i++ {
		go func() {
			key, _ := lFRing.Get()
			if key == nil {
				w.Done()
				return
			}
			if _, ok := m.Load(key); !ok {
				m.Store(key, struct{}{})
			} else {
				t.Error(key)
			}
			w.Done()
		}()
	}

	w.Wait()
}

func TestNewLockFreeRing(t *testing.T) {
	size := uint64(1024 * 1000)
	testLFQueue := NewLockFreeRing(1024 * 30)
	lfGroup := sync.WaitGroup{}
	lfGroup.Add(int(size))

	go func() {
		for {
			_, _ = testLFQueue.Get()
		}
	}()

	startTime := time.Now()
	for i := 0; i < int(size); i++ {
		//j := i
		go func() {
			_ = testLFQueue.Put(testBytes)
			lfGroup.Done()
		}()
	}
	lfGroup.Wait()
	fmt.Println("LockFreeRing: ", time.Since(startTime))
}
