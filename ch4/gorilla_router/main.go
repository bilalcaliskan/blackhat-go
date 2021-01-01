package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/foo", func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprint(writer, "hi foo")
	}).Methods("GET").Host("www.foo.com")

	// request parameter
	r.HandleFunc("/users/{user}", func(w http.ResponseWriter, req *http.Request) {
		user := mux.Vars(req)["user"]
		fmt.Fprintf(w, "hi %s\n", user)
	}).Methods("GET")

	// request parameter with regex
	r.HandleFunc("/users-with-regex/{user:[a-z]+}", func(w http.ResponseWriter, req *http.Request) {
		user := mux.Vars(req)["user"]
		fmt.Fprintf(w, "hi %s\n", user)
	}).Methods("GET")

	http.ListenAndServe(":8000", r)
}
