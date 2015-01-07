package auth

import (
	//"fmt"
	"net/http"
)

// Programmatically sign out
func SignOut(rw http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie(cookie)
	if err != nil {
		http.Redirect(rw, r, "/", 303)
		return
	}

	// Delete cookie
	c.MaxAge = -100
	c.Path = "/"
	http.SetCookie(rw, c)
	http.Redirect(rw, r, "/", 303)
}

// Manually sign out. This function will be wrap
// through another function, that check, if user
// is signed in to allow sign out.
func SelfSignOut() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		SignOut(rw, r)
	})
}
