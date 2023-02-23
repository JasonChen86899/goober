package ring

import "testing"

func TestNewList(t *testing.T) {
	rl := NewRing(8)

	var err error
	var item interface{}

	for i := 0; i < 7; i++ {
		err = rl.Put(i)
		if err != nil {
			t.Log(err)
		}
	}

	for i := 0; i < 7; i++ {
		item, err = rl.Get()
		if err != nil {
			t.Log(err)
		} else {
			t.Log(item)
		}
	}

	t.Log(rl.Length())
}
