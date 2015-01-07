package security

import (
	"fmt"
	"github.com/codegangsta/negroni"
	"gopkg.in/unrolled/secure.v1"
	"net/http"
)

type controller struct {
	response http.ResponseWriter
	request  *http.Request
	security *secure.Secure
}

// Customizing security options
func (rcv *controller) setOptions() {
	rcv.security = secure.New(secure.Options{
		AllowedHosts:          []string{"example.com", "ssl.example.com"},
		SSLRedirect:           true,
		SSLHost:               "ssl.example.com",
		SSLProxyHeaders:       map[string]string{"X-Forwarded-Proto": "https"},
		STSSeconds:            315360000,
		STSIncludeSubdomains:  true,
		FrameDeny:             true,
		ContentTypeNosniff:    true,
		BrowserXssFilter:      true,
		ContentSecurityPolicy: "default-src 'self'",
		IsDevelopment:         true,
	})
}

// If security requirements are not pass
func (rcv *controller) setRejector() {
	rcv.security.SetBadHostHandler(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		http.Error(rw, "Security rejecting", http.StatusBadRequest)
	}))
}

func (rcv *controller) handle(next http.HandlerFunc) {

	rcv.setOptions()
	rcv.setRejector()
	rcv.security.HandlerFuncWithNext(rcv.response, rcv.request, next)
}

// Middlware
func New() negroni.HandlerFunc {
	fmt.Println("Middleware security is started.")
	return negroni.HandlerFunc(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {

		c := &controller{request: r, response: rw}
		c.handle(next)
	})
}
