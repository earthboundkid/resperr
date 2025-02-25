// Package resperr contains helpers for associating http status codes
// and user messages with errors
package resperr

import (
	"errors"
	"fmt"
	"net/http"
)

// StatusCoder is an error with an associated HTTP status code
type StatusCoder interface {
	error
	StatusCode() int
}

// StatusCode returns the status code associated with an error.
// If no status code is found, it returns 500 http.StatusInternalServerError.
// As a special case, it checks for Timeout() and Temporary() errors and returns
// 504 http.StatusGatewayTimeout and 503 http.StatusServiceUnavailable
// respectively.
// If err is nil, it returns 200 http.StatusOK.
func StatusCode(err error) (code int) {
	if err == nil {
		return http.StatusOK
	}
	if sc := StatusCoder(nil); errors.As(err, &sc) {
		return sc.StatusCode()
	}
	var timeouter interface {
		error
		Timeout() bool
	}
	if errors.As(err, &timeouter) && timeouter.Timeout() {
		return http.StatusGatewayTimeout
	}
	var temper interface {
		error
		Temporary() bool
	}
	if errors.As(err, &temper) && temper.Temporary() {
		return http.StatusServiceUnavailable
	}
	return http.StatusInternalServerError
}

// UserMessenger is an error with an associated user-facing message
type UserMessenger interface {
	error
	UserMessage() string
}

// UserMessage returns the user message associated with an error.
// If no message is found, it checks StatusCode and returns that message.
// Because the default status is 500, the default message is
// "Internal Server Error".
// If err is nil, it returns "".
func UserMessage(err error) string {
	if err == nil {
		return ""
	}
	if um := UserMessenger(nil); errors.As(err, &um) {
		return um.UserMessage()
	}
	return http.StatusText(StatusCode(err))
}

// NotFound creates an error with a 404 status code and a user message
// showing the request path that was not found.
func NotFound(r *http.Request) error {
	return E{
		S: http.StatusNotFound,
		M: fmt.Sprintf("could not find path %q", r.URL.Path),
	}
}

// New is a convenience function for calling fmt.Errorf.
func New(code int, format string, v ...any) error {
	return E{
		S: code,
		E: fmt.Errorf(format, v...),
	}
}
