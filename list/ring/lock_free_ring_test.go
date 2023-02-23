package ring

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestRunLockFreeQueue(t *testing.T) {
	putGroup := sync.WaitGroup{}
	lfQueue := NewLockFreeQueue(8)
	for i := 0; i < 8; i++ {
		putGroup.Add(1)
		j := i
		go func() {
			if err := lfQueue.Put(j); err != nil {
				fmt.Println(err)
			}
			putGroup.Done()
		}()
	}
	putGroup.Wait()

	getGroup := sync.WaitGroup{}
	for i := 0; i < 7; i++ {
		//groupAdd.Add(1)
		//go func() {
		fmt.Println(lfQueue.Get())
		//groupAdd.Done()
		//}()
	}

	getGroup.Wait()
}

var testBytes = []byte("e7287wdfhjsbcjsye8wdbnjye3849dkscnjsyr93")

func TestNativeChannel(t *testing.T) {
	size := uint64(1024 * 1000)
	testChan := make(chan interface{}, 1024*30)
	chanGroup := sync.WaitGroup{}
	chanGroup.Add(int(size))

	go func() {
		for {
			select {
			case <-testChan:
			}
		}
	}()

	startTime := time.Now()
	for i := 0; i < int(size); i++ {
		//j := i
		go func() {
			select {
			case testChan <- testBytes:
				chanGroup.Done()
			}
		}()
	}
	chanGroup.Wait()
	fmt.Println("Channel: ", time.Since(startTime))
}

func TestLockFreeQueue(t *testing.T) {
	size := uint64(1024 * 1000)
	testLFQueue := NewLockFreeQueue(1024 * 30)
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
	fmt.Println("LockFreeQueue: ", time.Since(startTime))
}

func TestLockFreeQueueComparet(t *testing.T) {
	TestLockFreeQueue(t)
	TestNativeChannel(t)
}
