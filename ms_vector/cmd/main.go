package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello from the ms_vector service!")
	})

	port := ":8080"
	fmt.Printf("ms_vector service starting on port %s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
