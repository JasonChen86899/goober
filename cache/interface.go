package cache

import (
	"errors"
	"go.uber.org/atomic"
)

var (
	// ErrLoadTimeout load timeout error
	ErrLoadTimeout = errors.New("load timeout")
)

// Loader loader function
type Loader func(string) *Value

// Value loader function return innerValue
type Value struct {
	// Val store real value of Loader or Put
	Val interface{}

	// Err store err infos when Loader function has failed
	Err error
}

// A Cache is a generalized interface to a cache.  See cache.LRU for a specific
// implementation (bounded cache with LRU eviction)
type Cache interface {
	// Get retrieves an element based on a key
	// return false if the element does not exist
	Get(key string) (value *Value, existed bool)

	// Put adds an element to the cache, returning the previous element
	Put(key string, value interface{}, opts ...EntryOption) interface{}

	// Delete deletes an element in the cache
	Delete(key string)

	// Size returns the number of entries currently stored in the Cache
	Size() int

	// CompareAndSwap adds an element to the cache if the existing entry matches the old innerValue.
	// It returns the element in cache after function is executed and true if the element was replaced, false otherwise.
	CompareAndSwap(key string, old, new interface{}, opts ...EntryOption) (interface{}, bool)

	// Load innerValue by call loader function.
	// if you need add call f() timeout, can set EntryOption.loadTimeout
	// The loaded result is true if the value was loaded, false if calling loader function.
	Load(cacheKey string, opts ...EntryOption) (value *Value, loaded bool)

	// Load1(cacheKey string, opts ...EntryOption) (innerValue interface{}, err error, loaded bool)

}

// internalEntry inner entry of key in map
type internalEntry struct {
	key        string
	innerValue *Value
	expiration int64
	refresh    int64

	refreshed *atomic.Bool
}
