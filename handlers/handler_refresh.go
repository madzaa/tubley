package handlers

import (
	"net/http"
	"time"

	"github.com/madza/tubley/config"
	"github.com/madza/tubley/internal/auth"
)

func HandlerRefresh(cfg *config.ApiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type response struct {
			Token string `json:"token"`
		}

		refreshToken, err := auth.GetBearerToken(r.Header)
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, "Couldn't find token", err)
			return
		}

		user, err := cfg.Db.GetUserByRefreshToken(refreshToken)
		if err != nil {
			RespondWithError(w, http.StatusUnauthorized, "Couldn't get user for refresh token", err)
			return
		}

		accessToken, err := auth.MakeJWT(
			user.ID,
			cfg.JwtSecret,
			time.Hour,
		)
		if err != nil {
			RespondWithError(w, http.StatusUnauthorized, "Couldn't validate token", err)
			return
		}

		RespondWithJSON(w, http.StatusOK, response{
			Token: accessToken,
		})
	}
}

func HandlerRevoke(cfg *config.ApiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		refreshToken, err := auth.GetBearerToken(r.Header)
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, "Couldn't find token", err)
			return
		}

		err = cfg.Db.RevokeRefreshToken(refreshToken)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Couldn't revoke session", err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
