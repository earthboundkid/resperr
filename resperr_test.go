package resperr_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

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
			err := tc.error
			want := tc.int
			got := resperr.StatusCode(err)
			if got != want {
				t.Errorf("%v: %d != %d", err, got, want)
			}
		})
	}
}

func TestSetCode(t *testing.T) {
	t.Run("same-message", func(t *testing.T) {
		err := errors.New("hello")
		coder := resperr.WithStatusCode(err, 400)
		got := coder.Error()
		want := "[400] " + err.Error()
		if got != want {
			t.Errorf("error message %q != %q", got, want)
		}
	})
	t.Run("keep-chain", func(t *testing.T) {
		err := errors.New("hello")
		coder := resperr.WithStatusCode(err, 3)

		if !errors.Is(coder, err) {
			t.Errorf("broken chain: %v is not %v", coder, err)
		}
	})
	t.Run("set-nil", func(t *testing.T) {
		coder := resperr.WithStatusCode(nil, 400)
		if msg := coder.Error(); !strings.Contains(msg, http.StatusText(400)) {
			t.Errorf("message should contain text: %q", msg)
		}
	})
	t.Run("override-default", func(t *testing.T) {
		err := context.DeadlineExceeded
		coder := resperr.WithStatusCode(err, 3)

		if code := resperr.StatusCode(coder); code != 3 {
			t.Errorf("did not override code %d != 3", code)
		}
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
			err := tc.error
			want := tc.string
			got := resperr.UserMessage(err)
			if got != want {
				t.Errorf("%v: %q != %q", err, got, want)
			}
		})
	}
}

func TestSetMsg(t *testing.T) {
	t.Run("same-message", func(t *testing.T) {
		err := errors.New("hello")
		msgr := resperr.WithUserMessage(err, "a")
		got := msgr.Error()
		want := err.Error()
		if got != want {
			t.Errorf("error message %q != %q", got, want)
		}
	})
	t.Run("keep-chain", func(t *testing.T) {
		err := errors.New("hello")
		msgr := resperr.WithUserMessage(err, "a")

		if !errors.Is(msgr, err) {
			t.Errorf("broken chain: %v is not %v", msgr, err)
		}
	})
	t.Run("set-nil", func(t *testing.T) {
		msgr := resperr.WithUserMessage(nil, "a")
		if msg := msgr.Error(); msg != "a" {
			t.Errorf("%q != %q", "a", msg)
		}
	})
}

func TestMsgf(t *testing.T) {
	msg := "hello 1, 2, 3"
	err := resperr.WithUserMessagef(nil, "hello %d, %d, %d", 1, 2, 3)
	if got := resperr.UserMessage(err); msg != got {
		t.Errorf("%q != %q", got, msg)
	}
}

func TestNotFound(t *testing.T) {
	path := "/example/url"
	r, _ := http.NewRequest(http.MethodGet, path, nil)
	err := resperr.NotFound(r)
	if msg := err.Error(); !strings.Contains(msg, path) {
		t.Errorf("error message should contain path: %q", msg)
	}
	if msg := resperr.UserMessage(err); !strings.Contains(msg, path) {
		t.Errorf("user message should contain path: %q", msg)
	}
	if code := resperr.StatusCode(err); code != 404 {
		t.Errorf("wrong code: %d", code)
	}
}

func TestNew(t *testing.T) {
	t.Run("flat", func(t *testing.T) {
		err := resperr.New(404, "hello %s", "world")
		if gotMsg := resperr.UserMessage(err); gotMsg != "Not Found" {
			t.Errorf("user message %q != %q", gotMsg, "")
		}
		if code := resperr.StatusCode(err); code != 404 {
			t.Errorf("wrong code: %d", code)
		}
		if s := err.Error(); s != "[404] hello world" {
			t.Errorf("wrong error string: %q", s)
		}
	})
	t.Run("chain", func(t *testing.T) {
		const setMsg = "msg1"
		inner := resperr.WithUserMessage(nil, setMsg)
		w1 := resperr.New(5, "w1: %w", inner)
		w2 := resperr.New(6, "w2: %w", w1)
		if gotMsg := resperr.UserMessage(w2); gotMsg != setMsg {
			t.Errorf("user message %q != %q", gotMsg, setMsg)
		}
		if code := resperr.StatusCode(w1); code != 5 {
			t.Errorf("wrong code: %d", code)
		}
		if code := resperr.StatusCode(w2); code != 6 {
			t.Errorf("wrong code: %d", code)
		}
		if s := w2.Error(); s != "[6] w2: [5] w1: msg1" {
			t.Errorf("wrong error string: %q", s)
		}
	})
}
