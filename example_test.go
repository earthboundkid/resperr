package resperr_test

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"

	"github.com/earthboundkid/resperr/v2"
)

func Example() {
	ts := httptest.NewServer(http.HandlerFunc(myHandler))
	defer ts.Close()

	printResponse(ts.URL, "?")
	// logs: [403] bad user ""
	// response: {"status":403,"error":"Forbidden"}
	printResponse(ts.URL, "?user=admin")
	// logs: validation error: n=Please enter a number.
	// response: {"status":400,"error":"Bad Request","details":{"n":["Please enter a number."]}}
	printResponse(ts.URL, "?user=admin&n=x")
	// logs: validation error: n=Input is not a number.
	// response: {"status":400,"error":"Bad Request","details":{"n":["Input is not a number."]}}
	printResponse(ts.URL, "?user=admin&n=1")
	// logs: [404] 1 not found
	// response: {"status":404,"error":"Not Found"}
	printResponse(ts.URL, "?user=admin&n=2")
	// logs: could not connect to database (X_X)
	// response: {"status":500,"error":"Internal Server Error"}
	printResponse(ts.URL, "?user=admin&n=3")
	// response: {"data":"data 3"}

	// Output:
	// logged   ?: [403] bad user ""
	// response ?: {"status":403,"error":"Forbidden"}
	// logged   ?user=admin: validation error: n=Please enter a number.
	// response ?user=admin: {"status":400,"error":"Bad Request","details":{"n":["Please enter a number."]}}
	// logged   ?user=admin&n=x: validation error: n=Input is not a number.
	// response ?user=admin&n=x: {"status":400,"error":"Bad Request","details":{"n":["Input is not a number."]}}
	// logged   ?user=admin&n=1: [404] 1 not found
	// response ?user=admin&n=1: {"status":404,"error":"Not Found"}
	// logged   ?user=admin&n=2: could not connect to database (X_X)
	// response ?user=admin&n=2: {"status":500,"error":"Internal Server Error"}
	// response ?user=admin&n=3: {"data":"data 3"}
}

func replyError(w http.ResponseWriter, r *http.Request, err error) {
	logError(w, r, err)
	code := resperr.StatusCode(err)
	msg := resperr.UserMessage(err)
	details := resperr.ValidationErrors(err)
	replyJSON(w, r, code, struct {
		Status  int        `json:"status"`
		Error   string     `json:"error,omitzero"`
		Details url.Values `json:"details,omitzero"`
	}{
		code,
		msg,
		details,
	})
}

func myHandler(w http.ResponseWriter, r *http.Request) {
	// ... check user permissions...
	if err := hasPermissions(r); err != nil {
		replyError(w, r, err)
		return
	}
	// ...validate request...
	n, err := getItemNoFromRequest(r)
	if err != nil {
		replyError(w, r, err)
		return
	}
	// ...get the data ...
	item, err := getItemByNumber(n)
	if err != nil {
		replyError(w, r, err)
		return
	}
	replyJSON(w, r, http.StatusOK, item)
}

func getItemByNumber(n int) (item *Item, err error) {
	item, err = dbCall("...", n)
	if err == sql.ErrNoRows {
		// this is an anticipated 404
		return nil, resperr.New(
			http.StatusNotFound,
			"%d not found", n)
	}
	if err != nil {
		// this is an unexpected 500!
		return nil, err
	}
	// ...
	return
}

func getItemNoFromRequest(r *http.Request) (int, error) {
	var v resperr.Validator
	ns := r.URL.Query().Get("n")
	v.AddIf("n", ns == "", "Please enter a number.")
	n, err := strconv.Atoi(ns)
	v.AddIfUnset("n", err != nil, "Input is not a number.")
	return n, v.Err()
}

func hasPermissions(r *http.Request) error {
	// lol, don't do this!
	user := r.URL.Query().Get("user")
	if user == "admin" {
		return nil
	}
	return resperr.New(http.StatusForbidden,
		"bad user %q", user)
}

// boilerplate below:

type Item struct {
	Data string `json:"data"`
}

func dbCall(s string, i int) (*Item, error) {
	if i == 1 {
		return nil, sql.ErrNoRows
	}
	if i == 2 {
		return nil, fmt.Errorf("could not connect to database (X_X)")
	}
	return &Item{fmt.Sprintf("data %d", i)}, nil
}

func logError(w http.ResponseWriter, r *http.Request, err error) {
	fmt.Printf("logged   ?%s: %v\n", r.URL.RawQuery, err)
}

func replyJSON(w http.ResponseWriter, r *http.Request, statusCode int, data any) {
	b, err := json.Marshal(data)
	if err != nil {
		logError(w, r, err)
		w.WriteHeader(http.StatusInternalServerError)
		// Don't use replyJSON to write the error, due to possible loop
		w.Write([]byte(`{"status": 500, "error": "Internal server error"}`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, err = w.Write(b)
	if err != nil {
		logError(w, r, err)
	}
}

func printResponse(base, u string) {
	resp, err := http.Get(base + u)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	fmt.Printf("response %s: %s\n", u, b)
}
