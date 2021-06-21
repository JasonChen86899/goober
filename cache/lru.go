package cache

import (
	"container/list"
	"context"
	"fmt"
	"reflect"
	"runtime"
	"sync"
	"time"

	"go.uber.org/atomic"
)

type LRUCache struct {
	lock   sync.RWMutex
	values map[string]*list.Element

	lruLock sync.Mutex
	lruList *list.List

	cleanFullChan      chan struct{}
	cleanFullThreshold int

	opts Options
}

func NewLRUCache(opt ...Option) *LRUCache {
	opts := NewOptions()
	for _, o := range opt {
		o(&opts)
	}

	cache := &LRUCache{
		lock:   sync.RWMutex{},
		values: make(map[string]*list.Element, opts.maxSize),

		lruLock: sync.Mutex{},
		lruList: (&list.List{}).Init(),

		cleanFullChan:      make(chan struct{}, 1),
		cleanFullThreshold: int(float64(opts.maxSize) * opts.cleanFullThresholdPercent),

		opts: opts,
	}

	go cache.asyncClean()

	return cache
}

func (cache *LRUCache) lruMoveToFront(e *list.Element) {
	cache.lruLock.Lock()
	defer cache.lruLock.Unlock()

	cache.lruList.MoveToFront(e)
}

func (cache *LRUCache) lruInsert(cacheEntry *Entry) *list.Element {
	return cache.lruList.PushFront(cacheEntry)
}

func (cache *LRUCache) lruRemove(e *list.Element) {
	cache.lruList.Remove(e)
}

func (cache *LRUCache) asyncClean() {
	cleanDuration := defaultCleanDuration
	if cache.opts.cleanDuration > cleanDuration {
		cleanDuration = cache.opts.cleanDuration
	}

	t := time.NewTicker(cleanDuration)
	for {
		select {
		case <-t.C:
			cache.cleanExpired()
		case <-cache.cleanFullChan:
			cache.lock.Lock()
			cache.cleanFull()
			cache.lock.Unlock()
		}
	}
}

func (cache *LRUCache) deleteItem(e *list.Element) {
	cache.lruList.Remove(e)
	delete(cache.values, e.Value.(*Entry).key)
}

func (cache *LRUCache) cleanExpired() {
	if cache.lruList.Len() == 0 {
		return
	}
	cache.lock.Lock()
	defer cache.lock.Unlock()
	back := cache.lruList.Back()
	for i := 0; i < cache.opts.cleanSize && back != nil; i++ {
		tmp := back.Value
		if tmp.(*Entry).expiration <= time.Now().UnixNano() {
			cache.deleteItem(back)
		}
		back = back.Prev()
	}
}

func (cache *LRUCache) cleanFull() {
	for i := 0; i < cache.opts.cleanSize && cache.lruList.Len() > 0; i++ {
		e := cache.lruList.Back()
		cache.lruList.Remove(e)
		delete(cache.values, e.Value.(*Entry).key)
	}
}

func (cache *LRUCache) callLoader(key string, eOpts EntryOptions) (*Entry, error) {
	loadCtx, loadCancel := context.WithTimeout(context.Background(), eOpts.loadTimeout)
	defer loadCancel()

	retChan := make(chan FRet)
	go func() {
		res, err := eOpts.loader(key)
		retChan <- FRet{
			Res: res,
			Err: err,
		}
	}()

	ret := FRet{}
	select {
	case <-loadCtx.Done():
		ret.Err = fmt.Errorf(
			"function: %s, %w",
			runtime.FuncForPC(reflect.ValueOf(eOpts.loader).Pointer()).Name(),
			ErrLoadTimeout)
	case ret = <-retChan:
		// block until loader function return within loadTimeout
	}

	cacheEntry := &Entry{
		key:       key,
		Refreshed: atomic.NewBool(false),
	}

	if ret.Err != nil {
		// do this to protect f() when f() return err in high concurrent query
		cacheEntry.expiration = time.Now().Add(500 * time.Millisecond).UnixNano()
		cacheEntry.value = ret.Err
	} else {
		cacheEntry.expiration = time.Now().Add(eOpts.expireAfterWrite).UnixNano()
		cacheEntry.value = ret.Res
	}

	return cacheEntry, ret.Err
}

func (cache *LRUCache) asyncRefreshItem(cacheKey string, eOpts EntryOptions) {
	// call loader
	item, _ := cache.callLoader(cacheKey, eOpts)

	cache.lock.Lock()
	defer cache.lock.Unlock()

	v, ok := cache.values[cacheKey]
	// the key may be cleaned by the asyncClean goroutine.
	// if cleaned, should be insert to cache again, otherwise just re-assign.
	if ok {
		v.Value = item
		cache.lruMoveToFront(v)
	} else {
		e := cache.lruInsert(item)
		cache.values[cacheKey] = e
	}
}

