package main

import (
	"fmt"
	"log"
	"net/http"
)

func hello(w http.ResponseWriter, r *http.Request) {
	_, err := fmt.Fprintf(w, "Hello %s\n", r.URL.Query().Get("name"))
	if err != nil {
		log.Fatalln(err)
	}
}

func main() {
	http.HandleFunc("/hello", hello)
	http.ListenAndServe(":8000", nil)
}
