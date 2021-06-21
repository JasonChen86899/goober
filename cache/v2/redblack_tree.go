package v2

import (
	rbt "github.com/emirpasic/gods/trees/redblacktree"
	"sync"
)

type redBlackTree struct {
	sync.RWMutex
	tree *rbt.Tree
}

func newRedBlackTree() *redBlackTree {
	return &redBlackTree{tree: rbt.NewWithStringComparator()}
}

func (t *redBlackTree) get(key interface{}) (value interface{}, found bool) {
	t.RLock()
	defer t.RUnlock()

	return t.tree.Get(key)
}

func (t *redBlackTree) put(key, value interface{}) {
	t.Lock()
	defer t.Unlock()

	t.tree.Put(key, value)
}

func (t *redBlackTree) delete(key interface{}) {
	t.Lock()
	defer t.Unlock()

	t.tree.Remove(key)
}

func (t *redBlackTree) rangeTree(f func(key, value interface{})) {
	t.RLock()
	defer t.RUnlock()

	iterator := t.tree.Iterator()
	for iterator.Next() {
		key := iterator.Key()
		value := iterator.Value()
		f(key, value)
	}
}
