package security

import (
	"bytes"
	"fmt"
	"github.com/codegangsta/negroni"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func createHttpServerTest() *httptest.Server {

	n := negroni.Classic()
	n.Use(New())
	return httptest.NewServer(n)
}

func TestSecurity(t *testing.T) {

	client := &http.Client{}

	ts := createHttpServerTest()
	defer ts.Close()

	response, err := client.Get(ts.URL)
	assert.NoError(t, err, "Should not contain any error.")

	// buffer := new(bytes.Buffer)
	// io.Copy(buffer, response.Body)
	// fmt.Println(buffer.String())
	// assert.Contains(t, buffer.String(), "Security rejecting", "Should rejected request.")

}
