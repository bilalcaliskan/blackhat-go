package main

import (
	"fmt"
	"log"
	"net/http"
)

type router struct {
}

func (r *router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.URL.Path {
	case "/a":
		fmt.Fprint(w, "Executing /a")
	case "/b":
		fmt.Fprint(w, "Executing /b")
	case "/c":
		fmt.Fprint(w, "Executing /c")
	default:
		http.Error(w, "404 Not Found", 404)
	}
}

func main() {
	var r router
	log.Fatalln(http.ListenAndServe(":8000", &r))
}
