package list

import "errors"

var (
	ErrNotFound = errors.New("list element not found")
	ErrFull     = errors.New("list full")
	ErrEmpty    = errors.New("ring empty")
)

type List interface {
	// Put an element into list
	Put(interface{})
	// Get an element
	Get() (interface{}, error)
}
