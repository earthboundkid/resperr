# resperr [![GoDoc](https://godoc.org/github.com/earthboundkid/resperr?status.svg)](https://godoc.org/github.com/earthboundkid/resperr/v2) [![Go Report Card](https://goreportcard.com/badge/github.com/earthboundkid/resperr/v2)](https://goreportcard.com/report/github.com/earthboundkid/resperr/v2) [![Calver v2.YY.Minor](https://img.shields.io/badge/calver-v2.YY.Minor-22bfda.svg)](https://calver.org)

Resperr is a Go package to associate status codes and messages with errors.

## Example usage

See [blog post](https://blog.carlana.net/post/2020/working-with-errors-as/) for a full description or [read the test code](https://github.com/earthboundkid/resperr/blob/master/example_test.go) for more context:

```go
// write a simple handler that just checks for errors
// and replies with an error object if it gets one

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

// in the functions that your handler calls
// use resp err to associate different error conditions
// with appropriate HTTP status codes

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

// you can also return specific messages for users as needed

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
```
