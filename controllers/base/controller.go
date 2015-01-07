package base

import (
	"authsys/middlewares/auth"
	"authsys/middlewares/httphead"
	"authsys/models/account"
	"authsys/mux"
	"authsys/tools/caller"
	"authsys/tools/flash"
	"authsys/tools/i18n"
	"bytes"
	//"fmt"
	gomux "github.com/gorilla/mux"
	gohtml "html/template"
	"io/ioutil"
	"net/http"
	"path"
	"sync"
	gotext "text/template"
)

var (
	layoutCache string
	headCache   string
	bodyCache   string
	errorCache  string
	infoCache   string
	authenCache string
)

func init() {

	headCache = cacheHtmlTmpl(path.Join(caller.Path(), "views/head.html"))
	bodyCache = cacheHtmlTmpl(path.Join(caller.Path(), "views/body.html"))
	layoutCache = cacheHtmlTmpl(path.Join(caller.Path(), "views/layout.html"))
	authenCache = cacheHtmlTmpl(path.Join(caller.Path(), "views/authen.html"))

	// Messenger templates
	errorCache = cacheHtmlTmpl(path.Join(caller.Path(), "views/errors.html"))
	infoCache = cacheHtmlTmpl(path.Join(caller.Path(), "views/infos.html"))

}

func cacheHtmlTmpl(filename string) string {

	file, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err.Error())
	}

	return string(file)
}

type Controller struct {
	Request   *http.Request
	Response  http.ResponseWriter
	Router    *gomux.Router
	AuthPanel bool
	Local     string
	section   string

	// HTML parts
	title   string
	head    string
	body    string
	content string
	errors  []string
	infos   []string
}

func (rcv *Controller) renderHeaderPart() {

	if rcv.title == "" {
		panic("Title is not set.")
	}

	rcv.head = ""
	context := struct {
		Title string
	}{
		rcv.title,
	}

	buf := new(bytes.Buffer)
	tp, err := gotext.New("Head").Parse(headCache)
	if err != nil {
		panic(err)
	}

	err = tp.Execute(buf, context)
	if err != nil {
		panic(err)
	}
	rcv.head = buf.String()
}

func (rcv *Controller) renderBodyPart() {

	var errPart string
	var infoPart string

	rcv.body = ""
	bodyBuffer := new(bytes.Buffer)

	if rcv.content == "" {
		panic("Body render failed, check the content.")
	}

	// If several template exits, then merge it togheter
	if len(rcv.errors) > 0 {
		errPart = rcv.RenderErrorPart()
		rcv.errors = nil
	}

	if len(rcv.infos) > 0 {
		infoPart = rcv.RenderInfoPart()
		rcv.infos = nil
	}

	// Chaining all body content togheter
	bodyOutput := errPart + infoPart + rcv.content

	tp, err := gotext.New("Body").Parse(bodyCache)
	if err != nil {
		panic(err)
	}

	// Panel determine to show navigation bar
	data := struct {
		Auth, Body string
		AuthPanel  bool
	}{}

	data.Body = bodyOutput
	data.AuthPanel = rcv.AuthPanel
	data.Auth = rcv.renderLoggedUser()

	err = tp.Execute(bodyBuffer, data)
	if err != nil {
		panic(err)
	}

	rcv.body = bodyBuffer.String()
}

// If the user is signed in, it will
// the signed in user in the navbar
func (rcv *Controller) renderLoggedUser() string {
	buf := new(bytes.Buffer)

	tp, err := gotext.New("Logged").Parse(authenCache)
	if err != nil {
		panic(err)
	}
	//signed, user := logged.IsAuthenticated(rcv.Request, rcv.Response)
	user := auth.GetSignedInUser(rcv.Request)
	data := struct {
		//Signed bool
		User *account.SignedUser
	}{
		//signed,
		user,
	}

	err = tp.Execute(buf, data)
	if err != nil {
		panic(err)
	}

	return buf.String()
}

// Will append error message to error slice.
// Later it will render errors on content.
func (rcv *Controller) AppendError(msg error) {
	rcv.errors = append(rcv.errors, msg.Error())
}

// Will append infos message to infos slice.
func (rcv *Controller) AppendInfo(msg string) {
	rcv.infos = append(rcv.infos, msg)
}

func (rcv *Controller) RenderErrorPart() string {

	buf := new(bytes.Buffer)

	tp, err := gotext.New("Error").Parse(errorCache)
	if err != nil {
		panic(err)
	}

	err = tp.Execute(buf, rcv.errors)
	if err != nil {
		panic(err)
	}

	return buf.String()
}

func (rcv *Controller) RenderInfoPart() string {

	buf := new(bytes.Buffer)

	tp, err := gotext.New("Info").Parse(infoCache)
	if err != nil {
		panic(err)
	}

	err = tp.Execute(buf, rcv.infos)
	if err != nil {
		panic(err)
	}
	return buf.String()
}

// Render the view for controller
func (rcv *Controller) RenderContentPart(tpml string, params interface{}) {

	rcv.content = ""
	buf := new(bytes.Buffer)

	funcMap := gotext.FuncMap{
		// The name "title" is what the function will be called in the template text.
		"SetHtmlText": rcv.setHtmlText,
	}

	tp, err := gotext.New("Content").Funcs(funcMap).Parse(cacheHtmlTmpl(tpml))
	if err != nil {
		panic(err)
	}

	err = tp.Execute(buf, params)
	if err != nil {
		panic(err)
	}

	rcv.content = buf.String()
}

// Final page output
func (rcv *Controller) Render() error {

	w := new(sync.WaitGroup)

	w.Add(1)
	go func() {
		rcv.renderHeaderPart()
		w.Done()
	}()

	w.Add(1)
	go func() {
		rcv.renderBodyPart()
		w.Done()
	}()

	// Wait unti concurrency are finish
	w.Wait()
	context := struct {
		Head, Body gohtml.HTML
	}{}

	context.Head = gohtml.HTML(rcv.head)
	context.Body = gohtml.HTML(rcv.body)

	tp, err := gotext.New("HTML").Parse(layoutCache)
	if err != nil {
		return err
	}

	err = tp.Execute(rcv.Response, context)
	if err != nil {
		return err
	}

	return nil
}

func (rcv *Controller) Redirect(url string, code int) {

	http.Redirect(rcv.Response, rcv.Request, url, code)
}

// Set flash message
func (rcv *Controller) SetFlash(option, text string) {
	flash.Set(rcv.Request, option, rcv.Translate(text))
}

// Read text definition from i18n. Parameter text
// is the key, that defined in language file.
func (rcv *Controller) Translate(text string) string {
	return i18n.Translate(rcv.Local, rcv.section, text)
}

// Get flash message
func (rcv *Controller) GetFlash(option string) string {
	return flash.Get(rcv.Request, option)
}

// Set page title
func (rcv *Controller) SetTitle(text string, para ...string) {
	t := i18n.Translate(rcv.Local, "page/title", text)
	if len(para) > 0 {
		rcv.title = t + " " + para[0]
		return
	}
	rcv.title = t
}

// Translate text in html file
func (rcv *Controller) setHtmlText(section, text string) string {
	return i18n.Translate(rcv.Local, section, text)
}

func New(rw http.ResponseWriter, r *http.Request, authPanel bool, section ...string) *Controller {
	c := new(Controller)
	c.Request = r
	c.Response = rw
	c.AuthPanel = authPanel
	c.Router = mux.Router
	if len(section) > 0 {
		c.section = section[0]
	}
	c.Local = httphead.GetLang(r)
	return c
}
