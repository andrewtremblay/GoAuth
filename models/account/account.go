package account

import (
	"authsys/tools/caller"
	"authsys/tools/i18n"
	"code.google.com/p/go.crypto/bcrypt"
	"errors"
	"fmt"
	"github.com/jmcvetta/neoism"
	"io/ioutil"
	//"os"
	//"path/filepath"
	"path"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode"
)

const (
	i18nSec string = "models/account"
)

type SignedUser struct {
	Name, Email string
}

/*
 * Types
 */
type account struct {
	name           string
	email          string
	hashedPassword string
	createdAt      int64 //Time will be save as unix time in database
	termOf         bool
	activated      bool
	activatedAt    int64 //Time will be save as unix time in database
	closed         bool
	closedAt       int64 //Time will be save as unix time in database
	closedReason   string
	lastSignIn     int64 //Time will be save as unix time in database
	recoveryAt     int64 //Time will be save as unix time in database
}

/*
 * Constants
 */
const (
	dbUri string = "http://localhost:7474/db/data"
)

/*
 * Private variable
 */
var db *neoism.Database
var dbErr error
var prohibitedName []string
var dirtyNameFile string

func init() {
	fmt.Println("Initialization.")
	db, _ = neoism.Connect(dbUri)

	dirtyNameFile = path.Join(caller.Path(), "dirty_name.txt")
}

/*
 * Private functions
 */

func createHashPassword(password string) string {

	hashByte, _ := bcrypt.GenerateFromPassword([]byte(password), 5)
	hashedPassword := string(hashByte)
	return hashedPassword

}

func compareHashWithPassword(hashPassword, password string) bool {

	if err := bcrypt.CompareHashAndPassword([]byte(hashPassword), []byte(password)); err != nil {
		return false
	}

	return true

}

func filterDirtyName(name, local string) error {

	fileIsRead := false

	// Check if file with dirty name has already read
	for _, value := range prohibitedName {
		if value != "" {
			fileIsRead = true
		}
	}

	if !fileIsRead {
		contents, err := ioutil.ReadFile(dirtyNameFile)
		if err != nil {
			panic("Could not find file for dirtyname validation.")
		}

		prohibitedName = strings.Split(string(contents[:]), "\n")
	}

	for _, value := range prohibitedName {

		if strings.Contains(name, value) {
			return errors.New(i18n.Translate(local, i18nSec, "text11"))
		}
	}

	return nil

}

/*
 * Public functions
 */

/* Name rules:
 * at min 6 and max 15 character
 * will convert name to lower case
 * local is for translating, if error occurs
 */
func ValidateName(value, local string) error {

	fmt.Println("Validate name: ", value)

	pattern := regexp.MustCompile(`^[a-zA-Z0-9]{6,15}$`)
	nameLowerCase := strings.ToLower(value)

	if !pattern.MatchString(nameLowerCase) {
		return errors.New(i18n.Translate(local, i18nSec, "text01"))
	}

	if err := filterDirtyName(nameLowerCase, local); err != nil {
		return err
	}

	// Validate, if name already available in db
	res := []struct {
		Name string `json: "acc.name"`
	}{}

	cq := &neoism.CypherQuery{
		Statement: `
			MATCH (acc:Account {name: {name}})
			RETURN acc.name
		`,
		Parameters: neoism.Props{"name": nameLowerCase},
		Result:     &res,
	}

	db.Cypher(cq)
	if len(res) > 0 {
		return errors.New(i18n.Translate(local, i18nSec, "text02"))

	}

	return nil

}

func ValidateEmail(value, local string) error {

	fmt.Println("Validate email", value)

	pattern := regexp.MustCompile(`(([^<>()[\]\\.,;:\s@\"]+(\.[^<>()[\]\\.,;:\s@\"]+)*)|(\".+\"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$`)

	if !pattern.MatchString(value) {
		return errors.New(i18n.Translate(local, i18nSec, "text04"))
	}

	// Validate, if email already available in db
	result := []struct {
		Email string `json: "acc.email"`
	}{}
	cq := &neoism.CypherQuery{
		Statement: `
			MATCH (acc:Account {email: {email}})
			RETURN acc.email`,
		Parameters: neoism.Props{"email": value},
		Result:     &result,
	}
	db.Cypher(cq)

	if len(result) > 0 {
		return errors.New(i18n.Translate(local, i18nSec, "text05"))
	}

	return nil
}

