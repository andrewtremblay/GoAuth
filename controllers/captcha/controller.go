package captcha

import (
	"authsys/tools/redis"
	"bytes"
	"encoding/base64"
	//"fmt"
	"authsys/middlewares/httphead"
	"authsys/tools/i18n"
	"errors"
	"github.com/dchest/captcha"
	"github.com/dchest/uniuri"
	goredis "github.com/garyburd/redigo/redis"
	"github.com/gorilla/mux"
	"image/png"
	"net/http"
	"net/url"
	"strings"
)

var (
	// Captcha expired time in 10 minute
	expired int = 600
)

/*
 * Response image request for the client.
 */

type controller struct {
	request  *http.Request
	response http.ResponseWriter
}

// Will serve image, when user does not click reload
// yet.
func (rcv *controller) createNewImage(image string) error {

	c := redis.Get()
	secret, err := goredis.Bytes(c.Do("GET", image))
	if err != nil {
		return err
	}
	png.Encode(rcv.response, captcha.NewImage(image, secret, captcha.StdWidth, captcha.StdHeight))
	return nil
}

// Request new image, if the previous captcha is difficult to recognize.
func (rcv *controller) changeImage(image string) {

	c := redis.Get()
	// Configure out, if the image still available.
	_, err := goredis.Bytes(c.Do("GET", image))
	if err != nil {
		http.NotFound(rcv.response, rcv.request)
		return
	}
	secret := captcha.RandomDigits(7)
	if _, err := c.Do("SET", image, secret); err != nil {
		panic(err.Error())
	}

	if _, err := c.Do("EXPIRE", image, expired); err != nil {
		panic(err.Error())
	}

	png.Encode(rcv.response, captcha.NewImage(image, secret, captcha.StdWidth, captcha.StdHeight))
}

func (rcv *controller) serve() {

	// Read the filename of image
	image := strings.Split(mux.Vars(rcv.request)["image"], ".")[0]

	// Configure out, if the user reload image or not
	query, _ := url.ParseQuery(rcv.request.URL.RawQuery)
	if len(query) > 0 {
		rcv.changeImage(image)
		return
	}

	// When the image is violated, it will send an status error back
	if err := rcv.createNewImage(image); err != nil {
		http.NotFound(rcv.response, rcv.request)
	}
}

// Request handler
func New() http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		c := &controller{response: rw, request: r}
		c.serve()
	})
}

// Generate png filename and encode it. The encoded code will
// will save on html and will by post request decoded, that
// use to identify, if the user have enter the right captcha code.
func Create() (string, string) {
	image := uniuri.NewLen(25)
	secret := captcha.RandomDigits(7)

	c := redis.Get()
	if _, err := c.Do("SET", image, secret); err != nil {
		panic(err.Error())
	}

	if _, err := c.Do("EXPIRE", image, expired); err != nil {
		panic(err.Error())
	}

	return image, base64.StdEncoding.EncodeToString([]byte(image))
}

// Validate if the entered numbers match to stored number
func Validate(r *http.Request, certification, human string) error {

	// Error object
	err := errors.New(i18n.Translate(httphead.GetLang(r), "controller/account", "text09"))

	if human == "" {
		return err
	}

	decoded, err := base64.StdEncoding.DecodeString(certification)
	if err != nil {
		return err
	}

	c := redis.Get()
	// Configure out, if the image still available.
	values, err := goredis.Bytes(c.Do("GET", string(decoded)))
	if err != nil {
		return err
	}

	ns := make([]byte, len(human))
	for i := range ns {
		d := human[i]
		switch {
		case '0' <= d && d <= '9':
			ns[i] = d - '0'
		case d == ' ' || d == ',':
			// ignore
		default:
			return err
		}
	}

	if !bytes.Equal(values, ns) {
		return err
	}

	return nil

}
