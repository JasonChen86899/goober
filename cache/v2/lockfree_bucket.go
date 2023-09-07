package v2

import (
	"sync/atomic"
	"unsafe"
)

type node struct {
	key   interface{}
	value interface{}
	next  *node
}

func newNode(key, value interface{}) *node {
	return &node{
		key:   key,
		value: value,
		next:  nil,
	}
}

type bucket struct {
	head unsafe.Pointer
	tail unsafe.Pointer
}

func newBucket() *bucket {
	return &bucket{
		head: nil,
		tail: nil,
	}
}

func (b *bucket) find(key interface{}) (interface{}, bool) {
	v := atomic.LoadPointer(&b.head)
	if v == nil {
		return nil, false
	}

	tmp := (*node)(v)
	for tmp != nil {
		if tmp.key == key {
			return tmp.value, true
		}

		if tmp.next == nil {
			tmp = nil
			continue
		}

		tmp = tmp.next
	}

	return nil, false
}

func (b *bucket) Get(key interface{}) (interface{}, bool) {
	return b.find(key)
}

func (b *bucket) Put(key, value interface{}) {
	headAddr := &b.head
	tailAddr := &b.tail

	tmpPre := b.tail
	tmpNewTail := &node{
		key:   key,
		value: value,
		next:  nil,
	}

	// cas b.tail
	for !atomic.CompareAndSwapPointer(tailAddr, tmpPre, unsafe.Pointer(tmpNewTail)) {
		tmpPre = atomic.LoadPointer(tailAddr)
	}

	// has no head
	if tmpPre == nil {
		atomic.StorePointer(headAddr, unsafe.Pointer(tmpNewTail))
		return
	}

	// set pre.next
	(*node)(tmpPre).next = tmpNewTail
	return
}

func (b *bucket) forRange(f func(key, value interface{})) {
	n := (*node)(atomic.LoadPointer(&b.head))
	for n != nil {
		f(n.key, n.value)
		n = n.next
	}
}
