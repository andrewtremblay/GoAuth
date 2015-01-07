package captcha

import (
	//"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	//"strings"
	"bytes"
	"io"
	"testing"
)

func createHttpServerTest(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

func TestCreateImage(t *testing.T) {

	client := &http.Client{}
	ts := createHttpServerTest(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		c := &controller{response: rw, request: r}
		image, _ := Create()
		err := c.createNewImage(image)
		assert.NoError(t, err, "Should not contain any error.")
	}))

	defer ts.Close()
	response, _ := client.Get(ts.URL)

	buffer := new(bytes.Buffer)
	io.Copy(buffer, response.Body)
	assert.NotEmpty(t, buffer.String(), "Should contain picture.")

}
