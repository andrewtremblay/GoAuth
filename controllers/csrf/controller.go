package csrf

import (
	"authsys/controllers/base"
	"authsys/tools/caller"
	//"fmt"
	"github.com/justinas/nosurf"
	"net/http"
	"path"
)

var (
	vioTmpl string
)

func init() {
	vioTmpl = path.Join(caller.Path(), "violate.html")
}

func New(handler http.Handler) http.Handler {
	n := nosurf.New(handler)
	n.SetFailureHandler(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		b := base.New(rw, r, false)
		b.SetTitle("text11")
		b.RenderContentPart(vioTmpl, nil)
		b.Render()
	}))

	return n
}

func Token(req *http.Request) string {
	return nosurf.Token(req)
}
