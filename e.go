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
	code := e.StatusCode()
	if code == 0 {
		if m != "" {
			code = http.StatusBadRequest
		} else {
			code = http.StatusInternalServerError
		}
	}
	if err == nil {
		err = errors.New(http.StatusText(code))
	}
	if m != "" {
		return fmt.Sprintf("[%d] <%s> %v", code, e.M, err.Error())
	}
	return fmt.Sprintf("[%d] %v", code, err.Error())
}

func (e E) Unwrap() error { return e.E }

func (e E) StatusCode() int { return e.S }

func (e E) UserMessage() string { return e.M }
