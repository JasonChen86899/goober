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

func (cache *LRUCache) lruInsert(cacheEntry *internalEntry) *list.Element {
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
	delete(cache.values, e.Value.(*internalEntry).key)
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
		if tmp.(*internalEntry).expiration <= time.Now().UnixNano() {
			cache.deleteItem(back)
		}
		back = back.Prev()
	}
}

func (cache *LRUCache) cleanFull() {
	for i := 0; i < cache.opts.cleanSize && cache.lruList.Len() > 0; i++ {
		e := cache.lruList.Back()
		cache.lruList.Remove(e)
		delete(cache.values, e.Value.(*internalEntry).key)
	}
}

func (cache *LRUCache) callLoader(key string, eOpts EntryOptions) *internalEntry {
	loadCtx, loadCancel := context.WithTimeout(context.Background(), eOpts.loadTimeout)
	defer loadCancel()

	retChan := make(chan *Value)
	go func() {
		ret := eOpts.loader(key)
		retChan <- ret
	}()

	var ret *Value
	select {
	case <-loadCtx.Done():
		ret = &Value{}
		ret.Err = fmt.Errorf(
			"function: %s, %w",
			runtime.FuncForPC(reflect.ValueOf(eOpts.loader).Pointer()).Name(),
			ErrLoadTimeout)
	case ret = <-retChan:
		// block until loader function return within loadTimeout
	}

	cacheEntry := &internalEntry{
		key:       key,
		refreshed: atomic.NewBool(false),
	}

	if ret.Err != nil {
		// do this to protect f() when f() return err in high concurrent query
		cacheEntry.expiration = time.Now().Add(500 * time.Millisecond).UnixNano()
	} else {
		cacheEntry.expiration = time.Now().Add(eOpts.expireAfterWrite).UnixNano()
	}
	cacheEntry.innerValue = ret

	return cacheEntry
}

func (cache *LRUCache) asyncRefreshItem(cacheKey string, eOpts EntryOptions) {
	// call loader
	item := cache.callLoader(cacheKey, eOpts)

	cache.lock.Lock()
	defer cache.lock.Unlock()

	v, ok := cache.values[cacheKey]
	// the key may be cleaned by the asyncClean goroutine.
	// if cleaned, should be inserted to cache again, otherwise just re-assign.
	if ok {
		v.Value = item
		cache.lruMoveToFront(v)
	} else {
		e := cache.lruInsert(item)
		cache.values[cacheKey] = e
	}
}

func (cache *LRUCache) Load(cacheKey string, opts ...EntryOption) (*Value, bool) {
	eOpts := cache.opts.defaultEntryOpts
	for _, o := range opts {
		o(&eOpts)
	}

	// loader function is nil, then return Get(key) innerValue
	if eOpts.loader == nil {
		return cache.Get(cacheKey)
	}

	cache.lock.RLock()
	e, ok := cache.values[cacheKey]
	// check cache hist or not, and if cache hist and expired, async refresh this CacheEntry
	if ok {
		entry := e.Value.(*internalEntry)
		expired := entry.expiration <= time.Now().UnixNano()
		// not expired or async load just return innerValue in cache
		if !expired || !eOpts.syncLoad {
			defer cache.lock.RUnlock()

			// expired and async load need async refresh
			if expired && !eOpts.syncLoad {
				if entry.refreshed.CAS(false, true) {
					go cache.asyncRefreshItem(cacheKey, eOpts)
				}
			}

			return entry.innerValue, true
		}
	}
	cache.lock.RUnlock()

	cache.lock.Lock()
	defer cache.lock.Unlock()
	e, ok = cache.values[cacheKey]
	if ok {
		return e.Value.(*internalEntry).innerValue, true
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
	cacheEntry := cache.callLoader(cacheKey, eOpts)
	e = cache.lruInsert(cacheEntry)
	cache.values[cacheKey] = e

	return cacheEntry.innerValue, false
}

func (cache *LRUCache) Delete(cacheKey string) {
	cache.lock.Lock()
	defer cache.lock.Unlock()

	// check cache innerValue exist
	if v, ok := cache.values[cacheKey]; ok {
		cache.lruList.Remove(v)
		delete(cache.values, cacheKey)
	}
}

func (cache *LRUCache) Get(key string) (*Value, bool) {
	cache.lock.RLock()
	defer cache.lock.RUnlock()

	// check err
	e, ok := cache.values[key]
	if ok {
		entry := e.Value.(*internalEntry)
		return entry.innerValue, true
	}

	return nil, false
}

func (cache *LRUCache) Put(key string, value interface{}, opts ...EntryOption) interface{} {
	eOpts := cache.opts.defaultEntryOpts
	for _, o := range opts {
		o(&eOpts)
	}

	cache.lock.Lock()
	defer cache.lock.Unlock()

	e, ok := cache.values[key]
	cacheEntry := &internalEntry{
		key:        key,
		innerValue: &Value{Val: value},
		expiration: time.Now().Add(eOpts.expireAfterWrite).UnixNano(),
		refreshed:  atomic.NewBool(false),
	}

	if ok {
		pre := e.Value
		e.Value = cacheEntry
		cache.lruMoveToFront(e)

		return pre.(*internalEntry).innerValue
	}

	e = cache.lruInsert(cacheEntry)
	cache.values[key] = e

	return cacheEntry.innerValue
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
		pre = e.Value.(*internalEntry).innerValue
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
		pre = e.Value.(*internalEntry).innerValue
	}

	// not equal, change by other goroutine
	if pre != old {
		return pre, false
	}

	cacheEntry := &internalEntry{
		key:        key,
		innerValue: &Value{Val: new},
		expiration: time.Now().Add(eOpts.expireAfterWrite).UnixNano(),
		refreshed:  atomic.NewBool(false),
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
