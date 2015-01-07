package auth

import (
	"authsys/models/account"
	"authsys/tools/i18n"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"net/http"
	"regexp"
	"time"
)

type writer struct {
	response   http.ResponseWriter
	signedUser *account.SignedUser
	local      string
}

func (rcv *writer) validateEmail() error {

	pattern := regexp.MustCompile(`(([^<>()[\]\\.,;:\s@\"]+(\.[^<>()[\]\\.,;:\s@\"]+)*)|(\".+\"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$`)

	if !pattern.MatchString(rcv.signedUser.Email) {
		return errors.New(i18n.Translate(rcv.local, "models/account", "text04"))
	}

	return nil
}

// Write generated token into cookies
func (rcv *writer) token() error {

	if err := rcv.validateEmail(); err != nil {
		return err
	}

	token := jwt.New(jwt.GetSigningMethod("RS256"))

	// Set authentication data
	token.Claims["email"] = rcv.signedUser.Email
	token.Claims["name"] = rcv.signedUser.Name
	token.Claims["logged"] = true

	// Set expired time
	token.Claims["exp"] = time.Now().Add(time.Minute * 30).Unix()
	signed, err := token.SignedString(privateSignedKey)
	if err != nil {
		return err
	}

	rcv.setCookie(signed)
	return nil
}

// Write cookie and the generated jwt token is used
// as a identifier
func (rcv *writer) setCookie(signed string) {

	http.SetCookie(rcv.response, &http.Cookie{
		Name:     cookie,
		Value:    signed,
		Path:     "/",
		MaxAge:   0,
		Secure:   false, // For https connection
		HttpOnly: true,  // Prevent to access cookie via javascript
	})

}

// Remember signed in user
func SignedIn(response http.ResponseWriter, signedUser *account.SignedUser, local string) error {

	w := &writer{response: response, signedUser: signedUser, local: local}
	if err := w.token(); err != nil {
		return err
	}

	return nil

}
