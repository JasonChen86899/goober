package ring

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestRunChannel(t *testing.T) {
	condChan := NewChannel(1)

	go func() {
		err := condChan.Put("1")
		fmt.Println("Put 1: ", err)
		err = condChan.Put("2")
		fmt.Println("Put 2: ", err)
		err = condChan.Put("3")
		fmt.Println("Put 3: ", err)
	}()

	go func() {
		fmt.Println(condChan.Get())
		fmt.Println(condChan.Get())
		fmt.Println(condChan.Get())
	}()

	time.Sleep(time.Hour)
}

func TestChannel(t *testing.T) {
	condChan := NewChannel(100)

	a := 100000
	w := sync.WaitGroup{}
	w.Add(2)

	s1 := time.Now()
	go func() {
		for i := 0; i < a; i++ {
			condChan.Put(1)
		}

		w.Done()
	}()

	go func() {
		for i := 0; i < a; i++ {
			condChan.Get()
		}

		w.Done()
	}()

	w.Wait()
	fmt.Println("s1", time.Since(s1))
}

func TestGoChannel(t *testing.T) {
	a := 100000
	w := sync.WaitGroup{}
	w.Add(2)

	ch := make(chan interface{}, 100)
	s2 := time.Now()
	go func() {
		for i := 0; i < a; i++ {
			ch <- 1
		}

		w.Done()
	}()

	go func() {
		for i := 0; i < a; i++ {
			_ = <-ch
		}

		w.Done()
	}()

	w.Wait()
	fmt.Println("s2", time.Since(s2))
}

func TestCompare(t *testing.T) {
	TestChannel(t)
	TestGoChannel(t)
}

func TestChannel2(t *testing.T) {
	condChan := NewChannel(100)

	a := 100000
	w := sync.WaitGroup{}
	w.Add(a * 2)

	s1 := time.Now()
	go func() {
		for i := 0; i < a; i++ {
			go func() {
				condChan.Put(1)
				w.Done()
			}()
		}
	}()

	go func() {
		for i := 0; i < a; i++ {
			go func() {
				condChan.Get()
				w.Done()
			}()
		}
	}()

	w.Wait()
	fmt.Println("s1", time.Since(s1))
}

func TestGoChannel2(t *testing.T) {
	a := 100000
	w := sync.WaitGroup{}
	w.Add(a * 2)

	ch := make(chan interface{}, 100)
	s2 := time.Now()
	go func() {
		for i := 0; i < a; i++ {
			go func() {
				ch <- 1
				w.Done()
			}()
		}
	}()

	go func() {
		for i := 0; i < a; i++ {
			go func() {
				_ = <-ch
				w.Done()
			}()
		}
	}()

	w.Wait()
	fmt.Println("s2", time.Since(s2))
}

func TestCompare2(t *testing.T) {
	TestChannel2(t)
	TestGoChannel2(t)
}
