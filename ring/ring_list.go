package ring

import "errors"

var (
	ErrListFull  = errors.New("list full")
	ErrListEmpty = errors.New("list empty")
)

/*
Sacrifice one element space, that is to say the actual length is the underlying array length-1

ring empty: head=tail
ring full:  head=(tail+1)%size
*/
type RingList struct {
	// a array for store data
	items []interface{}
	size  uint64

	// first item index
	head uint64
	// last item next index
	tail uint64
}

func NewRingList(size uint64) *RingList {
	return &RingList{
		items: make([]interface{}, size, size),
		size:  size,
		head:  0,
		tail:  0,
	}
}

func (rl *RingList) checkListFull() bool {
	return rl.head == (rl.tail+1)%rl.size
}

func (rl *RingList) checkListEmpty() bool {
	return rl.head == rl.tail
}

func (rl *RingList) Put(item interface{}) error {
	if !rl.checkListFull() {
		rl.items[rl.tail] = item
		rl.tail = (rl.tail + 1) % rl.size
		return nil
	}

	return ErrListFull
}

func (rl *RingList) Get() (interface{}, error) {
	if !rl.checkListEmpty() {
		defer func() {
			rl.head = (rl.head + 1) % rl.size
		}()
		return rl.items[rl.head], nil
	}

	return nil, ErrListEmpty
}

func (rl *RingList) Length() uint64 {
	v := rl.tail - rl.head
	if v > 0 {
		return v - 1
	} else {
		return v - 1 + rl.size
	}
}
