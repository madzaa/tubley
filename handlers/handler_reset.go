package handlers

import (
	"net/http"

	"github.com/madza/tubley/config"
)

func HandlerReset(cfg *config.ApiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if cfg.Platform != "dev" {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("Reset is only allowed in dev environment."))
			return
		}

		err := cfg.Db.Reset()
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Couldn't reset database", err)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Database reset to initial state"))
	}
}
