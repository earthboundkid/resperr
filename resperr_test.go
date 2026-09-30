package resperr_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/earthboundkid/assert"
	"github.com/earthboundkid/resperr/v2"
)

func TestGetCode(t *testing.T) {
	base := resperr.E{S: 5}
	wrapped := fmt.Errorf("wrapping: %w", base)

	type testcase struct {
		error
		int
	}
	assert.RunAll(t, map[string]testcase{
		"nil":         {nil, 200},
		"default":     {errors.New(""), 500},
		"set":         {resperr.E{E: errors.New(""), S: 3}, 3},
		"set-nil":     {resperr.E{E: nil, S: 4}, 4},
		"wrapped":     {wrapped, 5},
		"set-message": {resperr.E{M: "xxx"}, 400},
		"set-both":    {resperr.E{S: 6, M: "xx"}, 6},
		"context":     {context.DeadlineExceeded, 504},
		"e":           {resperr.E{}, 500},
		"m":           {resperr.M("hi"), 400},
		"join-code-message": {errors.Join(
			resperr.New(404, "hi"),
			resperr.M("hello"),
		), 404},
		"join-message-code": {errors.Join(
			resperr.M("hello"),
			resperr.New(404, "hi"),
		), 404},
		"join-code-ii": {errors.Join(
			resperr.E{M: "yo"},
			resperr.New(401, "hi"),
		), 401},
	}, func(be assert.TB, tc testcase) {
		be.Equal(resperr.StatusCode(tc.error), tc.int)
	})
}

func TestSetCode(t *testing.T) {
	assert.FailsNow(t).
		Run("same-message", func(be assert.TB) {
			err := errors.New("hello")
			coder := resperr.E{E: err, S: 400}
			got := coder.Error()
			be.Equal(got, "[400] "+err.Error())
		}).
		Run("keep-chain", func(be assert.TB) {
			err := errors.New("hello")
			coder := resperr.E{E: err, S: 3}
			be.ErrorIs(coder, err)
		}).
		Run("set-nil", func(be assert.TB) {
			coder := resperr.E{S: 400}
			be.Match(coder.Error(), http.StatusText(400))
		}).
		Run("override-default", func(be assert.TB) {
			err := context.DeadlineExceeded
			coder := resperr.E{E: err, S: 3}
			code := resperr.StatusCode(coder)
			be.Equal(code, 3)
		})
}

func TestGetMsg(t *testing.T) {
	base := resperr.E{E: errors.New(""), M: "5"}
	wrapped := fmt.Errorf("wrapping: %w", base)
	type testcase struct {
		error
		string
	}
	assert.RunAll(t, map[string]testcase{
		"nil":     {nil, ""},
		"default": {errors.New(""), "Internal Server Error"},
		"set":     {resperr.E{E: errors.New(""), M: "3"}, "3"},
		"set-nil": {resperr.E{M: "4"}, "4"},
		"wrapped": {wrapped, "5"},
		"m":       {resperr.M("hi"), "hi"},
		"join-code-message": {errors.Join(
			resperr.New(404, "hi"),
			resperr.M("hello"),
		), "hello"},
		"join-message-code": {errors.Join(
			resperr.M("howdy"),
			resperr.New(404, "hi"),
		), "howdy"},
		"join-code-ii": {errors.Join(
			resperr.New(404, "hi"),
			resperr.E{S: 401},
			resperr.M("howdy"),
		), "howdy"},
	}, func(be assert.TB, tc testcase) {
		be.Equal(resperr.UserMessage(tc.error), tc.string)
	})
}

func TestSetMsg(t *testing.T) {
	assert.FailsNow(t).
		Run("has-cause", func(be assert.TB) {
			err := errors.New("hello")
			msgr := resperr.E{E: err, M: "a"}
			be.Match(msgr.Error(), err.Error())
		}).
		Run("keep-chain", func(be assert.TB) {
			err := errors.New("hello")
			msgr := resperr.E{E: err, M: "a"}
			be.ErrorIs(msgr, err)
		}).
		Run("has-message", func(be assert.TB) {
			msgr := resperr.E{M: "abc"}
			be.Match(msgr.Error(), `abc`)
		})
}

func TestMsgf(t *testing.T) {
	msg := "hello 1, 2, 3"
	err := resperr.E{M: fmt.Sprintf("hello %d, %d, %d", 1, 2, 3)}
	assert.FailsNow(t).Equal(msg, resperr.UserMessage(err))
}

func TestNotFound(t *testing.T) {
	path := "/example/url"
	r, _ := http.NewRequest(http.MethodGet, path, nil)
	err := resperr.NotFound(r)
	assert.FailsNow(t).
		Match(err.Error(), path).
		Match(resperr.UserMessage(err), path).
		Equal(resperr.StatusCode(err), 404)
}

func TestNew(t *testing.T) {
	assert.FailsNow(t).
		Run("flat", func(be assert.TB) {
			err := resperr.New(404, "hello %s", "world")
			be.Equal(resperr.UserMessage(err), "Not Found")
			be.Equal(resperr.StatusCode(err), 404)
			be.Equal(err.Error(), "[404] hello world")
		}).
		Run("chain", func(be assert.TB) {
			const setMsg = "msg1"
			inner := resperr.E{M: setMsg}
			w1 := resperr.New(5, "w1: %w", inner)
			w2 := resperr.New(6, "w2: %w", w1)
			be.Equal(setMsg, resperr.UserMessage(w2))
			be.Equal(resperr.StatusCode(w1), 5)
			be.Equal(resperr.StatusCode(w2), 6)
			be.Equal(w2.Error(), "[6] w2: [5] w1: [400] <msg1> Bad Request")
		})
}
