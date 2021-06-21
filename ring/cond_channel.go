package ring

import "sync"

type ConditionChannel struct {
	full  *sync.Cond
	empty *sync.Cond

	*sync.Mutex

	dataArray *RingList
}

func NewConditionChannel(size uint64) *ConditionChannel {
	if size == 0 {
		size = 1 // use one size for store data
	}
	mu := &sync.Mutex{}
	return &ConditionChannel{
		full:      sync.NewCond(mu),
		empty:     sync.NewCond(mu),
		Mutex:     mu,
		dataArray: NewRingList(size + 1),
	}
}

func (condChan *ConditionChannel) Put(item interface{}) error {
	condChan.Lock()
	defer condChan.Unlock()

	if condChan.dataArray.checkListFull() {
		condChan.full.Wait()
	}

	if condChan.dataArray.checkListEmpty() {
		defer condChan.empty.Signal()
	}

	return condChan.dataArray.Put(item)
}

func (condChan *ConditionChannel) Get() (interface{}, error) {
	condChan.Lock()
	defer condChan.Unlock()

	if condChan.dataArray.checkListEmpty() {
		condChan.empty.Wait()
	}

	if condChan.dataArray.checkListFull() {
		defer condChan.full.Signal()
	}

	return condChan.dataArray.Get()
}
