package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/smmmfrd/bootdev-chirpy/internal/database"
)

type apiConfig struct {
	fileServerHits atomic.Int32
	queries        *database.Queries
	platform       string
}

// Run with: go build -o out && ./out
func main() {
	godotenv.Load()

	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")

	db, err := sql.Open("postgres", dbURL)

	if err != nil {
		fmt.Println("Error accessing database")
		return
	}

	dbQueries := database.New(db)

	const port = "8080"

	cfg := apiConfig{
		fileServerHits: atomic.Int32{},
		queries:        dbQueries,
		platform:       platform,
	}

	mux := http.NewServeMux()
	mux.Handle("/app/", cfg.middlewareMetrics(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))

	mux.HandleFunc("GET /api/healthz", healthz)
	mux.HandleFunc("POST /api/chirps", cfg.CreateChirp)
	mux.HandleFunc("GET /api/chirps", cfg.GetAllChirps)
	mux.HandleFunc("POST /api/users", cfg.CreateUser)

	mux.HandleFunc("POST /admin/reset", cfg.reset)
	mux.HandleFunc("GET /admin/metrics", cfg.metrics)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Starting server on port %v\n", port)

	log.Fatal(server.ListenAndServe())
}

func (cfg *apiConfig) middlewareMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileServerHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) metrics(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-type", "text/html")
	w.WriteHeader(http.StatusOK)

	w.Write([]byte(fmt.Sprintf("<html>\n<body>\n<h1>Welcome, Chirpy Admin</h1>\n<p>Chirpy has been visited %d times!</p>\n</body>\n</html>", cfg.fileServerHits.Load())))
}

func (cfg *apiConfig) reset(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		respondWithError(w, http.StatusForbidden, "Reset is only allowed in dev environment")
		return
	}

	cfg.fileServerHits.Store(0)

	err := cfg.queries.Reset(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to reset the database")
	}

	w.Header().Set("Content-type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	w.Write([]byte("Hits reset to 0 and database reset"))
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	w.Write([]byte("OK"))
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	type responseError struct {
		Error string `json:"error"`
	}

	log.Print(msg)
	w.WriteHeader(code)

	resErr := responseError{
		Error: fmt.Sprint(msg),
	}

	data, err := json.Marshal(resErr)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.WriteHeader(code)

	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}
