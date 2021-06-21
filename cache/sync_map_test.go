package cache

import (
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"
)

func TestMap_Loader(t *testing.T) {
	m := Map{}

	f := func() (interface{}, error) {
		time.Sleep(100 * time.Millisecond)
		fmt.Println("test function")
		return "1", nil
	}
	w := sync.WaitGroup{}
	gNum := 2000000
	w.Add(gNum)
	once := sync.Once{}
	for i := 0; i < gNum; i++ {
		go func() {
			_, _ = m.Loader("test_1", f)
			//_, _ = m.Loader("test_2", valFunc)
			//_, _ = m.Loader("test_3", valFunc)
			//_, _ = m.Loader("test_4", valFunc)
			//_, _ = m.Loader("test_5", valFunc)
			//_, _ = m.Loader("test_6", valFunc)
			//_, _ = m.Loader("test_7", valFunc)
			//_, _ = m.Loader("test_8", valFunc)
			w.Done()
		}()
		if i == gNum/2 {
			once.Do(func() {
				m.Delete("test_1")
			})
		}
	}
	w.Wait()
}

func TestMap_LoaderWithExpired(t *testing.T) {
	m := Map{}

	f := func() (interface{}, error) {
		time.Sleep(100 * time.Millisecond)
		fmt.Println("test function")
		return "1", nil
	}

	//if v, err := m.LoaderWithExpired("test", valFunc, time.Minute); err != nil {
	//	t.Fatal(err)
	//}

	w := sync.WaitGroup{}
	gNum := 2000000
	w.Add(gNum)
	//once := sync.Once{}
	n := time.Now().Unix()
	for i := 0; i < gNum; i++ {
		go func() {
			_, _ = m.LoaderWithExpired("test_1", f, 2*time.Second)
			//_, _ = m.LoaderWithExpired("test_2", valFunc, time.Minute)
			//_, _ = m.LoaderWithExpired("test_3", valFunc, time.Minute)
			//_, _ = m.LoaderWithExpired("test_4", valFunc, time.Minute)
			//_, _ = m.LoaderWithExpired("test_5", valFunc, time.Minute)
			//_, _ = m.LoaderWithExpired("test_6", valFunc, time.Minute)
			//_, _ = m.LoaderWithExpired("test_7", valFunc, time.Minute)
			//_, _ = m.LoaderWithExpired("test_8", valFunc, time.Minute)
			w.Done()
		}()
		//if i == gNum/2 {
		//	once.Do(func() {
		//		m.Delete("test_1")
		//	})
		//}
	}
	w.Wait()
	fmt.Println("execute: ", time.Now().Unix()-n)
	fmt.Println("start sleep")
	time.Sleep(10 * time.Second)
	fmt.Println("sleep 10s")
	if v, err := m.LoaderWithExpired("test_1", f, time.Minute); err != nil {
		t.Fatal(err)
	} else {
		fmt.Println(v)
	}
}

func TestStructAddress(T *testing.T) {
	a := expiredItem{}

	b := a

	fmt.Println("a: ", unsafe.Pointer(&a))
	fmt.Println("b: ", unsafe.Pointer(&b))
}

func TestSyncMapCacheLoader(t *testing.T) {
	cache := Map{}
	f := func() (interface{}, error) {
		time.Sleep(1 * time.Millisecond)
		fmt.Printf("test called at :%d\n", time.Now().UnixNano()/1000)
		return "1", nil
	}
	w := sync.WaitGroup{}
	roundNum := 2
	gNum := 2000000
	for j := 0; j < roundNum; j++ {
		key := strconv.Itoa(j % 3)
		_, _ = cache.LoaderWithExpired(key, f, 2*time.Minute)
	}
	w.Add(gNum)
	start := time.Now().UnixNano()
	t.Logf("test start at :%d\n", start/1000)
	for i := 0; i < gNum; i++ {
		go func(i int) {
			for j := 0; j <= roundNum; j++ {
				key := strconv.Itoa(j % 3)
				_, _ = cache.LoaderWithExpired(key, f, 2*time.Minute)
			}
			w.Done()
		}(i)
	}
	w.Wait()
	end := time.Now().UnixNano()
	t.Logf("test end at :%d\n", end/1000)
	t.Logf("test cost: %d\n", (end-start)/1000)
}

type ifaceWords struct {
	typ  unsafe.Pointer
	data unsafe.Pointer
	ptr  *uint64
}

func TestAtomicOps(t *testing.T) {
	var num uint64
	num = 1000
	v := &ifaceWords{
		data: unsafe.Pointer(&num),
		typ:  unsafe.Pointer(&num),
		ptr:  &num,
	}
	round := 100000000
	start := time.Now().UnixNano()
	for i := 0; i < round; i++ {
		vp := (*ifaceWords)(unsafe.Pointer(v))
		d := atomic.LoadPointer(&vp.typ)
		l := atomic.LoadPointer(&vp.typ)
		p := atomic.CompareAndSwapUint64(&num, 0, 0)
		_ = d
		_ = l
		_ = p
	}
	end := time.Now().UnixNano()
	t.Logf("LoadPointer & CAS test cost: %d", (end-start)/1000)
	start = time.Now().UnixNano()
	for i := 0; i < round; i++ {
		k := atomic.AddUint64(&num, 1)
		_ = k
	}
	end = time.Now().UnixNano()
	t.Logf("AddUint64 test cost: %d", (end-start)/1000)
}
