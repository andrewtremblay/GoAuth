package welcome

import (
	"authsys/middlewares/session"
	"bytes"
	//"fmt"
	"authsys/tools/flash"
	"github.com/codegangsta/negroni"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func createHttpServerTest(handler http.Handler) *httptest.Server {
	return httptest.NewServer(handler)
}

func createMiddlewareHandler(middleware negroni.Handler, handler http.Handler) *negroni.Negroni {

	n := negroni.Classic()
	n.Use(middleware)
	n.UseHandler(handler)
	return n

}

func TestRequest(t *testing.T) {

	handler := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		c := new(controller)
		c.Request = r
		c.Response = rw
		c.Title = "Welcome"
		c.serve()
	})

	ts := createHttpServerTest(createMiddlewareHandler(session.New(), handler))
	defer ts.Close()

	client := &http.Client{}
	response, err := client.Get(ts.URL)
	assert.NoError(t, err, "Should not contain any error.")

	buffer := new(bytes.Buffer)
	io.Copy(buffer, response.Body)
	assert.Contains(t, buffer.String(), "Sign In", "Should contain sign in link.")

}

func TestFlashMessage(t *testing.T) {

	handler := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		c := new(controller)
		c.Request = r
		c.Response = rw
		c.Title = "Welcome"
		flash.Set(c.Request, "I", "This is the flash message")
		c.serve()
	})

	ts := createHttpServerTest(createMiddlewareHandler(session.New(), handler))
	defer ts.Close()

	client := &http.Client{}
	response, err := client.Get(ts.URL)
	assert.NoError(t, err, "Should not contain any error.")

	buffer := new(bytes.Buffer)
	io.Copy(buffer, response.Body)
	assert.Contains(t, buffer.String(), "This is the flash message", "Flash message.")

}

func TestNotValidFlashOption(t *testing.T) {

	handler := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		c := new(controller)
		c.Request = r
		c.Response = rw
		c.Title = "Welcome"
		assert.Panics(t, func() {
			flash.Set(c.Request, "T", "This option is not allowed.")
		}, "Should panic because wrong option.")
		c.serve()
	})

	ts := createHttpServerTest(createMiddlewareHandler(session.New(), handler))
	defer ts.Close()
}
