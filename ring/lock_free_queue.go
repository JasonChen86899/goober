package ring

import (
	"runtime"
	"sync/atomic"
)

type LockFreeQueue struct {
	RingList
}

func NewLockFreeQueue(size uint64) *LockFreeQueue {
	return &LockFreeQueue{
		RingList{
			size:  size,
			items: make([]interface{}, size, size),
			head:  0,
			tail:  0,
		},
	}
}

func (lfQueue *LockFreeQueue) Put(item interface{}) error {
	tailPointer := &lfQueue.tail
	for {
		// cas tail
		oldValue := lfQueue.tail
		newValue := (oldValue + 1) % lfQueue.size
		if lfQueue.head == newValue {
			return ErrListFull
		}

		if atomic.CompareAndSwapUint64(tailPointer, oldValue, newValue) {
			lfQueue.items[oldValue] = item
			return nil
		}
		runtime.Gosched()
	}
}

func (lfQueue *LockFreeQueue) Get() (interface{}, error) {
	headPointer := &lfQueue.head
	for {
		// cas head
		oldValue := lfQueue.head
		if lfQueue.tail == oldValue {
			return nil, ErrListEmpty
		}

		newValue := (oldValue + 1) % lfQueue.size
		if atomic.CompareAndSwapUint64(headPointer, oldValue, newValue) {
			return lfQueue.items[oldValue], nil
		}
		runtime.Gosched()
	}
}
