package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/smmmfrd/bootdev-chirpy/internal/database"
)

type apiConfig struct {
	fileServerHits atomic.Int32
	queries        *database.Queries
}

// Run with: go build -o out && ./out
func main() {
	godotenv.Load()

	dbURL := os.Getenv("DB_URL")

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
	}

	mux := http.NewServeMux()
	mux.Handle("/app/", cfg.middlewareMetrics(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))

	mux.HandleFunc("GET /api/healthz", healthz)
	mux.HandleFunc("POST /api/validate_chirp", validate_chirp)
	mux.HandleFunc("POST /api/users", cfg.createUser)

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

func (cfg *apiConfig) reset(w http.ResponseWriter, _ *http.Request) {
	cfg.fileServerHits.Store(0)
	w.Header().Set("Content-type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	w.Write([]byte("Hits reset to 0"))
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	w.Write([]byte("OK"))
}

type chirp struct {
	Body string `json:"body"`
}

type response struct {
	CleanedBody string `json:"cleaned_body"`
}

func validate_chirp(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	c := chirp{}

	if err := decoder.Decode(&c); err != nil {
		respondWithError(w, 500, fmt.Sprintf("Error decoding parameters: %s", err))
		return
	}

	if len(c.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}

	c.Body = badWordReplacement(c.Body)

	res := response{
		CleanedBody: c.Body,
	}

	respondWithJSON(w, 200, res)
}

func badWordReplacement(c string) string {
	split := strings.Split(c, " ")

	for i, str := range split {
		if slices.Contains([]string{"kerfuffle", "sharbert", "fornax"}, strings.ToLower(str)) {
			split[i] = "****"
		}
	}

	return strings.Join(split, " ")
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

func (cfg *apiConfig) createUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)

	e := parameters{}

	if err := decoder.Decode(&e); err != nil {
		respondWithError(w, 500, fmt.Sprintf("Error decoding parameters: %s", err))
		return
	}

	user, err := cfg.queries.CreateUser(r.Context(), e.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create user")
		return
	}

	respondWithJSON(w, 200, user)
}
