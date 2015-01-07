package resetpw

import (
	"authsys/controllers/base"
	//"authsys/middlewares/httphead"
	"authsys/controllers/csrf"
	maccount "authsys/models/account"
	"authsys/tools/caller"
	"authsys/tools/redis"
	//"fmt"
	"authsys/controllers/notfound"
	goredis "github.com/garyburd/redigo/redis"
	"github.com/gorilla/mux"
	"github.com/mholt/binding"
	//"html/template"
	"errors"
	"net/http"
	"path"
)

var (
	resetpwTmpl string
)

func init() {
	resetpwTmpl = path.Join(caller.Path(), "resetpw.html")
}

type account struct {
	NewPw, ConfirmPw, Csrf string
	*base.Controller
}

func (rcv *account) FieldMap() binding.FieldMap {
	return binding.FieldMap{
		&rcv.NewPw:     "newpw",
		&rcv.ConfirmPw: "confirmpw",
		&rcv.Csrf:      "csrf",
	}
}

func (rcv *account) Validate(r *http.Request, errs binding.Errors) binding.Errors {

	if rcv.NewPw == "" || rcv.ConfirmPw == "" {
		errs = append(errs, binding.Error{
			FieldNames:     []string{"Password"},
			Classification: "PasswordError",
			Message:        rcv.Translate("text18"),
		})

	} else if err := maccount.ValidatePassword(rcv.NewPw, rcv.Local); err != nil {
		errs = append(errs, binding.Error{
			FieldNames:     []string{"Password"},
			Classification: "PasswordError",
			Message:        err.Error(),
		})
	} else if rcv.NewPw != rcv.ConfirmPw {
		errs = append(errs, binding.Error{
			FieldNames:     []string{"Password"},
			Classification: "PasswordError",
			Message:        rcv.Translate("text17"),
		})
	}

	return errs
}

type controller struct {
	*base.Controller
	formUser *account
}

// Read id from link, for example /user/reset/12345.
// 12345 is the id to read from redis database.
// Return saved email address.
func (rcv *controller) isLinkValid() error {

	_, err := rcv.readEmailAddr()
	return err
}

// Read email address from redis, that mapped to the
// link id
func (rcv *controller) readEmailAddr() (string, error) {
	id := mux.Vars(rcv.Request)["id"]
	conn := redis.Get()
	return goredis.String(conn.Do("GET", id))
}

// Reset password in the database
func (rcv *controller) resetPassword(email string) error {

	return maccount.ResetPassword(email, rcv.formUser.NewPw, rcv.Local)
}

func (rcv *controller) get() error {

	// If the entered link is not valid
	if err := rcv.isLinkValid(); err != nil {
		return err
	}

	rcv.RenderContentPart(resetpwTmpl, nil)
	return nil
}

func (rcv *controller) post() []error {

	// Map html input value to fields
	if formErrs := binding.Bind(rcv.Request, rcv.formUser); formErrs != nil {

		var errs []error

		for _, e := range formErrs {
			errs = append(errs, errors.New(e.Message))
		}
		return errs
	}

	email, err := rcv.readEmailAddr()
	if err != nil {
		return []error{err}
	}

	if err := rcv.resetPassword(email); err != nil {
		return []error{err}
	}

	return nil
}

func (rcv *controller) serve() {

	rcv.formUser = &account{Controller: rcv.Controller}
	rcv.SetTitle("text06")

	switch rcv.Request.Method {
	case "GET":
		if err := rcv.get(); err != nil {
			notfound.Serve(rcv.Controller)
			return
		}
	case "POST":

		errs := rcv.post()
		if errs == nil {
			// If successfull process
			rcv.SetFlash("I", "text19")
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
		c := &controller{base.New(rw, r, false, "controller/account"), nil}
		c.serve()
	}))
}
