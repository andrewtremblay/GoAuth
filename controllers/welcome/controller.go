package welcome

import (
	"authsys/controllers/base"
	"authsys/tools/caller"
	//"fmt"
	//"html/template"
	"errors"
	"net/http"
	"path"
)

var (
	welcomeTmpl string
)

func init() {
	welcomeTmpl = path.Join(caller.Path(), "welcome.html")
}

type controller struct {
	*base.Controller
}

// Read infos from flash messaging
func (rcv *controller) getInfos() {

	if msg := rcv.GetFlash("I"); msg != "" {
		rcv.AppendInfo(msg)
	}
}

// Read errors from flash messaging
func (rcv *controller) getErrors() {
	if msg := rcv.GetFlash("E"); msg != "" {
		rcv.AppendError(errors.New(msg))
	}
}

func (rcv *controller) get() error {
	rcv.RenderContentPart(welcomeTmpl, nil)
	return nil
}

func (rcv *controller) serve() {
	rcv.SetTitle("text09")
	rcv.getErrors()
	rcv.getInfos()
	rcv.get()
	rcv.Render()
}

func New() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		c := &controller{base.New(rw, r, true)}
		c.serve()
	})
}
