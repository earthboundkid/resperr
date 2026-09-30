package resperr_test

import (
	"fmt"
	"net/http"
	"strconv"
	"testing"

	"github.com/earthboundkid/assert"
	"github.com/earthboundkid/resperr/v2"
)

func ExampleValidator() {
	var v resperr.Validator
	v.AddIf("heads", 2 > 1, "Two are better than one.")
	v.AddIf("heads", true, "I win, tails you lose.")
	err := v.Err()

	fmt.Println(resperr.StatusCode(err))
	for field, msgs := range resperr.ValidationErrors(err) {
		for _, msg := range msgs {
			fmt.Println(field, "=", msg)
		}
	}
	// Output:
	// 400
	// heads = Two are better than one.
	// heads = I win, tails you lose.
}

func ExampleValidator_AddIfUnset() {
	var v resperr.Validator
	x, err := strconv.Atoi("hello")
	v.AddIf("x", err != nil, "Could not parse x.")
	v.AddIf("x", x < 1, "X must be positive.")

	y, err := strconv.Atoi("hello")
	v.AddIf("y", err != nil, "Could not parse y.")
	v.AddIfUnset("y", y < 1, "Y must be positive.")
	fmt.Println(v.Err())
	// Output:
	// validation error: x=Could not parse x. x=X must be positive. y=Could not parse y.
}

func TestValidator(t *testing.T) {
	be := assert.FailsNow(t)
	var v1 resperr.Validator
	v1.AddIf("heads", 2 > 1, "Two are better than one.")
	v1.AddIf("heads", true, "I win, tails you lose.")
	err := v1.Err()
	fields := resperr.ValidationErrors(err)
	be.
		Truthy(err).
		False(v1.Valid()).
		EqualLength(fields, 1).
		EqualLength(fields["heads"], 2).
		Equal(resperr.StatusCode(err), http.StatusBadRequest)

	var v2 resperr.Validator
	v2.AddIf("heads", 2 < 1, "One is the loneliest number.")
	v2.AddIf("heads", false, "I win, tails you lose.")
	err = v2.Err()
	fields = resperr.ValidationErrors(err)
	be.
		True(v2.Valid()).
		NilError(err).
		Falsey(fields)

	// Don't allocate for valid messages
	allocs := testing.AllocsPerRun(10, func() {
		var v resperr.Validator
		v.AddIf("field", false, "message: %d", 1)
		err = v.Err()
	})
	be.Equal(allocs, 0)
}
