package delete

import (
	"authsys/controllers/base"
	"authsys/middlewares/auth"
	"authsys/models/account"
	"net/http"
)

type controller struct {
	*base.Controller
	loggedUser *account.SignedUser
}

func (rcv *controller) delete() error {

	if err := account.Delete(rcv.loggedUser.Email, rcv.Local); err != nil {
		return err
	}

	return nil
}

func (rcv *controller) serve() {

	rcv.loggedUser = auth.GetSignedInUser(rcv.Request)
	rcv.SetTitle("text02", rcv.loggedUser.Name)

	switch rcv.Request.Method {
	case "POST":
		auth.SignOut(rcv.Response, rcv.Request)
		if err := rcv.delete(); err != nil {
			rcv.SetFlash("E", "text20")
			rcv.Redirect("/", 303)
			return
		}
		rcv.SetFlash("I", "text21")
		rcv.Redirect("/", 303)
	}

}

func New() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		c := &controller{base.New(rw, r, false, "controller/account"), nil}
		c.serve()
	})
}
