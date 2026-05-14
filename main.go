package main

import (
	"log"
	"net/http"
)

// Run with: go build -o out && ./out
func main() {
	const port = "8080"

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(".")))
	mux.Handle("/healthz", http.HandlerFunc(healthz))

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Starting server on port %v\n", port)

	log.Fatal(server.ListenAndServe())
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	w.Write([]byte("OK"))
}