/*
 * Password rules:
 * at least 7 letters
 * at least 1 number
 * at least 1 upper case
 * at least 1 special character
 */
func ValidatePassword(value, local string) error {

	fmt.Println("Validate password", value)
	if len(value) < 7 {
		return errors.New(i18n.Translate(local, i18nSec, "text03"))
	}

	var num, lower, upper, spec bool
	for _, r := range value {
		switch {
		case unicode.IsDigit(r):
			num = true
		case unicode.IsUpper(r):
			upper = true
		case unicode.IsLower(r):
			lower = true
		case unicode.IsSymbol(r), unicode.IsPunct(r):
			spec = true
		}
	}
	if num && lower && upper && spec {
		return nil
	}

	return errors.New(i18n.Translate(local, i18nSec, "text03"))
}

/*
 * Account rules:
 * name will be saved lower case in database
 */
func Create(name, email, password, local string, termOf bool) []error {

	wait := new(sync.WaitGroup)
	mutex := new(sync.Mutex)
	errs := make([]error, 3)

	if !termOf {
		errs = append(errs, errors.New(i18n.Translate(local, i18nSec, "text06")))
	}

	wait.Add(1)
	go func() {
		if err := ValidateName(name, local); err != nil {
			mutex.Lock()
			errs = append(errs, err)
			mutex.Unlock()
		}
		wait.Done()
	}()

	wait.Add(1)
	go func() {
		if err := ValidateEmail(email, local); err != nil {
			mutex.Lock()
			errs = append(errs, err)
			mutex.Unlock()
		}
		wait.Done()
	}()

	wait.Add(1)
	go func() {
		if err := ValidatePassword(password, local); err != nil {
			mutex.Lock()
			errs = append(errs, err)
			mutex.Unlock()
		}
		wait.Done()
	}()

	wait.Wait()

	// If errors appear
	if len(errs) > 0 {
		return errs
	}

	newAccount := new(account)
	newAccount.name = strings.ToLower(name)
	newAccount.email = email
	newAccount.hashedPassword = createHashPassword(password)
	newAccount.createdAt = time.Now().Unix()
	newAccount.termOf = termOf

	res := []struct {
		N neoism.Node
	}{}

	cp := &neoism.CypherQuery{
		Statement: `
			CREATE (n:Account {name: {name}, email: {email}, hashed_password: {hashed_password}, term_of: {term_of},
							   created_at: {created_at}, activated: {activated}, activated_at: {activated_at},
							   closed: {closed}, closed_at: {closed_at}, closed_reason: {closed_reason},
							   last_signin: {last_signin}, recovery_at: {recovery_at}
					})
			RETURN n
			`,
		Parameters: neoism.Props{
			"name":            newAccount.name,
			"email":           newAccount.email,
			"hashed_password": newAccount.hashedPassword,
			"created_at":      newAccount.createdAt,
			"term_of":         newAccount.termOf,
			"activated":       newAccount.activated,
			"activated_at":    newAccount.activatedAt,
			"closed":          newAccount.closed,
			"closed_at":       newAccount.closedAt,
			"closed_reason":   newAccount.closedReason,
			"last_signin":     newAccount.lastSignIn,
			"recovery_at":     newAccount.recoveryAt,
		},
		Result: &res,
	}

	if err := db.Cypher(cp); err != nil {
		return []error{err}
	}

	return nil
}

func Delete(email, local string) error {

	_, err := Read(email, local)

	if err != nil {
		return err
	}

	cp := &neoism.CypherQuery{
		Statement: `
			MATCH (n:Account{ email: {email} })
			DELETE n
		`,
		Parameters: neoism.Props{
			"email": email,
		},
	}

	if err := db.Cypher(cp); err != nil {
		return err
	}

	return nil

}

func Read(email, local string) (*SignedUser, error) {

	res := []struct {
		Name  string `json:"n.name"`
		Email string `json:"n.email"`
	}{}

	cp := &neoism.CypherQuery{
		Statement: `
			MATCH (n:Account{ email: {email} })
			RETURN n.name, n.email
		`,
		Parameters: neoism.Props{
			"email": email,
		},
		Result: &res,
	}

	if err := db.Cypher(cp); err != nil {
		return nil, err
	}

	//Could not find account
	if len(res) == 0 {
		return nil, errors.New(i18n.Translate(local, i18nSec, "text07"))
	}

	return &SignedUser{Name: res[0].Name, Email: res[0].Email}, nil

}

