package v2

import (
	"sync/atomic"
	"unsafe"
)

// bucket node
type node struct {
	key   interface{}
	value interface{}
	next  *node
}

// return a new node
func newNode(key, value interface{}) *node {
	return &node{
		key:   key,
		value: value,
		next:  nil,
	}
}

// lock-free list
type bucket struct {
	head unsafe.Pointer
	tail unsafe.Pointer
}

// return a new bucket
func newBucket() *bucket {
	return &bucket{
		head: nil,
		tail: nil,
	}
}

// find value of key
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

		tmp = tmp.next
	}

	return nil, false
}

// Get return value of key.
// if not found return false
func (b *bucket) Get(key interface{}) (interface{}, bool) {
	return b.find(key)
}

//// Put set key and value to bucket
//func (b *bucket) Put(key, value interface{}) {
//	headAddr := &b.head
//	tailAddr := &b.tail
//
//	tmpPre := atomic.LoadPointer(tailAddr)
//	tmpNewTail := &node{
//		key:   key,
//		value: value,
//		next:  nil,
//	}
//
//	// cas
//	for {
//		tmpPre = atomic.LoadPointer(tailAddr)
//		if tmpPre == nil {
//			if atomic.CompareAndSwapPointer(headAddr, nil, unsafe.Pointer(tmpNewTail)) {
//				atomic.StorePointer(tailAddr, unsafe.Pointer(tmpNewTail))
//				return
//			}
//			continue
//		}
//
//		preNode := (*node)(tmpPre)
//		if !atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&preNode.next)),
//			nil, unsafe.Pointer(tmpNewTail)) {
//			continue
//		}
//
//		atomic.StorePointer(tailAddr, unsafe.Pointer(tmpNewTail))
//		return
//	}
//}

// Put set key and value to bucket if key exist then recover
func (b *bucket) Put(key, value interface{}) {
	headAddr := &b.head
	tmpPre := atomic.LoadPointer(headAddr)
	tmpNewTail := &node{
		key:   key,
		value: value,
		next:  nil,
	}

	for {
		// 1. if head nil then cas
		tmpPre = atomic.LoadPointer(headAddr)
		if tmpPre == nil {
			if atomic.CompareAndSwapPointer(headAddr, nil, unsafe.Pointer(tmpNewTail)) {
				return
			}
			continue
		}

		curNode := (*node)(tmpPre)
		preNode := curNode
		for {
			// 2. find key
			for curNode != nil {
				if curNode.key == key {
					curNode.value = value // recover
					return
				}
				preNode = curNode
				curNode = curNode.next
			}
			// 3. CAS preNode.next
			if !atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&preNode.next)),
				nil, unsafe.Pointer(tmpNewTail)) {
				// 4. if CAS fail then reset curNode
				curNode = preNode.next
				continue
			}
			return
		}
	}
}

func (b *bucket) rangeBucket(f func(key, value interface{})) {
	n := (*node)(atomic.LoadPointer(&b.head))
	for n != nil {
		f(n.key, n.value)
		n = n.next
	}
}
