package resperr_test

import (
	"fmt"
	"net/http"
	"strconv"
	"testing"

	"github.com/carlmjohnson/be"
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
	var v1 resperr.Validator
	v1.AddIf("heads", 2 > 1, "Two are better than one.")
	v1.AddIf("heads", true, "I win, tails you lose.")
	err := v1.Err()
	be.Nonzero(t, err)
	be.False(t, v1.Valid())
	fields := resperr.ValidationErrors(err)
	be.Equal(t, 1, len(fields))
	be.Equal(t, 2, len(fields["heads"]))
	be.Equal(t, http.StatusBadRequest, resperr.StatusCode(err))

	var v2 resperr.Validator
	v2.AddIf("heads", 2 < 1, "One is the loneliest number.")
	v2.AddIf("heads", false, "I win, tails you lose.")
	err = v2.Err()
	be.True(t, v2.Valid())
	be.NilErr(t, err)
	fields = resperr.ValidationErrors(err)
	be.Zero(t, fields)

	// Don't allocate for valid messages
	allocs := testing.AllocsPerRun(10, func() {
		var v resperr.Validator
		v.AddIf("field", false, "message: %d", 1)
		err = v.Err()
	})
	be.Equal(t, 0, allocs)
}
