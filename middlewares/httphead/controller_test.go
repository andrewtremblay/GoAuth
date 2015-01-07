package httphead

import (
	// "fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	//"strings"
	"testing"
)

func createHttpServerTest(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

func TestLanguage(t *testing.T) {

	var err error

	ts := createHttpServerTest(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		c := &controller{request: r}
		c.setLang()

		lang := GetLang(r)
		assert.Contains(t, lang, "en-US", "Should contains language")

	}))
	defer ts.Close()

	client := &http.Client{}
	req := &http.Request{}
	req.Method = "GET"
	req.Header = make(http.Header)
	req.Header.Set("Accept-Language", "en-US")
	req.URL, err = url.Parse(ts.URL)
	assert.NoError(t, err, "Should not contain any error.")
	client.Do(req)

}
