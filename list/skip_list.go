package list

import (
	"math/rand"
	"time"
)

// SkipList skip list is a sorted list.
// each list node has forwards pointer which can skip into another node
type SkipList struct {
	head *skipListNode
	tail *skipListNode

	length int
	level  int

	randSed *rand.Rand
}

type skipListNode struct {
	forward  []*skipListNode
	backward *skipListNode

	member interface{}
	score  float64
}

func NewSkipList(maxLevel int) *SkipList {
	slHead := &skipListNode{
		forward:  make([]*skipListNode, maxLevel),
		backward: nil,
		member:   nil,
		score:    0,
	}
	sl := &SkipList{
		head:    slHead,
		tail:    nil,
		length:  0,
		level:   maxLevel,
		randSed: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	return sl
}

// Put ...
func (s *SkipList) Put(member interface{}, score float64) {
	var (
		n      = s.head
		l      = s.level - 1
		update = make([]*skipListNode, s.level)
	)

	if n == nil && s.length == 0 {
		nn := s.newSkipNode(member, score)
		for i := 0; i < len(nn.forward); i++ {
			// record need update level and node
			// new node need change self forwards and forwards of nodes which behind it
			s.head.forward[i] = nn
		}
		s.length++
		return
	}

	// node.score <= score then forward
	// n1 -----> n2
	// node.score > score then down
	// n1 level n -----> n1 level n-1
	for l >= 0 && (n == s.head || n.score <= score) {
		for l >= 0 && (n.forward[l] == nil || n.forward[l].score > score) {
			// record node which need update!!!
			update[l] = n
			l--
		}
		if l >= 0 {
			n = n.forward[l]
		}
	}

	// build new skip node
	nn := s.newSkipNode(member, score)
	for i := 0; i < len(nn.forward); i++ {
		// for-range change self forwards and forwards of update nodes
		nn.forward[i] = update[i].forward[i]
		update[i].forward[i] = nn
	}
	s.length++
}

func (s *SkipList) randomLevel() int {
	return s.randSed.Intn(s.level)
}

func (s *SkipList) newSkipNode(member interface{}, score float64) *skipListNode {
	level := s.randomLevel()
	node := new(skipListNode)
	node.forward = make([]*skipListNode, level)
	node.score = score
	node.member = member
	return node
}

// Get ...
func (s *SkipList) Get(score float64) (interface{}, error) {
	var (
		l = s.level - 1
		n = s.head
	)

	// node.score <= score then forward
	// n1 -----> n2
	// node.score > score then down
	// n1 level n -----> n1 level n-1
	for l >= 0 && (n == s.head || n.score <= score) {
		for l >= 0 && (n.forward[l] == nil || n.forward[l].score > score) {
			l--
		}
		if l >= 0 {
			n = n.forward[l]
		}
	}
	if n.score != score {
		return nil, ErrNotFound
	}
	return n.member, nil
}
