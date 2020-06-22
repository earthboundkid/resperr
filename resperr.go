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

type statusCoder struct {
	error
	code int
}

func (sc statusCoder) Unwrap() error {
	return sc.error
}

func (sc statusCoder) Error() string {
	return fmt.Sprintf("[%d] %v", sc.code, sc.error)
}

func (sc statusCoder) StatusCode() int {
	return sc.code
}

// WithStatusCode adds a StatusCoder to err's error chain.
// Unlike pkg/errors, WithStatusCode will wrap nil error.
func WithStatusCode(err error, code int) error {
	if err == nil {
		err = errors.New(http.StatusText(code))
	}
	return statusCoder{err, code}
}

// StatusCode returns the status code associated with an error.
// If no status code is found, it returns 500 http.StatusInternalServerError.
// If err is nil, it returns 200 http.StatusOK.
func StatusCode(err error) (code int) {
	if err == nil {
		return http.StatusOK
	}
	if sc := StatusCoder(nil); errors.As(err, &sc) {
		return sc.StatusCode()
	}
	return http.StatusInternalServerError
}

// UserMessenger is an error with an associated user-facing message
type UserMessenger interface {
	error
	UserMessage() string
}

type messenger struct {
	error
	msg string
}

func (msgr messenger) Unwrap() error {
	return msgr.error
}

func (msgr messenger) UserMessage() string {
	return msgr.msg
}

// WithUserMessage adds a UserMessenger to err's error chain.
// Unlike pkg/errors, WithUserMessage will wrap nil error.
func WithUserMessage(err error, msg string) error {
	if err == nil {
		err = errors.New(msg)
	}
	return messenger{err, msg}
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

// WithCodeAndMessage is a convenience function for calling both
// WithStatusCode and WithUserMessage.
func WithCodeAndMessage(err error, code int, msg string) error {
	return WithStatusCode(WithUserMessage(err, msg), code)
}

// NotFound creates an error with a 404 status code and a user message
// showing the request path that was not found.
func NotFound(r *http.Request) error {
	return WithCodeAndMessage(
		nil,
		http.StatusNotFound,
		fmt.Sprintf("could not find path %q", r.URL.Path),
	)
}
