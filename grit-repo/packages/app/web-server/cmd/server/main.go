package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello from web-server!")
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK")
	})

	fmt.Println("Starting web server on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}