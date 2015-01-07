package auth

import (
	"authsys/middlewares/httphead"
	"authsys/models/account"
	"authsys/tools/caller"
	"authsys/tools/context"
	"authsys/tools/flash"
	"authsys/tools/i18n"
	"errors"
	//"fmt"
	"github.com/codegangsta/negroni"
	"github.com/dgrijalva/jwt-go"
	"io/ioutil"
	"net/http"
	"path"
)

const (
	cookie                = "VERIFIED"
	i18nSec        string = "controller/account"
	privateKeyFile string = "/token.rsa"
	publicKeyFile  string = "/token.rsa.pub"
)

var (
	// privateSignedKey is for generate a token
	privateSignedKey []byte
	// publicVerifyKey is for validate if token is valid
	publicVerifyKey []byte
)

func init() {

	var err error

	privateSignedKey, err = ioutil.ReadFile(path.Join(caller.Path(), privateKeyFile))
	if err != nil {
		panic(err)
	}

	publicVerifyKey, err = ioutil.ReadFile(path.Join(caller.Path(), publicKeyFile))
	if err != nil {
		panic(err)
	}

}

func IsSignedInUser(r *http.Request) bool {
	user := context.Get(r, context.SIGNEDID)
	if user == nil {
		return false
	}
	return true
}

// Get the current signed in user
func GetSignedInUser(r *http.Request) *account.SignedUser {
	user := context.Get(r, context.SIGNEDID)
	if user == nil {
		return nil
	}
	return user.(*account.SignedUser)
}

// If an user is already signed in and nevertheless call for example
// /user/signin, it redirect to signed in page view
// func PreventVisit() negroni.Handler {
// 	return negroni.HandlerFunc(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
// 		signed := IsSignedInUser(r)
// 		if !signed {
// 			next(rw, r)
// 			return
// 		}

// 		http.Redirect(rw, r, "/", 303)
// 	})
// }

func PreventVisit(handler http.Handler) http.Handler {
	n := negroni.New()
	n.Use(negroni.HandlerFunc(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		signed := IsSignedInUser(r)
		if !signed {
			next(rw, r)
			return
		}

		http.Redirect(rw, r, "/", 303)
	}))
	n.UseHandler(handler)
	return n
}

// Validate, if the user is authorized to use the handler.
// Will use like middleware.
func AllowVisit(handler http.Handler) http.Handler {
	n := negroni.New(negroni.HandlerFunc(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {

		signed := IsSignedInUser(r)

		if !signed {
			flash.Set(r, "E", i18n.Translate(httphead.GetLang(r), i18nSec, "text12"))
			http.Redirect(rw, r, "/user/signin", 303)
			return
		}
		// If authorization is passed
		next(rw, r)
	}))
	n.UseHandler(handler)
	return n
}

type expiredTokenError struct {
	email string
}

func (rcv *expiredTokenError) Error() string {
	return "Session expired error, please sing in again."
}

type jwtTokenError struct {
}

func (rcv *jwtTokenError) Error() string {
	return "Internal error."
}

type violatedTokenError struct {
}

func (rcv *violatedTokenError) Error() string {
	return "Violation token error."
}

type notSignedError struct {
	local string
}

func (rcv *notSignedError) Error() string {
	return i18n.Translate(rcv.local, i18nSec, "text12")
}

/*
 * Read authorized token from cookie and validate
 * if the user has permissions to handle controllers
 */
type reader struct {
	request  *http.Request
	response http.ResponseWriter
	local    string

	// Public
	SignedUser *account.SignedUser
}

// Read jwt token from cookie
func (rcv *reader) readCookie() (string, error) {
	c, err := rcv.request.Cookie(cookie)
	if err != nil {
		return "", errors.New(i18n.Translate(rcv.local, i18nSec, "text12"))
	}

	return c.Value, nil
}

// Validate jwt token
func (rcv *reader) readToken(signed string) error {
	// validate the token
	token, err := jwt.Parse(signed, func(token *jwt.Token) (interface{}, error) {

		// Check if token contains the keys

		if _, ok := token.Claims["email"]; !ok {
			return nil, &violatedTokenError{}
		}

		if _, ok := token.Claims["name"]; !ok {
			return nil, &violatedTokenError{}
		}

		if _, ok := token.Claims["logged"]; !ok {
			return nil, &violatedTokenError{}
		}

		// since we only use the one private key to sign the tokens,
		// we also use only the public counter part to verify it.
		return publicVerifyKey, nil
	})

	if token == nil {
		return &jwtTokenError{}
	}

	if err != nil {

		switch err.(*jwt.ValidationError).Errors {

		case jwt.ValidationErrorExpired:
			return &expiredTokenError{}

		default:
			return &jwtTokenError{}
		}

	}

	if signedIn := token.Claims["logged"].(bool); !signedIn {
		return &notSignedError{local: rcv.local}
	}

	// Renew authentication token every request, to ensure that
	// the user is still active and the signed in timeout will be
	// renew.
	rcv.SignedUser = &account.SignedUser{Name: token.Claims["name"].(string), Email: token.Claims["email"].(string)}
	if err = SignedIn(rcv.response, rcv.SignedUser, rcv.local); err != nil {
		return err
	}

	// Provide signed user during a request
	rcv.setSignedUserContext()

	return nil
}

// Set signed user during a request
func (rcv *reader) setSignedUserContext() {
	context.Set(rcv.request, context.SIGNEDID, rcv.SignedUser)
}

func New() negroni.Handler {
	return negroni.HandlerFunc(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		rd := &reader{request: r, response: rw, local: httphead.GetLang(r)}

		if cookie, err := rd.readCookie(); err == nil {
			rd.readToken(cookie)
		}

		// If authorization is passed
		next(rw, r)
	})
}
