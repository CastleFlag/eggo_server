package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		log.Println("request")
		w.Write([]byte("hello wolrd\n"))
	})
	http.ListenAndServe(":8000", nil)
}
