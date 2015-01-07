package signup

import (
	"authsys/controllers/base"
	"authsys/controllers/captcha"
	"authsys/controllers/csrf"
	maccount "authsys/models/account"
	"authsys/tools/caller"
	"authsys/tools/mail"
	"authsys/tools/redis"
	"errors"
	//"fmt"
	"github.com/dchest/uniuri"
	"github.com/mholt/binding"
	"net/http"
	"path"
	"time"
)

var (
	signUpTmpl string
)

func init() {
	signUpTmpl = path.Join(caller.Path(), "signup.html")
}

/*
 * Model
 */

type account struct {
	*base.Controller
	Name, Email, Emailconfirm, Password, Csrf, Captcha, Certification, Human string
	TermOf                                                                   bool
}

func (rcv *account) FieldMap() binding.FieldMap {
	return binding.FieldMap{
		&rcv.Name:          "name",
		&rcv.Email:         "email",
		&rcv.Emailconfirm:  "emailconfirm",
		&rcv.Password:      "password",
		&rcv.TermOf:        "termof",
		&rcv.Csrf:          "csrf",
		&rcv.Human:         "human",
		&rcv.Certification: "certification",
	}
}

func (rcv *account) Validate(r *http.Request, errs binding.Errors) binding.Errors {

	if rcv.Name == "" {
		errs = append(errs, binding.Error{
			FieldNames:     []string{"Name"},
			Classification: "EmptyError",
			Message:        rcv.Translate("text01"),
		})
	} else if err := maccount.ValidateName(rcv.Name, rcv.Local); err != nil {
		errs = append(errs, binding.Error{
			FieldNames:     []string{"Name"},
			Classification: "NameError",
			Message:        err.Error(),
		})
	}

	if rcv.Email == "" {
		errs = append(errs, binding.Error{
			FieldNames:     []string{"Email"},
			Classification: "EmptyError",
			Message:        rcv.Translate("text02"),
		})
	} else if err := maccount.ValidateEmail(rcv.Email, rcv.Local); err != nil {
		errs = append(errs, binding.Error{
			FieldNames:     []string{"Email"},
			Classification: "EmailError",
			Message:        err.Error(),
		})
	}

	if rcv.Emailconfirm == "" {
		errs = append(errs, binding.Error{
			FieldNames:     []string{"Emailconfirm"},
			Classification: "EmptyError",
			Message:        rcv.Translate("text03"),
		})
	} else if rcv.Emailconfirm != rcv.Email {
		errs = append(errs, binding.Error{
			FieldNames:     []string{"Email equality"},
			Classification: "NotEqualError",
			Message:        rcv.Translate("text07"),
		})
	}

	if rcv.Password == "" {
		errs = append(errs, binding.Error{
			FieldNames:     []string{"Password"},
			Classification: "EmptyError",
			Message:        rcv.Translate("text04"),
		})
	} else if err := maccount.ValidatePassword(rcv.Password, rcv.Local); err != nil {
		errs = append(errs, binding.Error{
			FieldNames:     []string{"Password"},
			Classification: "PasswordError",
			Message:        err.Error(),
		})
	}

	if rcv.TermOf == false {
		errs = append(errs, binding.Error{
			FieldNames:     []string{"Emailconfirm"},
			Classification: "EmptyError",
			Message:        rcv.Translate("text06"),
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
		return []error{errors.New(rcv.Translate("text09"))}
	}

	if errs := maccount.Create(rcv.formUser.Name, rcv.formUser.Email, rcv.formUser.Password, rcv.Local, rcv.formUser.TermOf); errs != nil {
		return errs
	}

	if err := rcv.sendActivationLink(rcv.formUser.Email); err != nil {
		return []error{err}
	}

	return nil

}

func (rcv *controller) get() {

	// Generate and set csrf token into html
	rcv.formUser.Csrf = csrf.Token(rcv.Request)
	rcv.formUser.Captcha, rcv.formUser.Certification = captcha.Create()
	rcv.RenderContentPart(signUpTmpl, rcv.formUser)

}

// After successfully signed up, it will send a confirmation
// email to user with the to activated link. Redis will keep
// this uri for 24 hours. If the user does not activated the
// account within this time, it will deleted from neo4j data-
// base and the user have to sign up again.
func (rcv *controller) sendActivationLink(email string) error {

	uri := uniuri.NewLen(20)
	expired := time.Now().Unix() + 86400
	con := redis.Get()
	_, err := con.Do("HMSET", uri, "email", email, "expired", expired)
	if err != nil {
		return err
	}

	// Will delete the data in 48 hours
	con.Do("EXPIREAT", uri, time.Now().Add(time.Hour*48).Unix())

	url, err := rcv.Router.Get("userav").URL("id", uri)
	if err != nil {
		return err
	}
	link := rcv.Request.Host + url.String()

	if err = mail.Send(email, link); err != nil {
		return err
	}
	return nil
}

func (rcv *controller) serve() {

	rcv.formUser = &account{Controller: rcv.Controller}
	rcv.SetTitle("text08")

	switch rcv.Request.Method {
	case "GET":
		rcv.get()
	case "POST":

		errs := rcv.post()
		if errs == nil {
			rcv.SetFlash("I", "text08")
			rcv.Redirect("/", 303)
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
		c := &controller{base.New(rw, r, true, "controller/account"), nil}
		c.serve()
	}))
}
