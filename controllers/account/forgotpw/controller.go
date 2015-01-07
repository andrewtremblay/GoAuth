package forgotpw

import (
	"authsys/controllers/base"
	"authsys/controllers/captcha"
	"authsys/controllers/csrf"
	maccount "authsys/models/account"
	"authsys/tools/caller"
	"authsys/tools/mail"
	"authsys/tools/redis"
	"errors"
	"github.com/dchest/uniuri"
	//"fmt"
	"github.com/mholt/binding"
	"net/http"
	"path"
	"time"
)

var (
	forgotpwTmpl string
)

func init() {
	forgotpwTmpl = path.Join(caller.Path(), "forgotpw.html")
}

/*
 * Model
 */

type account struct {
	Email, Csrf, Captcha, Certification, Human string
	*base.Controller
}

func (rcv *account) FieldMap() binding.FieldMap {
	return binding.FieldMap{
		&rcv.Email:         "email",
		&rcv.Csrf:          "csrf",
		&rcv.Human:         "human",
		&rcv.Certification: "certification",
	}
}

func (rcv *account) Validate(r *http.Request, errs binding.Errors) binding.Errors {

	if rcv.Email == "" {
		errs = append(errs, binding.Error{
			FieldNames:     []string{"Email"},
			Classification: "EmptyError",
			Message:        rcv.Translate("text02"),
		})
	} else if _, err := maccount.Read(rcv.Email, rcv.Local); err != nil {
		errs = append(errs, binding.Error{
			FieldNames:     []string{"Email"},
			Classification: "EmailNotExists",
			Message:        err.Error(),
		})
	}

	return errs
}

/*
 * Controller
 */
type controller struct {
	*base.Controller
	formUser *account
}

// Validate entered data, included if the entered email is exist
func (rcv *controller) validate() []error {

	var errs []error

	// Map html input value to fields
	if fieldsErrs := binding.Bind(rcv.Request, rcv.formUser); fieldsErrs != nil {

		for _, e := range fieldsErrs {
			errs = append(errs, errors.New(e.Message))
		}
		return errs
	}

	if _, err := maccount.Read(rcv.formUser.Email, rcv.Local); err != nil {
		return []error{err}
	}

	// Validate captcha
	if err := captcha.Validate(rcv.Request, rcv.formUser.Certification, rcv.formUser.Human); err != nil {
		return []error{err}
	}

	return errs
}

// Build the reset link, that user can call page
// to reset the password
func (rcv *controller) buildLink() (string, error) {
	id := uniuri.NewLen(17)

	con := redis.Get()
	if _, err := con.Do("SET", id, rcv.formUser.Email); err != nil {
		return "", err
	}

	// The link will be deleted in 24 hours. After then, the user
	// have to reqeuest for changing password again.
	con.Do("EXPIREAT", id, time.Now().Add(time.Hour*24).Unix())

	url, err := rcv.Router.Get("resetpw").URL("id", id)
	if err != nil {
		return "", err
	}

	link := rcv.Request.Host + url.String()

	return link, nil

}

// Send the link to email addresse
func (rcv *controller) sendLink(link string) error {

	if err := mail.Send(rcv.formUser.Email, link); err != nil {
		return err
	}
	return nil

}

func (rcv *controller) post() []error {

	if errs := rcv.validate(); errs != nil {
		return errs
	}

	url, err := rcv.buildLink()
	if err != nil {
		return []error{err}
	}

	if err := rcv.sendLink(url); err != nil {
		return []error{err}
	}

	return nil
}

func (rcv *controller) get() {

	// Generate and set csrf token into html
	rcv.formUser.Csrf = csrf.Token(rcv.Request)
	rcv.formUser.Captcha, rcv.formUser.Certification = captcha.Create()
	rcv.RenderContentPart(forgotpwTmpl, rcv.formUser)

}

func (rcv *controller) serve() {
	rcv.SetTitle("text04")
	rcv.formUser = &account{Controller: rcv.Controller}

	switch rcv.Request.Method {
	case "GET":
		rcv.get()
	case "POST":
		if errs := rcv.post(); errs != nil {
			for _, e := range errs {
				rcv.AppendError(e)
			}
			rcv.get()
		} else {
			rcv.SetFlash("I", "text22")
			rcv.Redirect("/", 303)
			return
		}

	}
	rcv.Render()

}

func New() http.Handler {
	return csrf.New(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		c := &controller{base.New(rw, r, true, "controller/account"), nil}
		c.serve()
	}))
}
