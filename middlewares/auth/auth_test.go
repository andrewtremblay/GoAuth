package auth

import (
	"authsys/middlewares/httphead"
	"authsys/middlewares/session"
	"authsys/models/account"
	"bytes"
	"fmt"
	"github.com/codegangsta/negroni"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testWriter() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		u := &account.SignedUser{Name: "Foo", Email: "foo@example.com"}
		if err := Signed(rw, u, "en-US"); err != nil {
			http.Error(rw, err.Error(), 501)
		}
	})
}

func testReader() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(rw, "You are signed")
	})
}

func TestSetCookie(t *testing.T) {
	ts := httptest.NewServer(testWriter())
	defer ts.Close()
	response, err := http.Get(ts.URL)
	assert.NoError(t, err, "Should not contain any error")

	cookie := response.Cookies()[0]
	assert.Contains(t, cookie.Name, "VERIFIED", "Should contain verified cookie. That's mean successfully sign in.")
}

func TestClient(t *testing.T) {
	ts := httptest.NewServer(testWriter())
	defer ts.Close()

	c := &http.Client{}
	response, err := c.Get(ts.URL)
	assert.NoError(t, err, "Should not contain any error")

	cookie := response.Cookies()[0]
	assert.Contains(t, cookie.Name, "VERIFIED", "Should contain verified cookie. That's mean successfully sign in.")

}

func TestAuthorizationReader(t *testing.T) {

	tw := httptest.NewServer(testWriter())
	n := negroni.New(httphead.New(), session.New(), authorize())
	n.UseHandler(testReader())
	tr := httptest.NewServer(n)
	defer tw.Close()
	defer tr.Close()

	c := &http.Client{}
	rs, err := c.Get(tw.URL)
	assert.NoError(t, err, "Should not contain any error")

	rq, err := http.NewRequest("GET", tr.URL, nil)
	assert.NoError(t, err, "Should not contain any error")
	rq.AddCookie(rs.Cookies()[0])
	rs.Header.Add("Accept-Language", "en-US")
	rs, err = c.Do(rq)

	buffer := new(bytes.Buffer)
	io.Copy(buffer, rs.Body)
	assert.Contains(t, buffer.String(), "You are signed", "User should be signed.")

}

func TestFailAuthorization(t *testing.T) {

	n := negroni.New(httphead.New(), session.New(), authorize())
	n.UseHandler(testReader())
	tr := httptest.NewServer(n)
	defer tr.Close()

	redirectPolicyFunc := func(req *http.Request, via []*http.Request) error {
		return nil
	}

	rq, err := http.NewRequest("GET", tr.URL, nil)
	rq.Header.Add("Accept-Language", "en-US")
	c := &http.Client{CheckRedirect: redirectPolicyFunc}
	_, err = c.Do(rq)
	//fmt.Println(rs.Location())
	assert.NoError(t, err, "Should not contain any error")

}
