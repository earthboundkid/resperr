package resperr_test

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"

	"github.com/carlmjohnson/resperr"
)

func Example() {
	ts := httptest.NewServer(http.HandlerFunc(myHandler))
	defer ts.Close()

	printResponse(ts.URL, "?")
	printResponse(ts.URL, "?user=admin")
	printResponse(ts.URL, "?user=admin&n=x")
	printResponse(ts.URL, "?user=admin&n=1")
	printResponse(ts.URL, "?user=admin&n=2")
	printResponse(ts.URL, "?user=admin&n=3")
	// Output:
	// logged   ?: [403] bad user ""
	// response ?: {"status":403,"message":"Forbidden"}
	// logged   ?user=admin: [400] missing ?n= in query
	// response ?user=admin: {"status":400,"message":"Please enter a number."}
	// logged   ?user=admin&n=x: [400] strconv.Atoi: parsing "x": invalid syntax
	// response ?user=admin&n=x: {"status":400,"message":"Input is not a number."}
	// logged   ?user=admin&n=1: [404] 1 not found
	// response ?user=admin&n=1: {"status":404,"message":"Not Found"}
	// logged   ?user=admin&n=2: could not connect to database (X_X)
	// response ?user=admin&n=2: {"status":500,"message":"Internal Server Error"}
	// response ?user=admin&n=3: {"data":"data 3"}
}

func replyError(w http.ResponseWriter, r *http.Request, err error) {
	logError(w, r, err)
	code := resperr.StatusCode(err)
	msg := resperr.UserMessage(err)
	replyJSON(w, r, code, struct {
		Status  int    `json:"status"`
		Message string `json:"message"`
	}{
		code,
		msg,
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
	ns := r.URL.Query().Get("n")
	if ns == "" {
		return 0, resperr.WithUserMessage(
			resperr.New(
				http.StatusBadRequest,
				"missing ?n= in query"),
			"Please enter a number.")
	}
	n, err := strconv.Atoi(ns)
	if err != nil {
		return 0, resperr.WithCodeAndMessage(
			err, http.StatusBadRequest,
			"Input is not a number.")
	}
	return n, nil
}

func hasPermissions(r *http.Request) error {
	user := r.URL.Query().Get("user")
	if user == "admin" {
		return nil
	}
	return resperr.New(http.StatusForbidden,
		"bad user %q", user)
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

type Item struct {
	Data string `json:"data"`
}

func logError(w http.ResponseWriter, r *http.Request, err error) {
	fmt.Printf("logged   ?%s: %v\n", r.URL.RawQuery, err)
}

func replyJSON(w http.ResponseWriter, r *http.Request, statusCode int, data interface{}) {
	b, err := json.Marshal(data)
	if err != nil {
		logError(w, r, err)
		w.WriteHeader(http.StatusInternalServerError)
		// Don't use replyJSON to write the error, due to possible loop
		w.Write([]byte(`{"status": 500, "message": "Internal server error"}`))
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
	b, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf("response %s: %s\n", u, b)
}
