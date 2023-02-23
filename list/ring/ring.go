package ring

import "goober/list"

// Ring
// Sacrifice one element space, that means the actual length is the underlying array length-1
//
// It is concurrency unsafe
// ring empty: head=tail
// ring full:  head=(tail+1)%size
type Ring struct {
	// a array for store data
	items []interface{}
	size  uint64

	// first item index
	head uint64
	// last item next index
	tail uint64
}

// NewRing new a ring
func NewRing(size uint64) *Ring {
	return &Ring{
		items: make([]interface{}, size+1, size+1),
		size:  size + 1,
		head:  0,
		tail:  0,
	}
}

func (rl *Ring) checkRingFull() bool {
	return rl.head == (rl.tail+1)%rl.size
}

func (rl *Ring) checkRingEmpty() bool {
	return rl.head == rl.tail
}

// Put ...
func (rl *Ring) Put(item interface{}) error {
	if !rl.checkRingFull() {
		rl.items[rl.tail] = item
		rl.tail = (rl.tail + 1) % rl.size
		return nil
	}

	return list.ErrFull
}

// Get ...
func (rl *Ring) Get() (interface{}, error) {
	if !rl.checkRingEmpty() {
		defer func() {
			rl.head = (rl.head + 1) % rl.size
		}()
		return rl.items[rl.head], nil
	}

	return nil, list.ErrEmpty
}

// Length ...
func (rl *Ring) Length() uint64 {
	v := rl.tail - rl.head
	if v > 0 {
		return v - 1
	} else {
		return v - 1 + rl.size
	}
}
