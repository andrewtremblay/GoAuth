package controllers

import (
	"authsys/controllers/account/activate"
	"authsys/controllers/account/delete"
	"authsys/controllers/account/edit"
	"authsys/controllers/account/forgotpw"
	"authsys/controllers/account/password"
	"authsys/controllers/account/resetpw"
	"authsys/controllers/account/signin"
	"authsys/controllers/account/signup"
	"authsys/controllers/account/view"
	"authsys/controllers/captcha"
	"authsys/controllers/notfound"
	"authsys/controllers/welcome"
	"authsys/middlewares/auth"
	"authsys/mux"
	gomux "github.com/gorilla/mux"
)

func Routes() *gomux.Router {

	// Router handlers
	r := mux.Router
	r.NotFoundHandler = notfound.New()

	r.Handle("/", welcome.New()).Methods("GET")
	r.Handle("user/activate/{id}", activate.New()).Methods("GET").Name("userav")
	r.Handle("/user/signup", auth.PreventVisit(signup.New())).Methods("GET", "POST")
	r.Handle("/user/signin", auth.PreventVisit(signin.New())).Methods("GET", "POST")
	r.Handle("/user/forgot", auth.PreventVisit(forgotpw.New())).Methods("GET", "POST")
	r.Handle("/user/reset/{id}", resetpw.New()).Methods("GET", "POST").Name("resetpw")

	r.Handle("/user/{name}/signout", auth.AllowVisit(auth.SelfSignOut())).Methods("GET")
	r.Handle("/user/{name}/view", auth.AllowVisit(view.New())).Methods("GET").Name("userview")
	r.Handle("/user/{name}/edit", auth.AllowVisit(edit.New())).Methods("GET", "POST").Name("useredit")
	r.Handle("/user/{name}/password", auth.AllowVisit(password.New())).Methods("GET", "POST").Name("userpw")
	r.Handle("/user/{name}/delete", auth.AllowVisit(delete.New())).Methods("POST")
	r.Handle("/captcha/{image}", captcha.New())

	return r

}
