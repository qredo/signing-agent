package defs

import "errors"

var ErrNotFound = errors.New("not found")

type Err struct {
	error string
}
