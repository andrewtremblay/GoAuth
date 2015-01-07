package view

import (
	"authsys/controllers/base"
	//"authsys/models/account"
	"authsys/middlewares/auth"
	"authsys/tools/caller"
	//"fmt"
	//"html/template"
	//"errors"
	"net/http"
	"path"
)

var (
	viewTmpl string
)

func init() {
	viewTmpl = path.Join(caller.Path(), "view.html")
}

type controller struct {
	*base.Controller
}

func (rcv *controller) get() error {
	signedUser := auth.GetSignedInUser(rcv.Request)
	rcv.SetTitle("text12", signedUser.Name)
	rcv.RenderContentPart(viewTmpl, signedUser)
	return nil
}

func (rcv *controller) serve() {
	rcv.get()
	rcv.Render()
}

func New() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		c := &controller{base.New(rw, r, true, "controller/account")}
		c.serve()
	})
}
