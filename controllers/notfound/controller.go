package notfound

import (
	"authsys/controllers/base"
	"authsys/tools/caller"
	"net/http"
	"path"
)

var (
	errorTmpl string
)

func init() {
	errorTmpl = path.Join(caller.Path(), "404.html")
}

type controller struct {
	*base.Controller
}

func (rcv *controller) serve() {
	rcv.SetTitle("text10")
	rcv.RenderContentPart(errorTmpl, nil)
	rcv.Render()
}

func Serve(bc *base.Controller) {
	c := &controller{bc}
	c.serve()
}

func New() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		c := &controller{base.New(rw, r, false)}
		c.serve()
	})
}
