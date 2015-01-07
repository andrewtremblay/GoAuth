package context

import (
	gontext "github.com/gorilla/context"
	"net/http"
)

type key int

// Define the identification keys for incoming requests
const (
	LANGUAGE key = iota
	SID
	SIGNEDID
)

// Will return the value of the key
func Get(r *http.Request, key interface{}) interface{} {
	return gontext.Get(r, key)
}

// Parameter key is the identification and data holder during
// a request.
func Set(r *http.Request, key, val interface{}) {
	gontext.Set(r, key, val)
}
