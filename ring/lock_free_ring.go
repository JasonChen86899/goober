package ring

import (
	"runtime"
	"sync/atomic"
)

type LockFreeRing struct {
	// a array for store data
	items []interface{}
	size  uint64

	// first item index
	head uint64
	// last item next index
	tail uint64
}

func NewLockFreeRing(size uint64) *LockFreeRing {
	return &LockFreeRing{
		items: make([]interface{}, size, size),
		size:  size,
		head:  0,
		tail:  0,
	}
}

func (lFRing *LockFreeRing) Put(item interface{}) error {
	for {
		if !lFRing.checkListFull() {
			tail := lFRing.tail
			newTail := (tail + 1) % lFRing.size
			if atomic.CompareAndSwapUint64(&lFRing.tail, tail, newTail) {
				lFRing.items[tail] = item
				return nil
			} else {
				runtime.Gosched()
				continue
			}
		}

		return ErrListFull
	}
}

func (lFRing *LockFreeRing) Get() (interface{}, error) {
	for {
		if !lFRing.checkListEmpty() {
			head := lFRing.head
			newHead := (head + 1) % lFRing.size
			if atomic.CompareAndSwapUint64(&lFRing.head, head, newHead) {
				return lFRing.items[head], nil
			} else {
				runtime.Gosched()
				continue
			}
		}

		return nil, ErrListEmpty
	}
}

func (lFRing *LockFreeRing) checkListFull() bool {
	return lFRing.head == (lFRing.tail+1)%lFRing.size
}

func (lFRing *LockFreeRing) checkListEmpty() bool {
	return lFRing.head == lFRing.tail
}
