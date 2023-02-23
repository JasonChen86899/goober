package list

import (
	"math/rand"
	"testing"
	"time"
)

func TestSkipList_Get(t *testing.T) {
	slHead := &skipListNode{
		forward:  make([]*skipListNode, 10),
		backward: nil,
		member:   nil,
		score:    0,
	}
	sl := &SkipList{
		head:    slHead,
		tail:    nil,
		length:  0,
		level:   10,
		randSed: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	t.Log(sl.Get(1))
}

func TestSkipList_Put(t *testing.T) {
	slHead := &skipListNode{
		forward:  make([]*skipListNode, 10),
		backward: nil,
		member:   nil,
		score:    0,
	}
	sl := &SkipList{
		head:    slHead,
		tail:    nil,
		length:  0,
		level:   10,
		randSed: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	sl.Put("a", 1)
	t.Log(sl.Get(1))
	sl.Put("b", 2)
	sl.Put("c", 1.5)
	t.Log(sl.Get(1.5))
	sl.Put("d", 1.5)
}