func (cache *LRUCache) Load(cacheKey string, opts ...EntryOption) (interface{}, error, bool) {
	eOpts := cache.opts.defaultEntryOpts
	for _, o := range opts {
		o(&eOpts)
	}

	// loader function is nil, then return Get(key) value
	if eOpts.loader == nil {
		v, err, ok := cache.Get(cacheKey)
		if !ok {
			return nil, nil, false
		}

		if err != nil {
			return nil, err, true
		}
		return v, nil, true
	}

	cache.lock.RLock()
	e, ok := cache.values[cacheKey]
	// check cache hist or not, and if cache hist and expired, async refresh this CacheEntry
	if ok {
		entry := e.Value.(*Entry)

		expired := entry.expiration <= time.Now().UnixNano()
		// not expired or async load just return value in cache
		if !expired || !eOpts.syncLoad {
			defer cache.lock.RUnlock()

			// expired and async load need async refresh
			if expired && !eOpts.syncLoad {
				if entry.Refreshed.CAS(false, true) {
					go cache.asyncRefreshItem(cacheKey, eOpts)
				}
			}

			// check err
			if err, ok := entry.value.(error); ok {
				return nil, err, true
			}

			return e.Value.(*Entry).value, nil, true
		}
	}
	cache.lock.RUnlock()

	cache.lock.Lock()
	defer cache.lock.Unlock()
	e, ok = cache.values[cacheKey]
	if ok {
		if err, ok := e.Value.(*Entry).value.(error); ok {
			return nil, err, true
		}

		return e.Value.(*Entry).value, nil, true
	}

	// check cache if full
	if cache.lruList.Len() >= cache.cleanFullThreshold {
		// attention: if size larger or equal maxsize, sync clean full!
		if cache.lruList.Len() >= cache.opts.maxSize {
			cache.cleanFull()
		}

		select {
		case cache.cleanFullChan <- struct{}{}:
			//do nothing
		default:
			// do nothing
		}
	}

	// call loader
	cacheEntry, err := cache.callLoader(cacheKey, eOpts)
	e = cache.lruInsert(cacheEntry)
	cache.values[cacheKey] = e

	if err != nil {
		return nil, err, true
	}

	return cacheEntry.value, nil, true
}

func (cache *LRUCache) Delete(cacheKey string) {
	cache.lock.Lock()
	defer cache.lock.Unlock()

	// check cache value exist
	if v, ok := cache.values[cacheKey]; ok {
		cache.lruList.Remove(v)
		delete(cache.values, cacheKey)
	}
}

func (cache *LRUCache) Get(key string) (interface{}, error, bool) {
	cache.lock.RLock()
	defer cache.lock.RUnlock()

	// check err
	e, ok := cache.values[key]
	if ok {
		entry := e.Value.(*Entry)
		if err, ok := entry.value.(error); ok {
			return nil, err, true
		}

		return entry.value, nil, true
	}

	return nil, nil, false

}

func (cache *LRUCache) Put(key string, value interface{}, opts ...EntryOption) interface{} {
	eOpts := cache.opts.defaultEntryOpts
	for _, o := range opts {
		o(&eOpts)
	}

	cache.lock.Lock()
	defer cache.lock.Unlock()

	e, ok := cache.values[key]
	cacheEntry := &Entry{
		key:        key,
		value:      value,
		expiration: time.Now().Add(eOpts.expireAfterWrite).UnixNano(),
		Refreshed:  atomic.NewBool(false),
	}

	if ok {
		pre := e.Value
		e.Value = cacheEntry
		cache.lruMoveToFront(e)

		return pre.(*Entry).value
	}

	e = cache.lruInsert(cacheEntry)
	cache.values[key] = e

	return cacheEntry.value
}

func (cache *LRUCache) Size() int {
	cache.lock.RLock()
	defer cache.lock.RUnlock()

	return len(cache.values)
}

func (cache *LRUCache) CompareAndSwap(key string, old, new interface{}, opts ...EntryOption) (interface{}, bool) {
	eOpts := cache.opts.defaultEntryOpts
	for _, o := range opts {
		o(&eOpts)
	}

	cache.lock.RLock()
	e, ok := cache.values[key]

	var pre interface{}
	if !ok {
		pre = nil
	} else {
		pre = e.Value.(*Entry).value
	}

	// not equal
	if pre != old {
		cache.lock.RUnlock()
		return pre, false
	}

	// equal
	cache.lock.RUnlock()

	cache.lock.Lock()
	defer cache.lock.Unlock()

	e, ok = cache.values[key]
	if !ok {
		pre = nil
	} else {
		pre = e.Value.(*Entry).value
	}

	// not equal, change by other goroutine
	if pre != old {
		return pre, false
	}

	cacheEntry := &Entry{
		key:        key,
		value:      new,
		expiration: time.Now().Add(eOpts.expireAfterWrite).UnixNano(),
		Refreshed:  atomic.NewBool(false),
	}

	// check if exist
	if ok {
		e.Value = cacheEntry
		cache.lruMoveToFront(e)
	} else {
		e = cache.lruInsert(cacheEntry)
		cache.values[key] = e
	}

	return new, true
}
