package cache

import (
	"math/rand"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"goober/cache/v2"
)

var (
	benchMarkIndex int64
)

func BenchmarkRedBlackMap_PutOrGet(B *testing.B) {
	m := v2.NewRedBlackMap()

	rand.Seed(time.Now().UnixNano())
	B.ResetTimer()
	B.RunParallel(func(pb *testing.PB) {
		id := int(atomic.AddInt64(&benchMarkIndex, 1) - 1)
		i := id * B.N
		for ; pb.Next(); i++ {
			j := strconv.Itoa(i)

			if rand.Intn(2) == 1 {
				m.Put(j, j)
				continue
			}

			_, _ = m.Get(j)
		}
	})
}

func BenchmarkLRUMap_PutOrGet(B *testing.B) {
	m := NewLRUCache()

	rand.Seed(time.Now().UnixNano())
	B.ResetTimer()
	B.RunParallel(func(pb *testing.PB) {
		id := int(atomic.AddInt64(&benchMarkIndex, 1) - 1)
		i := id * B.N
		for ; pb.Next(); i++ {
			j := strconv.Itoa(i)

			if rand.Intn(2) == 1 {
				m.Put(j, j)
				continue
			}

			_, _ = m.Get(j)
		}
	})
}

func BenchmarkSynMap_PutOrGet(B *testing.B) {
	m := sync.Map{}

	rand.Seed(time.Now().UnixNano())
	B.ResetTimer()
	B.RunParallel(func(pb *testing.PB) {
		id := int(atomic.AddInt64(&benchMarkIndex, 1) - 1)
		i := id * B.N
		for ; pb.Next(); i++ {
			j := strconv.Itoa(i)

			if rand.Intn(2) == 1 {
				m.Store(j, j)
				continue
			}

			_, _ = m.Load(j)
		}
	})
}
