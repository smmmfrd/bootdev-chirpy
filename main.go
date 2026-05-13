package main

import (
	"fmt"
	"net/http"
)

// Run with: go build -o out && ./out
func main() {
	const port = "8080"

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(".")))

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	fmt.Printf("Starting server on port %v\n", port)

	server.ListenAndServe()
}