func Update(email, newName, newEmail, local string) []error {

	errChan := make(chan error, 2)
	errSli := []error{}
	wait := &sync.WaitGroup{}
	//index := 0

	//Check if account is available
	_, err := Read(email, local)

	if err != nil {
		return []error{err}
	}

	wait.Add(1)
	go func() {
		if err := ValidateEmail(newEmail, local); err != nil {
			errChan <- err
		}
		wait.Done()
	}()

	wait.Add(1)
	go func() {
		if err := ValidateName(newName, local); err != nil {
			errChan <- err
		}
		wait.Done()
	}()

	wait.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			errSli = append(errSli, err)
		}
	}

	if len(errSli) > 0 {
		return errSli
	}

	res := []struct {
		Name  string `json:"n.name"`
		Email string `json:"n.email"`
	}{}

	cp := &neoism.CypherQuery{
		Statement: `
			MATCH (n:Account {email: {email}})
			SET n.name = {newName} , n.email = {newEmail}
			RETURN n.name, n.email
		`,
		Parameters: neoism.Props{
			"email":    email,
			"newEmail": newEmail,
			"newName":  newName,
		},
		Result: &res,
	}

	if err := db.Cypher(cp); err != nil {
		return []error{err}
	}

	return nil
}

// Update email address and deactivate user account.
// User will receive email confirmation to activate
// account again to ensure, that the new email address
// is valid.
func UpdateEmail(oldEmail, newEmail, local string) error {

	//Check if account is available
	_, err := Read(oldEmail, local)

	if err != nil {
		return err
	}

	if err := ValidateEmail(newEmail, local); err != nil {
		return err
	}

	res := []struct {
		Name  string `json:"n.name"`
		Email string `json:"n.email"`
	}{}

	cp := &neoism.CypherQuery{
		Statement: `
			MATCH (n:Account {email: {email}})
			SET n.email = {newEmail}, n.activated = false
			RETURN n.name, n.email
		`,
		Parameters: neoism.Props{
			"email":    oldEmail,
			"newEmail": newEmail,
		},
		Result: &res,
	}

	if err := db.Cypher(cp); err != nil {
		return err
	}

	return nil

}

func UpdateName(email, newName, local string) error {

	//Check if account is available
	_, err := Read(email, local)

	if err != nil {
		return err
	}

	if err := ValidateName(newName, local); err != nil {
		return err
	}

	res := []struct {
		Name  string `json:"n.name"`
		Email string `json:"n.email"`
	}{}

	cp := &neoism.CypherQuery{
		Statement: `
			MATCH (n:Account {email: {email}})
			SET n.name = {name}
			RETURN n.name, n.email
		`,
		Parameters: neoism.Props{
			"email": email,
			"name":  newName,
		},
		Result: &res,
	}

	if err := db.Cypher(cp); err != nil {
		return err
	}

	return nil

}

// Activate account, after user confirm email
func Activate(email, local string) error {

	//Check if account is available
	if _, err := Read(email, local); err != nil {
		return err
	}

	//Check if account is already activated
	res := []struct {
		Email     string `json:"n.email"`
		Activated bool   `json:"n.activated,bool"`
	}{}

	cp := &neoism.CypherQuery{
		Statement: `
			MATCH (n:Account {email: {email}})
			RETURN n.email, n.activated
		`,

		Parameters: neoism.Props{
			"email": email,
		},
		Result: &res,
	}

	if err := db.Cypher(cp); err != nil {
		return err
	}

	if len(res) > 0 && res[0].Activated {
		return errors.New(i18n.Translate(local, i18nSec, "text08"))
	}
	//Activate account and the activation date
	cp = &neoism.CypherQuery{
		Statement: `
			MATCH (n:Account {email: {email}})
			SET n.activated = true, n.activated_at = {activated_at}
			RETURN n.email, n.activated
		`,

		Parameters: neoism.Props{
			"email":        email,
			"activated_at": time.Now().Unix(),
		},
		Result: &res,
	}

	if err := db.Cypher(cp); err != nil {
		return err
	}

	if !res[0].Activated {
		return errors.New(i18n.Translate(local, i18nSec, "text09"))
	}

	return nil
}

