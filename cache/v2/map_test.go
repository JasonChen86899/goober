package v2

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"reflect"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"
)

func TestAtomicValue(t *testing.T) {
	var v atomic.Value

	t.Log(v.Load())
}

func BenchmarkLockFreeBucket_Put(B *testing.B) {
	b := newBucket()

	rand.Seed(time.Now().UnixNano())
	B.ResetTimer()
	B.RunParallel(func(pb *testing.PB) {
		i := atomic.Int64{}
		for pb.Next() {
			j := strconv.Itoa(int(i.Add(1)))
			b.Put(j, j)
			//b.Get(j)
		}
	})
}

func TestBucket_PutAndGet(t *testing.T) {
	m := newBucket()

	wg := sync.WaitGroup{}

	a := 100
	b := 2 * a
	wg.Add(b)

	chans := make([]chan struct{}, a)
	for i := 0; i < a; i++ {
		j := i
		kj := strconv.Itoa(j)
		chans[j] = make(chan struct{})
		go func() {
			m.Put(kj, j)
			chans[j] <- struct{}{}
			wg.Done()
		}()
	}

	for i := 0; i < a; i++ {
		j := i
		kj := strconv.Itoa(j)
		go func() {
			<-chans[j]
			vj, ok := m.Get(kj)
			wg.Done()
			if !ok {
				t.Log(kj)
				return
			}
			// assert.Equal(t, true, ok)
			assert.Equal(t, j, vj.(int))
		}()
	}

	wg.Wait()

	//for benchMarkIndex:=0; benchMarkIndex<100; benchMarkIndex++ {
	//	v, ok := b.Get(benchMarkIndex)
	//	if !ok {
	//		t.Log(false)
	//		//w.Done()
	//		continue
	//	}
	//
	//	t.Log(v.(int) == benchMarkIndex)
	//	//w.Done()
	//}
	//
	////w.Wait()
}

func TestUnsafe(t *testing.T) {
	//n := 10
	//b := make([]byte, n)
	//end := unsafe.Pointer(uintptr(unsafe.Pointer(&b[0])) + uintptr(n))
	//bb := *(*[]byte)(end)
	//t.Log(len(bb))
	//
	//u := uintptr(unsafe.Pointer(&b[0]))
	//p := unsafe.Pointer(u + 9)
	//t.Log(*(*byte)(p))
	//
	//var hdr reflect.StringHeader
	//hdr.Data = uintptr(unsafe.Pointer(p))
	//hdr.Len = 1
	//s := *(*string)(unsafe.Pointer(&hdr)) // p possibly already lost
	//t.Log(s)
	//
	//c := &bytes.Buffer{}
	//var z interface{}
	//z = c
	//var y io.ByteReader
	//y = c
	//t.Log(z == y)

	s := "1"
	ss := (*reflect.StringHeader)(unsafe.Pointer(&s))
	t.Log(*(*byte)(unsafe.Pointer(ss.Data)))

	i := int32(1)
	ii := (*int8)(unsafe.Pointer(&i))
	t.Log(*ii)

	type ct struct {
		i int
	}
	type pt struct {
		ct
	}

	p := pt{}
	pp := (*ct)(unsafe.Pointer(&p))
	t.Log(*pp)

	//l := rate.NewLimiter()
	//re := l.Reserve()
	//re.Cancel()
}

func TestNewRedBlackMap(t *testing.T) {
	m := NewRedBlackMap()
	wg := sync.WaitGroup{}
	a := 10000
	b := 2 * a
	wg.Add(b)

	chans := make([]chan struct{}, a)
	for i := 0; i < a; i++ {
		j := i
		kj := strconv.Itoa(j)
		chans[j] = make(chan struct{})
		go func() {
			m.Put(kj, j)
			chans[j] <- struct{}{}
			wg.Done()
		}()
	}

	for i := 0; i < a; i++ {
		j := i
		kj := strconv.Itoa(j)
		go func() {
			<-chans[j]
			vj, ok := m.Get(kj)
			wg.Done()
			if !ok {
				t.Log(kj)
				return
			}
			// assert.Equal(t, true, ok)
			assert.Equal(t, j, vj.(int))
		}()
	}

	wg.Wait()
}

func TestNewLockFreeMap(t *testing.T) {
	m := NewLockFreeMap()
	wg := sync.WaitGroup{}
	a := 100000
	b := 2 * a
	wg.Add(b)

	chans := make([]chan struct{}, a)
	for i := 0; i < a; i++ {
		j := i
		kj := strconv.Itoa(j)
		chans[j] = make(chan struct{})
		go func() {
			m.Put(kj, j)
			chans[j] <- struct{}{}
			wg.Done()
		}()
	}

	for i := 0; i < a; i++ {
		j := i
		kj := strconv.Itoa(j)
		go func() {
			<-chans[j]
			vj, ok := m.Get(kj)
			wg.Done()
			//if !ok {
			//	t.Log(kj)
			//	return
			//}
			assert.Equal(t, true, ok)
			assert.Equal(t, j, vj.(int))
		}()
	}

	wg.Wait()
}
