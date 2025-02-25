package resperr

import (
	"cmp"
	"errors"
	"fmt"
	"net/http"
)

// E is a simple struct for building response errors.
type E struct {
	S int    // StatusCode
	M string // UserMessage
	E error  // Cause
}

func (e E) Error() string {
	err := e.E
	m := e.M
	// Flatten chains of E
	for {
		ee, ok := err.(E)
		if !ok {
			break
		}
		m = cmp.Or(m, ee.M)
		err = ee.E
	}
	if err == nil {
		err = errors.New(http.StatusText(e.StatusCode()))
	}
	if m != "" {
		return fmt.Sprintf("[%d] <%s> %v", e.StatusCode(), e.M, err.Error())
	}
	return fmt.Sprintf("[%d] %v", e.StatusCode(), err.Error())
}

func (e E) Unwrap() error { return e.E }

func (e E) StatusCode() int {
	if e.S != 0 {
		return e.S
	}
	if e.E != nil {
		return StatusCode(e.E)
	}
	if e.M != "" {
		return http.StatusBadRequest
	}
	return http.StatusInternalServerError
}

func (e E) UserMessage() string {
	if e.M != "" {
		return e.M
	}
	if um := UserMessenger(nil); errors.As(e.E, &um) {
		return um.UserMessage()
	}
	return http.StatusText(StatusCode(e))
}
