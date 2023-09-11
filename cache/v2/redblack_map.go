package v2

import (
	"math"
	"sync"

	"go.uber.org/atomic"
)

type redBlackBucket struct {
	b         []*redBlackTree
	reHashing bool
}

type RedBlackMap struct {
	mu sync.RWMutex

	cap   uint64
	count *atomic.Uint64

	buckets    *atomic.Value
	oldBuckets *atomic.Value

	reHashIndex int64

	growing *atomic.Bool
}

func NewRedBlackMap() *RedBlackMap {
	m := &RedBlackMap{
		cap:   uint64(math.Pow(2, initBPower)),
		count: atomic.NewUint64(0),

		oldBuckets: &atomic.Value{},
		buckets:    &atomic.Value{},

		growing: atomic.NewBool(false),
	}

	bs := m.cap
	buckets := make([]*redBlackTree, bs)
	for i := uint64(0); i < bs; i++ {
		buckets[i] = newRedBlackTree()
	}
	m.buckets.Store(redBlackBucket{
		b:         buckets,
		reHashing: false,
	})

	return m
}

func (m *RedBlackMap) Get(key string) (value interface{}, ok bool) {
	keyHash := strHash(key)
	rbBucket := m.buckets.Load().(redBlackBucket)
	i := keyHash % uint64(len(rbBucket.b))
	b := rbBucket.b[i]
	value, ok = b.get(key)
	if ok {
		return
	}

	// no rehashing
	if !rbBucket.reHashing {
		return value, ok
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	value, ok = m.getFromBucket(key, keyHash)
	if ok {
		return
	}

	return m.getFromOldBucket(key, keyHash)
}

func (m *RedBlackMap) getFromBucket(key string, keyHash uint64) (interface{}, bool) {
	rbBucket := m.buckets.Load().(redBlackBucket)
	i := keyHash % uint64(len(rbBucket.b))
	b := rbBucket.b[i]

	return b.get(key)
}

func (m *RedBlackMap) getFromOldBucket(key string, keyHash uint64) (interface{}, bool) {
	rbBucket := m.oldBuckets.Load().(redBlackBucket)
	i := keyHash % uint64(len(rbBucket.b))
	b := rbBucket.b[i]

	return b.get(key)
}

func (m *RedBlackMap) Put(key string, value interface{}) {
	m.mu.RLock()
	m.put(key, value)
	m.mu.RUnlock()

	m.count.Add(1)

	// check if map need rehash
	if m.checkReHashThreshold() {
		go m.reHashing()
	}
}

func (m *RedBlackMap) put(key string, value interface{}) {
	keyHash := strHash(key)
	curBucket := m.buckets.Load().(redBlackBucket)
	i := keyHash % uint64(len(curBucket.b))
	b := curBucket.b[i]
	b.put(key, value)
}

func (m *RedBlackMap) checkReHashThreshold() bool {
	rbBucket := m.buckets.Load().(redBlackBucket)
	if rbBucket.reHashing {
		return true
	}
	m.mu.RLock()
	ratio := float64(m.count.Load()) / float64(m.cap)
	m.mu.RUnlock()

	if ratio >= reHashThreshold {
		m.mu.Lock()
		defer m.mu.Unlock()

		rbBucket := m.buckets.Load().(redBlackBucket)
		if rbBucket.reHashing {
			return true
		}

		m.cap = m.cap << 1 // cap * 2
		bs := m.cap
		buckets := make([]*redBlackTree, bs)
		for i := uint64(0); i < bs; i++ {
			buckets[i] = newRedBlackTree()
		}

		m.oldBuckets.Store(rbBucket)
		m.buckets.Store(redBlackBucket{
			b:         buckets,
			reHashing: true,
		})

		return true
	}

	return false
}

func (m *RedBlackMap) reHashing() {
	if !m.growing.CAS(false, true) {
		return
	}
	defer m.growing.Store(false)

	m.mu.Lock()
	defer m.mu.Unlock()

	reHashIndex := m.reHashIndex
	curBucket := m.buckets.Load().(redBlackBucket)
	oldBucket := m.oldBuckets.Load().(redBlackBucket)

	if !curBucket.reHashing {
		return
	}

	oldBucket.b[reHashIndex].rangeTree(func(key, value interface{}) {
		keyHash := strHash(key.(string))
		i := keyHash % uint64(len(curBucket.b))
		b := curBucket.b[i]
		if _, ok := b.get(key); !ok {
			m.put(key.(string), value)
		}
	})

	// rehash end
	if reHashIndex == int64(len(oldBucket.b)-1) {
		oldBucket.b = nil
		curBucket.reHashing = false
		m.reHashIndex = 0
	} else {
		m.reHashIndex++
	}
}
