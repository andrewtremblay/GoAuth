package edit

import (
	"authsys/controllers/base"
	"authsys/controllers/captcha"
	"authsys/controllers/csrf"
	"authsys/middlewares/auth"
	maccount "authsys/models/account"
	"authsys/tools/caller"
	"authsys/tools/mail"
	"authsys/tools/redis"
	//"fmt"
	"github.com/dchest/uniuri"
	"github.com/mholt/binding"
	//"html/template"
	//"errors"
	"errors"
	"net/http"
	"path"
	"time"
)

const (
	i18nSec string = "controller/account"
)

var (
	editTmpl string
)

func init() {
	editTmpl = path.Join(caller.Path(), "edit.html")
}

type account struct {
	Name, Email, Csrf, Captcha, Certification, Human string
	*base.Controller
}

func (rcv *account) FieldMap() binding.FieldMap {
	return binding.FieldMap{
		&rcv.Name:          "name",
		&rcv.Email:         "email",
		&rcv.Csrf:          "csrf",
		&rcv.Human:         "human",
		&rcv.Certification: "certification",
	}
}

func (rcv *account) Validate(r *http.Request, errs binding.Errors) binding.Errors {

	// Get signed in user
	loggedUser := auth.GetSignedInUser(r)

	if rcv.Name == loggedUser.Name && rcv.Email == loggedUser.Email {
		errs = append(errs, binding.Error{
			FieldNames:     []string{"Name"},
			Classification: "NameError",
			Message:        rcv.Translate("text14"),
		})

		return errs
	}

	// Validate if the same name
	if rcv.Name != loggedUser.Name {
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
	}

	if rcv.Email != loggedUser.Email {
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
	}

	return errs
}

type controller struct {
	*base.Controller
	formUser   *account
	loggedUser *maccount.SignedUser
}

func (rcv *controller) get() {
	// Keep the input email
	if rcv.formUser.Email == "" {
		rcv.formUser.Email = rcv.loggedUser.Email
	}

	if rcv.formUser.Name == "" {
		rcv.formUser.Name = rcv.loggedUser.Name
	}

	rcv.formUser.Csrf = csrf.Token(rcv.Request)
	rcv.formUser.Captcha, rcv.formUser.Certification = captcha.Create()
	rcv.RenderContentPart(editTmpl, rcv.formUser)
}

func (rcv *controller) put() []error {
	var errs []error

	// Map html input value to fields
	if fieldsErrs := binding.Bind(rcv.Request, rcv.formUser); fieldsErrs != nil {

		for _, e := range fieldsErrs {
			errs = append(errs, errors.New(e.Message))
		}
		return errs
	}

	// Validate captcha
	if err := captcha.Validate(rcv.Request, rcv.formUser.Certification, rcv.formUser.Human); err != nil {
		return []error{err}
	}

	if err := rcv.updateEmail(); err != nil {
		errs = append(errs, err)
	}

	if err := rcv.updateName(); err != nil {
		errs = append(errs, err)
	}

	return errs
}

// Updated only, when email address has changed.
func (rcv *controller) updateEmail() error {

	if rcv.loggedUser.Email != rcv.formUser.Email {
		if err := maccount.UpdateEmail(rcv.loggedUser.Email, rcv.formUser.Email, rcv.Local); err != nil {
			return err
		}

		if err := rcv.sendActivationLink(rcv.formUser.Email); err != nil {
			return err
		}
		// Email activation message
		rcv.SetFlash("I", "text16")
	}

	return nil
}

// Updated only, when name has changed.
func (rcv *controller) updateName() error {

	if rcv.loggedUser.Name != rcv.formUser.Name {

		if err := maccount.UpdateName(rcv.loggedUser.Email, rcv.formUser.Name, rcv.Local); err != nil {
			return err
		}

		rcv.SetFlash("I", "text15")
	}

	return nil

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

	// Will delete the link in 48 hours
	con.Do("EXPIREAT", uri, time.Now().Add(time.Hour*48).Unix())

	link := rcv.Request.Host + "/activate/" + uri
	if err = mail.Send(email, link); err != nil {
		return err
	}
	return nil
}

func (rcv *controller) serve() {
	rcv.formUser = &account{Controller: rcv.Controller}
	rcv.loggedUser = auth.GetSignedInUser(rcv.Request)
	rcv.SetTitle("text03", rcv.loggedUser.Name)

	switch rcv.Request.Method {
	case "GET":
		rcv.get()
	case "POST":

		errs := rcv.put()
		if errs == nil {
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
