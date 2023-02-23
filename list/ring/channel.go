package ring

import "sync"

type Channel struct {
	full  *sync.Cond
	empty *sync.Cond

	*sync.Mutex

	dataArray *Ring
}

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

func (c *Channel) Put(item interface{}) error {
	c.Lock()
	defer c.Unlock()

	if c.dataArray.checkRingFull() {
		c.full.Wait()
	}

	if c.dataArray.checkRingEmpty() {
		defer c.empty.Signal()
	}

	return c.dataArray.Put(item)
}

func (c *Channel) Get() (interface{}, error) {
	c.Lock()
	defer c.Unlock()

	if c.dataArray.checkRingEmpty() {
		c.empty.Wait()
	}

	if c.dataArray.checkRingFull() {
		defer c.full.Signal()
	}

	return c.dataArray.Get()
}
