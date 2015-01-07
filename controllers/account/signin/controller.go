package signin

import (
	"authsys/controllers/base"
	"authsys/controllers/csrf"
	"authsys/middlewares/auth"
	maccount "authsys/models/account"
	"authsys/tools/caller"
	"errors"
	//"fmt"
	"github.com/mholt/binding"
	"net/http"
	"path"
)

var (
	signInTmpl string
)

func init() {
	signInTmpl = path.Join(caller.Path(), "signin.html")
}

/*
 * Model
 */

type account struct {
	Email, Password, Csrf string
	*base.Controller
}

func (rcv *account) FieldMap() binding.FieldMap {
	return binding.FieldMap{
		&rcv.Email:    "email",
		&rcv.Password: "password",
		&rcv.Csrf:     "csrf",
	}
}

func (rcv *account) Validate(r *http.Request, errs binding.Errors) binding.Errors {

	if rcv.Email == "" {
		errs = append(errs, binding.Error{
			FieldNames:     []string{"Email"},
			Classification: "EmptyError",
			Message:        rcv.Translate("text02"),
		})
	}

	if rcv.Password == "" {
		errs = append(errs, binding.Error{
			FieldNames:     []string{"Password"},
			Classification: "EmptyError",
			Message:        rcv.Translate("text04"),
		})
	}

	return errs
}

/*
 * Controller
 */
type controller struct {
	*base.Controller
	signedUser *maccount.SignedUser
}

func (rcv *controller) post() error {

	var err error

	formUser := &account{Controller: rcv.Controller}

	if errBind := binding.Bind(rcv.Request, formUser); errBind != nil {

		return errors.New(rcv.Translate("text13"))
	}

	rcv.signedUser, err = maccount.SignIn(formUser.Email, formUser.Password, rcv.Local)
	if err != nil {
		return err
	}

	if err := auth.SignedIn(rcv.Response, rcv.signedUser, rcv.Local); err != nil {
		return err
	}

	return nil
}

func (rcv *controller) get() {
	rcv.RenderContentPart(signInTmpl, csrf.Token(rcv.Request))
}

func (rcv *controller) serve() {
	rcv.SetTitle("text07")

	switch rcv.Request.Method {
	case "GET":
		rcv.get()
	case "POST":
		err := rcv.post()
		if err == nil {
			// If everything was successfully
			rcv.Redirect("/", 303)
			return
		}

		rcv.AppendError(err)
		rcv.get()

	}
	rcv.Render()

}

func New() http.Handler {
	return csrf.New(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		c := &controller{base.New(rw, r, true, "controller/account"), nil}
		c.serve()
	}))
}
