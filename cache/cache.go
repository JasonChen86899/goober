package cache

import (
	"errors"

	"go.uber.org/atomic"
)

var (
	ErrLoadTimeout = errors.New("load timeout")
)

type Loader func(string) (interface{}, error)

type FRet struct {
	Res interface{}
	Err error
}

type Entry struct {
	key        string
	value      interface{}
	expiration int64
	refresh    int64

	Refreshed *atomic.Bool
}

// A Cache is a generalized interface to a cache.  See cache.LRU for a specific
// implementation (bounded cache with LRU eviction)
type Cache interface {
	// Get retrieves an element based on a key
	//returning false if the element does not exist
	Get(key string) (value interface{}, err error, existed bool)

	// Put adds an element to the cache, returning the previous element
	Put(key string, value interface{}, opts ...EntryOption) interface{}

	// Delete deletes an element in the cache
	Delete(key string)

	// Size returns the number of entries currently stored in the Cache
	Size() int

	// CompareAndSwap adds an element to the cache if the existing entry matches the old value.
	// It returns the element in cache after function is executed and true if the element was replaced, false otherwise.
	CompareAndSwap(key string, old, new interface{}, opts ...EntryOption) (interface{}, bool)

	// Load load value by call loader function.
	// If need add call f() timeout, can wrapper f(string) and call context inner
	Load(cacheKey string, opts ...EntryOption) (value interface{}, err error, loaded bool)
}
