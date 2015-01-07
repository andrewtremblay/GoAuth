package session

import (
	"authsys/tools/caller"
	"authsys/tools/context"
	"authsys/tools/redis"
	"errors"
	"fmt"
	"github.com/codegangsta/negroni"
	"github.com/dchest/uniuri"
	"github.com/dgrijalva/jwt-go"
	redigo "github.com/garyburd/redigo/redis"
	"io/ioutil"
	"net/http"
	"path"
	"time"
)

/* Generate and validate json web token object.
 * The key is valid for 15 minutes, after then it
 * will generate new key and send to client.
 */

// Error is an error implementation that includes a time and message.
type expiredTokenError struct {
}

func (r *expiredTokenError) Error() string {
	return "Session id token is expired."
}

type generallyTokenError struct {
}

func (r *generallyTokenError) Error() string {
	return "Session id token error."
}

const (
	cookieName            = "STORAGE"
	privateKeyFile string = "/session.rsa"
	publicKeyFile  string = "/session.rsa.pub"
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

/*
 * Use JWT to verify user, replace session. Generate the
 * JWT token and save in cookie on client side.
 *
 */

type controller struct {
	response http.ResponseWriter
	request  *http.Request
	cookie   *http.Cookie
}

// The first returned value is session id and second is error state
func (rcv *controller) readJwt(publicKey string) (string, error) {

	// validate the token
	token, err := jwt.Parse(publicKey, func(token *jwt.Token) (interface{}, error) {

		if _, ok := token.Claims["id"]; !ok {
			return "", &generallyTokenError{}
		}

		// since we only use the one private key to sign the tokens,
		// we also use only the public counter part to verify it.
		return publicVerifyKey, nil
	})

	if token == nil {
		return "", &generallyTokenError{}
	}

	if err != nil {
		errType := err.(*jwt.ValidationError)

		switch errType.Errors {

		case jwt.ValidationErrorExpired:
			return token.Claims["id"].(string), &expiredTokenError{}

		default:
			return "", &generallyTokenError{}
		}

	}
	return token.Claims["id"].(string), nil
}

// Generate JWT. Would the token be expired, then the user
// will get new token but keep the same id. Parameter is optional,
// because when user visit the page first time, it will generate
// random id.
func (rcv *controller) createJwt(oldSid ...string) (string, error) {

	token := jwt.New(jwt.GetSigningMethod("RS256"))

	// Assign the already exists session id
	if oldSid != nil {
		token.Claims["id"] = oldSid[0]
	} else {
		token.Claims["id"] = uniuri.NewLen(20)
	}

	// Set session id on context, that would be available, when user
	// visit the site on first time.
	context.Set(rcv.request, context.SID, token.Claims["id"])

	token.Claims["exp"] = time.Now().Add(time.Minute * 15).Unix()
	signed, err := token.SignedString(privateSignedKey)
	if err != nil {
		return "", err
	}

	return signed, nil

}

func (rcv *controller) createCookie(oldSid ...string) {

	token, err := rcv.createJwt(oldSid...)
	if err != nil {
		rcv.response.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(rcv.response, &http.Cookie{
		Name:     cookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   0,
		Secure:   false, // For https connection
		HttpOnly: true,  // Prevent to access cookie via javascript
	})

}

// Validate, if cookie is already set, otherwise
// it will be set.
func (rcv *controller) readCookie() error {

	// Read cookie from client
	c, err := rcv.request.Cookie(cookieName)

	// If cookie does not exist on the client yet
	if err != nil {
		rcv.createCookie()
		return nil
	}

	sid, err := rcv.readJwt(c.Value)

	switch err.(type) {
	case *expiredTokenError:
		//fmt.Println("Token expired set new one.")
		// Set new token on cookie, but keep the session id
		rcv.createCookie(sid)
		return nil
	case *generallyTokenError:
		return errors.New("Their is something wrong with verification. Please restart the browser.")

	}
	// Set session identification in the current context
	context.Set(rcv.request, context.SID, sid)
	rcv.renewTime(sid)
	return nil

}

// After every incoming request, refresh the time to live
// of session, to ensure that user is still send request
// to the server
func (rcv *controller) renewTime(sid string) error {

	// Validate, if the session identification already exists in redis
	c := redis.Get()
	exists, err := redigo.Bool(c.Do("EXISTS", sid))
	if err != nil {
		return err
	}

	if !exists {
		return nil
	}

	c.Do("EXPIREAT", sid, time.Now().Add(time.Minute*30).Unix())
	return nil
}

func (rcv *controller) handle() error {

	if err := rcv.readCookie(); err != nil {
		return err
	}

	return nil

}

func New() negroni.HandlerFunc {
	fmt.Println("Middleware session is started.")
	return negroni.HandlerFunc(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {

		c := &controller{request: r, response: rw}
		if err := c.handle(); err != nil {
			fmt.Fprintf(rw, err.Error())
			return
		}
		next(rw, r)
	})
}
