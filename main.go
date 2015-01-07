package main

import (
	"authsys/controllers"
	"authsys/middlewares"
	"bytes"
	"fmt"
	"github.com/codegangsta/negroni"
	//"github.com/stretchr/graceful"
	"log"
	"runtime"
	//"time"
)

var (
	buf    bytes.Buffer
	logger *log.Logger = log.New(&buf, "log: ", log.Lshortfile)
)

func main() {

	logger.Println("Application starts.")
	// User all cpu cores
	runtime.GOMAXPROCS(runtime.NumCPU())

	n := negroni.Classic()
	for _, v := range middlewares.New() {
		n.Use(v)
	}
	n.UseHandler(controllers.Routes())

	// Run the application and listen the port
	fmt.Println(&buf)
	n.Run(":3000")
	//graceful.Run(":3000", 5*time.Second, n)

}
