package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/madza/tubley/config"
	"github.com/madza/tubley/internal/auth"
	"github.com/madza/tubley/internal/database"
)

func HandlerUsersCreate(cfg *config.ApiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type parameters struct {
			Password string `json:"password"`
			Email    string `json:"email"`
		}

		decoder := json.NewDecoder(r.Body)
		params := parameters{}
		err := decoder.Decode(&params)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
			return
		}

		if params.Password == "" || params.Email == "" {
			RespondWithError(w, http.StatusBadRequest, "Email and password are required", nil)
			return
		}

		hashedPassword, err := auth.HashPassword(params.Password)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Couldn't hash password", err)
			return
		}

		user, err := cfg.Db.CreateUser(database.CreateUserParams{
			Email:    params.Email,
			Password: hashedPassword,
		})
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Couldn't create user", err)
			return
		}

		RespondWithJSON(w, http.StatusCreated, user)
	}
}
