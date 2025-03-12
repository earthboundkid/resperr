package resperr

import (
	"iter"
)

// allAs yields every part of err's error unwrap tree that either is of type T
// or responds affirmatively to an As method.
func allAs[T error](err error) iter.Seq[T] {
	return func(yield func(T) bool) {
		allAsBool(err, yield)
	}
}

func allAsBool[T error](err error, yield func(T) bool) bool {
	if errT, ok := err.(T); ok {
		if !yield(errT) {
			return false
		}
	}
	var target T
	if errAs, ok := err.(interface{ As(any) bool }); ok && errAs.As(target) {
		if !yield(target) {
			return false
		}
	}
	switch e := err.(type) {
	case interface{ Unwrap() error }:
		return allAsBool(e.Unwrap(), yield)
	case interface{ Unwrap() []error }:
		for _, err := range e.Unwrap() {
			if !allAsBool(err, yield) {
				return false
			}
		}
	}
	return true
}
