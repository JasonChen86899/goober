package cache

import (
	"fmt"
	"math"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewCacheLoader(t *testing.T) {
	cache := NewLRUCache()

	f := func(key string) (interface{}, error) {
		t.Log("loader")
		time.Sleep(time.Second * 2)
		return "test", nil
	}
	s := "1"
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(t *testing.T, i int) {
			eOpts := []EntryOption{
				WithLoader(f),
				ExpirationOption(time.Minute),
			}
			rsp, err, _ := cache.Load(s, eOpts...)
			//assert.Error(t, err)
			t.Log(i, rsp, err)
			wg.Done()
		}(t, i)
	}

	wg.Wait()
}

func TestCacheClean(t *testing.T) {
	v := int(math.Pow(2, 20))
	cache := NewLRUCache(CleanDuration(time.Minute), MaxSize(v))

	f := func(key string) {
		group := sync.WaitGroup{}
		group.Add(v)
		count := int32(0)
		for i := 0; i < v; i++ {
			j := i
			go func(j int) {
				startTime := time.Now()
				f := func(string) (ii interface{}, err error) {
					return fmt.Sprintf("value_%d", j), nil
				}
				_, _, _ = cache.Load(fmt.Sprintf("%s_%d", key, j), WithLoader(f), ExpirationOption(time.Hour))
				if time.Since(startTime).Seconds() > 3 {
					atomic.AddInt32(&count, 1)
				}
				group.Done()
			}(j)
			time.Sleep(10 * time.Nanosecond)
		}

		group.Wait()
		t.Log(count)
	}

	f("1")
	f("2")
}

func TestOldCacheLoader(t *testing.T) {
	cache := NewLRUCache(CleanDuration(2*time.Minute), MaxSize(10240))
	f := func(string2 string) (interface{}, error) {
		time.Sleep(1 * time.Millisecond)
		t.Logf("test called at :%d", time.Now().UnixNano()/1000)
		return "1", nil
	}
	w := sync.WaitGroup{}
	roundNum := 2
	gNum := 2000000
	w.Add(gNum)
	start := time.Now().UnixNano()
	for j := 0; j <= roundNum; j++ {
		key := strconv.Itoa(j % 3)
		_, _, _ = cache.Load(key, WithLoader(f), ExpirationOption(2*time.Minute))
	}
	t.Logf("test start at :%d", start/1000)
	for i := 0; i < gNum; i++ {
		go func(i int) {
			for j := 0; j <= roundNum; j++ {
				key := strconv.Itoa(j % 3)
				_, _, _ = cache.Load(key, WithLoader(f), ExpirationOption(1*time.Minute))
			}
			w.Done()
		}(i)
	}
	w.Wait()
	end := time.Now().UnixNano()
	t.Logf("test end at :%d", end/1000)
	t.Logf("test cost: %d", (end-start)/1000)
}
