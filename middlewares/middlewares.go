package middlewares

import (
	"authsys/middlewares/auth"
	"authsys/middlewares/httphead"
	"authsys/middlewares/security"
	"authsys/middlewares/session"
	"github.com/codegangsta/negroni"
)

// Return all middlwares
func New() []negroni.Handler {
	return []negroni.Handler{session.New(), httphead.New(), security.New(), auth.New()}
}
