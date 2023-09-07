package cache

import "container/list"

type Element struct {
	*list.Element
	Cnt int
}
