package httphead

/*
 * Read header information from http request and
 * save into gorilla context during a request.
 */

import (
	"authsys/tools/context"
	"fmt"
	"github.com/codegangsta/negroni"
	"net/http"
	"strings"
)

func GetLang(r *http.Request) string {
	return context.Get(r, context.LANGUAGE).(string)
}

type controller struct {
	request *http.Request
}

func (rcv *controller) setLang() {
	str := strings.Split(rcv.request.Header.Get("Accept-Language"), ",")
	context.Set(rcv.request, context.LANGUAGE, str[0])
}

func (rcv *controller) handle() {
	rcv.setLang()
}

func New() negroni.HandlerFunc {
	fmt.Println("Middleware header is started.")
	return negroni.HandlerFunc(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		c := &controller{request: r}
		c.handle()
		next(rw, r)
	})
}
