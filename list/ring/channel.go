package ring

import "sync"

// Channel action like go native channel
type Channel struct {
	full  *sync.Cond
	empty *sync.Cond

	*sync.Mutex

	dataArray *Ring
}

// NewChannel new a channel
func NewChannel(size uint64) *Channel {
	if size == 0 {
		size = 1 // use one size for store data
	}
	mu := &sync.Mutex{}
	return &Channel{
		full:      sync.NewCond(mu),
		empty:     sync.NewCond(mu),
		Mutex:     mu,
		dataArray: NewRing(size),
	}
}

// Put an item into channel
func (c *Channel) Put(item interface{}) error {
	c.Lock()
	defer c.Unlock()

	// 1. check full then wait
	for c.dataArray.checkRingFull() {
		c.full.Wait()
	}
	// 2. check empty then signal
	if c.dataArray.checkRingEmpty() {
		defer c.empty.Broadcast()
	}

	return c.dataArray.Put(item)
}

// Get an item from channel
func (c *Channel) Get() (interface{}, error) {
	c.Lock()
	defer c.Unlock()

	// 1. check empty then wait
	for c.dataArray.checkRingEmpty() {
		c.empty.Wait()
	}
	// 2. check full then signal
	if c.dataArray.checkRingFull() {
		defer c.full.Broadcast()
	}

	return c.dataArray.Get()
}