// Close account, if user did some prohibited
func Close(email, reason, local string) error {

	//Check if account is available
	if _, err := Read(email, local); err != nil {
		return err
	}

	res := []struct {
		Email  string `json:"n.email"`
		Closed bool   `json:"n.closed,bool"`
	}{}

	cp := &neoism.CypherQuery{
		Statement: `
			MATCH (n:Account {email: {email}})
			SET n.closed = true, n.closed_at = {closed_at}, n.closed_reason = {closed_reason}
			RETURN n.email, n.closed
		`,

		Parameters: neoism.Props{
			"email":         email,
			"closed_at":     time.Now().Unix(),
			"closed_reason": reason,
		},
		Result: &res,
	}

	if err := db.Cypher(cp); err != nil {
		return err
	}

	if !res[0].Closed {
		return errors.New(i18n.Translate(local, "models/account", "text10"))
	}

	return nil

}

// Write last sign in from user
func SignIn(email, password, local string) (*SignedUser, error) {

	//Check if account is available
	if _, err := Read(email, local); err != nil {
		return nil, err
	}

	res := []struct {
		Name           string `json:"n.name"`
		Email          string `json:"n.email"`
		HashedPassword string `json:"n.hashed_password"`
		Activated      bool   `json:"n.activated,bool"`
	}{}

	cp := &neoism.CypherQuery{
		Statement: `
			MATCH (n:Account {email: {email}})
			RETURN n.name, n.email, n.hashed_password, n.activated
		`,

		Parameters: neoism.Props{
			"email": email,
		},
		Result: &res,
	}

	if err := db.Cypher(cp); err != nil {
		return nil, err
	}

	if !res[0].Activated {
		return nil, errors.New(i18n.Translate(local, "models/account", "text14"))
	}

	if !compareHashWithPassword(res[0].HashedPassword, password) {
		return nil, errors.New(i18n.Translate(local, "models/account", "text13"))
	}

	// Set sign in date
	cp = &neoism.CypherQuery{
		Statement: `
			MATCH (n:Account {email: {email}})
			SET n.last_signin = {last_signin}
		`,
		Parameters: neoism.Props{
			"email":       email,
			"last_signin": time.Now().Unix(),
		},
	}

	if err := db.Cypher(cp); err != nil {
		return nil, err
	}

	return &SignedUser{Name: res[0].Name, Email: res[0].Email}, nil
}

func ChangePassword(email, oldPassword, newPassword, local string) error {

	//Check if account is available
	if _, err := Read(email, local); err != nil {
		return err
	}

	res := []struct {
		Email          string `json:"n.email"`
		HashedPassword string `json:"n.hashed_password"`
	}{}

	cp := &neoism.CypherQuery{
		Statement: `
			MATCH (n:Account {email: {email}})
			RETURN n.email, n.hashed_password
		`,

		Parameters: neoism.Props{
			"email": email,
		},
		Result: &res,
	}

	if err := db.Cypher(cp); err != nil {
		return err
	}

	if !compareHashWithPassword(res[0].HashedPassword, oldPassword) {
		return errors.New(i18n.Translate(local, "models/account", "text12"))
	}

	//Testing security requirements for new Password
	if err := ValidatePassword(newPassword, local); err != nil {
		return err
	}

	cp = &neoism.CypherQuery{
		Statement: `
			MATCH (n:Account {email: {email}})
			SET n.hashed_password = {hashed_password}
		`,

		Parameters: neoism.Props{
			"email":           email,
			"hashed_password": createHashPassword(newPassword),
		},
	}

	if err := db.Cypher(cp); err != nil {
		return err
	}

	return nil
}

func ResetPassword(email, newPassword, local string) error {

	//Check if account is available
	if _, err := Read(email, local); err != nil {
		return err
	}

	if err := ValidatePassword(newPassword, local); err != nil {
		return err
	}

	cp := &neoism.CypherQuery{
		Statement: `
			MATCH (n:Account {email: {email}})
			SET n.hashed_password = {hashed_password}
		`,

		Parameters: neoism.Props{
			"email":           email,
			"hashed_password": createHashPassword(newPassword),
		},
	}

	if err := db.Cypher(cp); err != nil {
		return err
	}

	return nil

}
