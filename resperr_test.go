package resperr_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/earthboundkid/resperr/v2"
)

func TestGetCode(t *testing.T) {
	base := resperr.E{S: 5}
	wrapped := fmt.Errorf("wrapping: %w", base)

	testCases := map[string]struct {
		error
		int
	}{
		"nil":         {nil, 200},
		"default":     {errors.New(""), 500},
		"set":         {resperr.E{E: errors.New(""), S: 3}, 3},
		"set-nil":     {resperr.E{E: nil, S: 4}, 4},
		"wrapped":     {wrapped, 5},
		"set-message": {resperr.E{M: "xxx"}, 400},
		"set-both":    {resperr.E{S: 6, M: "xx"}, 6},
		"context":     {context.DeadlineExceeded, 504},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			be.Equal(t, tc.int, resperr.StatusCode(tc.error))
		})
	}
}

func TestSetCode(t *testing.T) {
	t.Run("same-message", func(t *testing.T) {
		err := errors.New("hello")
		coder := resperr.E{E: err, S: 400}
		got := coder.Error()
		want := "[400] " + err.Error()
		be.Equal(t, want, got)
	})
	t.Run("keep-chain", func(t *testing.T) {
		err := errors.New("hello")
		coder := resperr.E{E: err, S: 3}
		be.True(t, errors.Is(coder, err))
	})
	t.Run("set-nil", func(t *testing.T) {
		coder := resperr.E{S: 400}
		be.In(t, http.StatusText(400), coder.Error())
	})
	t.Run("override-default", func(t *testing.T) {
		err := context.DeadlineExceeded
		coder := resperr.E{E: err, S: 3}
		code := resperr.StatusCode(coder)
		be.Equal(t, 3, code)
	})
}

func TestGetMsg(t *testing.T) {
	base := resperr.E{E: errors.New(""), M: "5"}
	wrapped := fmt.Errorf("wrapping: %w", base)

	testCases := map[string]struct {
		error
		string
	}{
		"nil":     {nil, ""},
		"default": {errors.New(""), "Internal Server Error"},
		"set":     {resperr.E{E: errors.New(""), M: "3"}, "3"},
		"set-nil": {resperr.E{M: "4"}, "4"},
		"wrapped": {wrapped, "5"},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			be.Equal(t, tc.string, resperr.UserMessage(tc.error))
		})
	}
}

func TestSetMsg(t *testing.T) {
	t.Run("has-cause", func(t *testing.T) {
		err := errors.New("hello")
		msgr := resperr.E{E: err, M: "a"}
		be.In(t, err.Error(), msgr.Error())
	})
	t.Run("keep-chain", func(t *testing.T) {
		err := errors.New("hello")
		msgr := resperr.E{E: err, M: "a"}
		be.True(t, errors.Is(msgr, err))
	})
	t.Run("has-message", func(t *testing.T) {
		msgr := resperr.E{M: "abc"}
		be.In(t, "abc", msgr.Error())
	})
}

func TestMsgf(t *testing.T) {
	msg := "hello 1, 2, 3"
	err := resperr.E{M: fmt.Sprintf("hello %d, %d, %d", 1, 2, 3)}
	be.Equal(t, msg, resperr.UserMessage(err))
}

func TestNotFound(t *testing.T) {
	path := "/example/url"
	r, _ := http.NewRequest(http.MethodGet, path, nil)
	err := resperr.NotFound(r)
	be.In(t, path, err.Error())
	be.In(t, path, resperr.UserMessage(err))
	be.Equal(t, 404, resperr.StatusCode(err))
}

func TestNew(t *testing.T) {
	t.Run("flat", func(t *testing.T) {
		err := resperr.New(404, "hello %s", "world")
		be.Equal(t, "Not Found", resperr.UserMessage(err))
		be.Equal(t, 404, resperr.StatusCode(err))
		be.Equal(t, "[404] hello world", err.Error())
	})
	t.Run("chain", func(t *testing.T) {
		const setMsg = "msg1"
		inner := resperr.E{M: setMsg}
		w1 := resperr.New(5, "w1: %w", inner)
		w2 := resperr.New(6, "w2: %w", w1)
		be.Equal(t, setMsg, resperr.UserMessage(w2))
		be.Equal(t, 5, resperr.StatusCode(w1))
		be.Equal(t, 6, resperr.StatusCode(w2))
		be.Equal(t, "[6] w2: [5] w1: [400] <msg1> Bad Request", w2.Error())
	})
}
