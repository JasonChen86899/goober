package v2

import (
	"go.uber.org/atomic"
)

const (
	reHashThreshold = 0.75
	initBPower      = 3
)

type Map struct {
	count *atomic.Uint64

	buckets    []*bucket
	oldBuckets []*bucket

	bPower uint64

	growingIndex *atomic.Int64
}

func NewMap() *Map {
	m := &Map{
		count: atomic.NewUint64(0),

		buckets: nil,
		bPower:  initBPower,
	}

	bs := 2 ^ m.bPower
	m.buckets = make([]*bucket, bs)
	for i := uint64(0); i < bs; i++ {
		m.buckets[i] = newBucket()
	}

	return m
}

func (m *Map) Get(key string) (interface{}, bool) {
	keyHash := strHash(key)
	i := keyHash % uint64(len(m.buckets))
	b := m.buckets[i]

	return b.Get(key)
}

func (m *Map) Put(key string, value interface{}) {
	keyHash := strHash(key)
	i := keyHash % uint64(len(m.buckets))
	b := m.buckets[i]

	b.Put(key, value)
	m.count.Inc()
}

func (m *Map) doGrowWork() {
	//moveBucket := m.oldBuckets[m.growingIndex.Load()]

}
