package wrappers

import "fmt"

type WrappedError string

var DoesNotExist WrappedError = "does not exist"
var	NotFound  WrappedError = "not found"

func newErrWithString(s string, w WrappedError) error {
	return fmt.Errorf("%s %s", s, w)
}


func NewErrDoesNotExist(s string) error {
	return newErrWithString(s, DoesNotExist)
}

func NewErrNotFound(s string) error {
	return newErrWithString(s, NotFound)
}