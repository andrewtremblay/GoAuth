package password

import (
	"authsys/controllers/base"
	"authsys/controllers/captcha"
	"authsys/controllers/csrf"
	"authsys/middlewares/auth"
	maccount "authsys/models/account"
	"authsys/tools/caller"
	"github.com/mholt/binding"
	//"fmt"
	//"html/template"
	//"errors"
	"errors"
	"net/http"
	"path"
)

var (
	passwordTmpl string
)

func init() {
	passwordTmpl = path.Join(caller.Path(), "password.html")
}

type account struct {
	Name, OldPw, NewPw, ConfirmPw, Csrf, Captcha, Certification, Human string
	*base.Controller
}

func (rcv *account) FieldMap() binding.FieldMap {
	return binding.FieldMap{
		&rcv.NewPw:         "newpw",
		&rcv.OldPw:         "oldpw",
		&rcv.ConfirmPw:     "confirmpw",
		&rcv.Csrf:          "csrf",
		&rcv.Human:         "human",
		&rcv.Certification: "certification",
	}
}

func (rcv *account) Validate(r *http.Request, errs binding.Errors) binding.Errors {

	if rcv.OldPw != "" && rcv.NewPw != "" && rcv.ConfirmPw != "" {

		if rcv.NewPw != rcv.ConfirmPw {
			errs = append(errs, binding.Error{
				FieldNames:     []string{"ConfirmPw"},
				Classification: "ConfirmPwError",
				Message:        rcv.Translate("text17"),
			})
		}

	} else if (rcv.OldPw != "" && rcv.NewPw == "") || (rcv.OldPw == "" && rcv.NewPw != "") || (rcv.OldPw != "" && rcv.NewPw != "" && rcv.ConfirmPw == "") {
		errs = append(errs, binding.Error{
			FieldNames:     []string{"FieldsPw"},
			Classification: "EmptyError",
			Message:        rcv.Translate("text18"),
		})
	} else if rcv.OldPw == "" && rcv.NewPw == "" && rcv.ConfirmPw == "" {
		errs = append(errs, binding.Error{
			FieldNames:     []string{"FieldsPw"},
			Classification: "EmptyError",
			Message:        rcv.Translate("text18"),
		})
	}

	return errs
}

type controller struct {
	*base.Controller
	formUser   *account
	loggedUser *maccount.SignedUser
}

func (rcv *controller) get() {
	rcv.formUser.Name = rcv.loggedUser.Name
	rcv.formUser.Csrf = csrf.Token(rcv.Request)
	rcv.formUser.Captcha, rcv.formUser.Certification = captcha.Create()
	rcv.RenderContentPart(passwordTmpl, rcv.formUser)
}

func (rcv *controller) post() []error {
	var errs []error

	// Map html input value to fields
	if formErrs := binding.Bind(rcv.Request, rcv.formUser); formErrs != nil {

		for _, e := range formErrs {
			errs = append(errs, errors.New(e.Message))
		}

		return errs
	}

	// Validate captcha
	if err := captcha.Validate(rcv.Request, rcv.formUser.Certification, rcv.formUser.Human); err != nil {
		return []error{err}
	}

	if err := rcv.updatePassword(); err != nil {
		errs = append(errs, err)
	}

	return errs
}

// Password will be only updated, if all require password
// fields are filled.
func (rcv *controller) updatePassword() error {

	if err := maccount.ChangePassword(rcv.loggedUser.Email, rcv.formUser.OldPw, rcv.formUser.NewPw, rcv.Local); err != nil {
		return err
	}
	return nil

}

func (rcv *controller) serve() {

	rcv.formUser = &account{Controller: rcv.Controller}
	rcv.loggedUser = auth.GetSignedInUser(rcv.Request)
	rcv.SetTitle("text05", rcv.loggedUser.Name)

	switch rcv.Request.Method {
	case "GET":
		rcv.get()
	case "POST":

		errs := rcv.post()
		if errs == nil {
			// If successfull process
			rcv.SetFlash("I", "text19")
			auth.SignOut(rcv.Response, rcv.Request)
			return
		}

		for _, e := range errs {
			rcv.AppendError(e)
		}

		rcv.get()
	}
	rcv.Render()
}

func New() http.Handler {
	return csrf.New(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		c := &controller{base.New(rw, r, true, "controller/account"), nil, nil}
		c.serve()
	}))
}
