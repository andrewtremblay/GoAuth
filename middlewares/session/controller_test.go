package session

import (
	// "fmt"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	//"strings"
	"testing"
)

func createHttpServerTest(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

func TestCreateAndReadJwt(t *testing.T) {

	c := &controller{}
	token, err := c.createJwt()
	assert.NoError(t, err, "Should not contain any error.")
	assert.NotEmpty(t, token, "Should contain jwt token.")

	sid, err := c.readJwt(token)
	assert.NoError(t, err, "Should not contain any error.")
	assert.NotEmpty(t, sid, "Should contain session token.")

}

func TestCreateAndVerifyCookie(t *testing.T) {

	Convey("Testing create and verify cookie.", t, func() {

		client := &http.Client{}

		Convey("Create cookie for client.", func() {

			ts := createHttpServerTest(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				c := &controller{response: rw, request: r}
				c.createCookie()
			}))

			defer ts.Close()
			response, err := client.Get(ts.URL)
			assert.NoError(t, err, "Should not contain any error.")
			assert.Len(t, response.Cookies(), 1, "Should contain storage cookie.")
		})

		Convey("Verify cookie on client.", func() {

			ts := createHttpServerTest(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				c := &controller{response: rw, request: r}
				err := c.readCookie()
				assert.NoError(t, err, "Should not contain any error.")
			}))

			defer ts.Close()
			response, err := client.Get(ts.URL)
			assert.NoError(t, err, "Should not contain any error.")
			assert.Len(t, response.Cookies(), 1, "Should contain storage cookie.")
		})
	})

}

func TestLiveTimeOfSid(t *testing.T) {
	c := &controller{}

	err := c.renewTime("sid")
	assert.NoError(t, err, "Should not contains any error.")
}

func TestStorage(t *testing.T) {

	ts := createHttpServerTest(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {

		InsertData(r, "KEY", "VALUE")

		data, err := ReadData(r, "KEY")
		assert.Contains(t, data, "VALUE", "Should contains value.")
		assert.NoError(t, err, "Should not contains any error.")

	}))

	defer ts.Close()

}
