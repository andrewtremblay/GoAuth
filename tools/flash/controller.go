package flash

/*
 * Read stored flash message from redis and generated html
 * output. It should use only in combination with redirect.
 * Session will be save in different types like E => Error
 * I => Information and W => Warning
 */

import (
	"authsys/middlewares/session"
	//"fmt"
	"net/http"
)

// Save flash message into database
func Set(r *http.Request, option, msg string) {

	// Validate if the option is valid or not
	if option != "I" && option != "E" && option != "W" {
		panic("Flash option does not exist.")
	}

	session.InsertData(r, option, msg)

}

// Read flash message from database and delete after read
func Get(r *http.Request, option string) string {
	msg, err := session.ReadData(r, option)
	if err != nil {
		return ""
	}
	session.DeleteData(r, option)
	return msg.(string)
}
