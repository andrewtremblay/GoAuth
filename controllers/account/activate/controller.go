package activate

import (
	"authsys/controllers/base"
	"authsys/controllers/notfound"
	maccount "authsys/models/account"
	"authsys/tools/caller"
	"authsys/tools/redis"
	"errors"
	//"fmt"
	goredis "github.com/garyburd/redigo/redis"
	"github.com/gorilla/mux"
	"net/http"
	"path"
	"time"
)

type data struct {
	Email   string `redis:"email"`
	Expired int64  `redis:"expired"`
}

type expiredError struct {
	*base.Controller
}

func (rcv *expiredError) Error() string {
	return rcv.Translate("text10")
}

var (
	expiredTmpl string
)

func init() {
	expiredTmpl = path.Join(caller.Path(), "expired.html")
}

type controller struct {
	*base.Controller
	store *data
}

// Read activate id from url and validate if the id
// can be activated.
func (rcv *controller) read() error {
	id := mux.Vars(rcv.Request)["id"]

	con := redis.Get()
	// Get the saved id to activate from redis
	reply, err := goredis.Values(con.Do("HGETALL", id))
	con.Do("DEL", id)
	if err != nil {
		return err
	}

	rcv.store = new(data)
	if err := goredis.ScanStruct(reply, rcv.store); err != nil {
		return err
	}

	return nil
}

// Validate if the acccount is registered for
// activation.
func (rcv *controller) validate() error {

	if err := rcv.read(); err != nil {
		return err
	}

	// If account does not exists
	if rcv.store.Email == "" {
		return errors.New(rcv.Translate("text11"))
	}

	// If time for activating account is expired
	if time.Now().Unix() > rcv.store.Expired {
		// Delete registered user from neo4j
		maccount.Delete(rcv.store.Email, rcv.Local)
		return &expiredError{rcv.Controller}
	}

	return nil
}

func (rcv *controller) get() error {

	// Activate account
	if err := maccount.Activate(rcv.store.Email, rcv.Local); err != nil {
		rcv.RenderContentPart(expiredTmpl, err.Error())
		return err
	}

	rcv.SetFlash("I", "text11")
	rcv.Redirect("/", 303)
	return nil
}

func (rcv *controller) serve() {

	rcv.SetTitle("text01")
	// Validate information for activating
	if err := rcv.validate(); err != nil {
		switch err.(type) {
		case *expiredError:
			rcv.RenderContentPart(expiredTmpl, err.Error())
		default:
			notfound.Serve(rcv.Controller)
			return
		}

	}

	if err := rcv.get(); err == nil {
		return
	}

	rcv.Render()

}

func New() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		c := &controller{base.New(rw, r, false, "controller/account"), nil}
		c.serve()
	})
}
