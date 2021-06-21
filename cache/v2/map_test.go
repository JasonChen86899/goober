package v2

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestAtomicValue(t *testing.T) {
	var v atomic.Value

	t.Log(v.Load())
}

func TestBucket_Get(t *testing.T) {

}

func TestBucket_Put(t *testing.T) {
	b := newBucket()

	w := sync.WaitGroup{}
	w.Add(100)
	for i := 0; i < 100; i++ {
		j := i
		go func() {
			b.Put(j, j)
			//w.Done()
		}()

		go func() {
			v, ok := b.Get(j)
			defer w.Done()

			if !ok {
				t.Log(false)
				return
			}

			t.Log(v.(int) == j)
		}()
	}

	w.Wait()

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
