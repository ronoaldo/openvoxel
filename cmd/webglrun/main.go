package main

import (
	"log"
	"net/http"
)

func main() {
	log.Print("Starting server for wasmrun ...")
	http.Handle("/", http.FileServer(http.Dir("./")))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
