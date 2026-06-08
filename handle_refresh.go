package main

import (
	"net/http"
	"time"

	"github.com/smmmfrd/bootdev-chirpy/internal/auth"
)

func (cfg *apiConfig) handleRefresh(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Token string `json:"token"`
	}

	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't find refresh token")
		return
	}

	user, err := cfg.queries.GetUserRefreshToken(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't get a user from refresh token")
		return
	}

	accessToken, err := auth.MakeJWT(
		user.ID,
		cfg.authSecret,
		time.Hour,
	)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate token")
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		Token: accessToken,
	})
}

func (cfg *apiConfig) handleRevoke(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't find token in request")
		return
	}

	_, err = cfg.queries.RevokeRefreshToken(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't revoke session")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
