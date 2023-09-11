package v2

import (
	"go.uber.org/atomic"
	"math"
	"sync"
)

type bucketPackage struct {
	buckets    []*bucket
	oldBuckets []*bucket
}

type Map struct {
	cap   *atomic.Uint64
	count *atomic.Uint64

	//buckets    []*bucket
	//oldBuckets []*bucket
	bkPkg *atomic.Value

	growing      *atomic.Bool
	growingIndex *atomic.Int64

	sync.RWMutex
}

func NewLockFreeMap() *Map {
	m := &Map{
		cap:   atomic.NewUint64(uint64(math.Pow(2, initBPower))),
		count: atomic.NewUint64(0),

		//buckets: nil,
		bkPkg: &atomic.Value{},

		growing:      atomic.NewBool(false),
		growingIndex: atomic.NewInt64(-1),
	}

	bs := m.cap.Load()
	buckets := make([]*bucket, bs)
	for i := uint64(0); i < bs; i++ {
		buckets[i] = newBucket()
	}
	m.bkPkg.Store(&bucketPackage{buckets: buckets})

	return m
}

func (m *Map) Get(key string) (interface{}, bool) {
	m.RLock()
	defer m.RUnlock()

	bkPkg := m.bkPkg.Load().(*bucketPackage)
	oldBs := bkPkg.oldBuckets
	newBs := bkPkg.buckets

	v, ok := m.getFormBuckets(key, newBs)
	if ok {
		return v, ok
	}
	return m.getFormBuckets(key, oldBs)
}

func (m *Map) getFormBuckets(key string, buckets []*bucket) (interface{}, bool) {
	if len(buckets) == 0 {
		return nil, false
	}
	keyHash := strHash(key)
	i := keyHash % uint64(len(buckets))
	b := buckets[i]

	return b.Get(key)
}

func (m *Map) Put(key string, value interface{}) {
	m.put(key, value)
	m.count.Inc()
	if m.checkReHashThreshold() {
		go m.doGrowWork()
	}
}

func (m *Map) put(key string, value interface{}) {
	// need a lock
	// Make sure you get new buckets or will write to old bucket!!!
	// old bucket must be not written because of copying
	m.RLock()
	defer m.RUnlock()

	bkPkg := m.bkPkg.Load().(*bucketPackage)
	newBs := bkPkg.buckets

	keyHash := strHash(key)
	i := keyHash % uint64(len(newBs))
	b := newBs[i]

	b.Put(key, value)
}

func (m *Map) checkReHashThreshold() bool {
	if m.growingIndex.Load() >= 0 {
		return true
	}

	m.Lock()
	defer m.Unlock()

	if m.growingIndex.Load() >= 0 {
		return true
	}

	cnt := m.count.Load()
	cp := m.cap.Load()
	ratio := float64(cnt) / float64(cp)
	if ratio >= reHashThreshold {
		cp = cp << 1
		//fmt.Println("cap:", cp, "|", cp>>1, cnt, ratio)
		m.cap.Store(cp)
		buckets := make([]*bucket, cp)
		for i := range buckets {
			buckets[i] = newBucket()
		}

		//m.Lock()
		oldBkPkg := m.bkPkg.Load().(*bucketPackage)
		m.bkPkg.Store(&bucketPackage{
			buckets:    buckets,
			oldBuckets: oldBkPkg.buckets,
		})
		//m.Unlock()

		m.growingIndex.Store(0)
		return true
	}
	return false
}

func (m *Map) doGrowWork() {
	if !m.growing.CAS(false, true) {
		return
	}
	defer m.growing.Store(false)

	reHashIndex := m.growingIndex.Load()
	if reHashIndex < 0 {
		return
	}

	// this lock protect Put and following reHashing
	// make sure use new bucket value first
	m.Lock()
	defer m.Unlock()
	reHashIndex = m.growingIndex.Load()
	if reHashIndex < 0 {
		return
	}

	bkPkg := m.bkPkg.Load().(*bucketPackage)
	oldBs := bkPkg.oldBuckets
	newBs := bkPkg.buckets

	// rehash end
	if reHashIndex == int64(len(oldBs)) {
		// m.oldBuckets = nil
		bkPkg.oldBuckets = nil
		// fmt.Println("reHashIndex: +++++++++++", reHashIndex)
		m.growingIndex.Store(-1)
		return
	}

	//fmt.Println("reHashIndex: -----------", reHashIndex)
	oldBs[reHashIndex].rangeBucket(func(key, value interface{}) {
		keyHash := strHash(key.(string))
		i := keyHash % uint64(len(newBs))
		b := newBs[i]
		if _, ok := b.Get(key); !ok {
			b.Put(key, value)
		}
	})
	m.growingIndex.Inc()
}
