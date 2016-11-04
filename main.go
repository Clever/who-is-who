package main

import (
	"flag"
	"log"

	"github.com/Clever/who-is-who/gen-go/server"
)

// MyController implements server.Controller
type MyController struct{}

func main() {
	addr := flag.String("addr", ":8080", "Address to listen at")
	flag.Parse()

	myController := MyController{}
	s := server.New(myController, *addr)
	// Serve should not return
	log.Fatal(s.Serve())
}
