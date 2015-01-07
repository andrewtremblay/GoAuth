package notfound

import (
	"bytes"
	//"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func createHttpServerTest(handler http.Handler) *httptest.Server {
	return httptest.NewServer(handler)
}

func TestNotFoundRequest(t *testing.T) {

	ts := createHttpServerTest(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		New(rw, r)
	}))
	defer ts.Close()

	client := &http.Client{}
	response, err := client.Get(ts.URL)
	assert.NoError(t, err, "Should not contain any error.")

	buffer := new(bytes.Buffer)
	io.Copy(buffer, response.Body)
	assert.Contains(t, buffer.String(), "Page not found", "Should contain text Page not found.")

}
