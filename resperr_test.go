package resperr_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/carlmjohnson/resperr"
)

func TestGetCode(t *testing.T) {
	base := resperr.WithStatusCode(errors.New(""), 5)
	wrapped := fmt.Errorf("wrapping: %w", base)

	testCases := map[string]struct {
		error
		int
	}{
		"nil":     {nil, 200},
		"default": {errors.New(""), 500},
		"set":     {resperr.WithStatusCode(errors.New(""), 3), 3},
		"set-nil": {resperr.WithStatusCode(nil, 4), 4},
		"wrapped": {wrapped, 5},
		"context": {context.DeadlineExceeded, 504},
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
		coder := resperr.WithStatusCode(err, 400)
		got := coder.Error()
		want := "[400] " + err.Error()
		be.Equal(t, want, got)
	})
	t.Run("keep-chain", func(t *testing.T) {
		err := errors.New("hello")
		coder := resperr.WithStatusCode(err, 3)
		be.True(t, errors.Is(coder, err))
	})
	t.Run("set-nil", func(t *testing.T) {
		coder := resperr.WithStatusCode(nil, 400)
		be.In(t, http.StatusText(400), coder.Error())
	})
	t.Run("override-default", func(t *testing.T) {
		err := context.DeadlineExceeded
		coder := resperr.WithStatusCode(err, 3)
		code := resperr.StatusCode(coder)
		be.Equal(t, 3, code)
	})
}

func TestGetMsg(t *testing.T) {
	base := resperr.WithUserMessage(errors.New(""), "5")
	wrapped := fmt.Errorf("wrapping: %w", base)

	testCases := map[string]struct {
		error
		string
	}{
		"nil":     {nil, ""},
		"default": {errors.New(""), "Internal Server Error"},
		"set":     {resperr.WithUserMessage(errors.New(""), "3"), "3"},
		"set-nil": {resperr.WithUserMessage(nil, "4"), "4"},
		"wrapped": {wrapped, "5"},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			be.Equal(t, tc.string, resperr.UserMessage(tc.error))
		})
	}
}

func TestSetMsg(t *testing.T) {
	t.Run("same-message", func(t *testing.T) {
		err := errors.New("hello")
		msgr := resperr.WithUserMessage(err, "a")
		be.Equal(t, err.Error(), msgr.Error())
	})
	t.Run("keep-chain", func(t *testing.T) {
		err := errors.New("hello")
		msgr := resperr.WithUserMessage(err, "a")
		be.True(t, errors.Is(msgr, err))
	})
	t.Run("set-nil", func(t *testing.T) {
		msgr := resperr.WithUserMessage(nil, "a")
		be.Equal(t, "a", msgr.Error())
	})
}

func TestMsgf(t *testing.T) {
	msg := "hello 1, 2, 3"
	err := resperr.WithUserMessagef(nil, "hello %d, %d, %d", 1, 2, 3)
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
		inner := resperr.WithUserMessage(nil, setMsg)
		w1 := resperr.New(5, "w1: %w", inner)
		w2 := resperr.New(6, "w2: %w", w1)
		be.Equal(t, setMsg, resperr.UserMessage(w2))
		be.Equal(t, 5, resperr.StatusCode(w1))
		be.Equal(t, 6, resperr.StatusCode(w2))
		be.Equal(t, "[6] w2: [5] w1: msg1", w2.Error())
	})
}
