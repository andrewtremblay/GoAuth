package flash

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func createHttpServerTest(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

func TestFlashMessage(t *testing.T) {
	ts := createHttpServerTest(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		Set(r, "E", "This this an error")
		msg1 := Get(r, "E")
		assert.Equal(t, "This this an error", msg1, "Should contain the message 'This this an error'")

		msg2 := Get(r, "S")
		assert.Empty(t, msg2, "Shoud not find message.")

	}))

	defer ts.Close()
}

func TestValidationMessageOption(t *testing.T) {
	ts := createHttpServerTest(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		err := Set(r, "T", "This this an error")
		assert.Error(t, err, "Should contain error, because option does not allowed.")

	}))

	defer ts.Close()
}
