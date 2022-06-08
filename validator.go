package resperr

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// Validator creates a map of fields to error messages.
type Validator url.Values

// Ensure adds the provided message to field if cond is not met.
// Ensure works with the zero value of Validator.
func (v *Validator) Ensure(field string, cond bool, message string, a ...any) {
	if cond {
		return
	}
	if *v == nil {
		*v = make(Validator)
	}
	(*url.Values)(v).Add(field, fmt.Sprintf(message, a...))
}

// EnsureIf adds the provided message to field if cond is not met and the field does not already have a validation message.
// EnsureIf works with the zero value of Validator.
func (v *Validator) EnsureIf(field string, cond bool, message string, a ...any) {
	if len((*v)[field]) > 0 {
		return
	}
	v.Ensure(field, cond, message, a...)
}

// Err transforms v to a ValidatorError if v is not empty.
// The error created shares the same underlying map reference as v.
func (v *Validator) Err() error {
	if len(*v) < 1 {
		return nil
	}
	return validatorErrors(*v)
}

// Valid reports whether v had any validation failures.
func (v *Validator) Valid() bool {
	return len(*v) == 0
}

// ValidationErrors returns any ValidationError found in err's error chain
// or an empty map.
func ValidationErrors(err error) url.Values {
	if ve := (ValidationError)(nil); err != nil && errors.As(err, &ve) {
		return ve.ValidationErrors()
	}
	return nil
}

// ValidationError is an error with an associated set of validation messages for request fields
type ValidationError interface {
	error
	ValidationErrors() url.Values
}

type validatorErrors url.Values

var _ ValidationError = validatorErrors{}
var _ StatusCoder = validatorErrors{}

func (ve validatorErrors) Error() string {
	s, _ := url.QueryUnescape(url.Values(ve).Encode())
	return fmt.Sprintf("validation error: %s", strings.ReplaceAll(s, "&", " "))
}

func (ve validatorErrors) ValidationErrors() url.Values {
	return url.Values(ve)
}

func (ve validatorErrors) StatusCode() int {
	return http.StatusBadRequest
}
