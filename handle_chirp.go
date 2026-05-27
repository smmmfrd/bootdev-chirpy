package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/smmmfrd/bootdev-chirpy/internal/database"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) CreateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body   string    `json:"body"`
		UserId uuid.UUID `json:"user_id"`
	}

	type response struct {
		Chirp
	}

	decoder := json.NewDecoder(r.Body)
	c := parameters{}

	if err := decoder.Decode(&c); err != nil {
		respondWithError(w, 500, fmt.Sprintf("Error decoding parameters: %s", err))
		return
	}

	if len(c.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}

	c.Body = badWordReplacement(c.Body)

	res, err := cfg.queries.CreateChirp(r.Context(), database.CreateChirpParams{Body: c.Body, UserID: c.UserId})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating chirp")
		return
	}

	respondWithJSON(w, 201, response{
		Chirp: Chirp{
			ID:        res.ID,
			CreatedAt: res.CreatedAt,
			UpdatedAt: res.UpdatedAt,
			Body:      res.Body,
			UserID:    res.UserID,
		},
	})
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
