package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/madza/tubley/config"
	"github.com/madza/tubley/internal/auth"
	"github.com/madza/tubley/internal/database"
)

func HandlerVideoMetaCreate(cfg *config.ApiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type parameters struct {
			database.CreateVideoParams
		}

		token, err := auth.GetBearerToken(r.Header)
		if err != nil {
			RespondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
			return
		}
		userID, err := auth.ValidateJWT(token, cfg.JwtSecret)
		if err != nil {
			RespondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
			return
		}

		decoder := json.NewDecoder(r.Body)
		params := parameters{}
		err = decoder.Decode(&params)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
			return
		}
		params.UserID = userID

		video, err := cfg.Db.CreateVideo(params.CreateVideoParams)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Couldn't create video", err)
			return
		}

		RespondWithJSON(w, http.StatusCreated, video)
	}
}

func HandlerVideoMetaDelete(cfg *config.ApiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		videoIDString := r.PathValue("videoID")
		videoID, err := uuid.Parse(videoIDString)
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, "Invalid ID", err)
			return
		}

		token, err := auth.GetBearerToken(r.Header)
		if err != nil {
			RespondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
			return
		}
		userID, err := auth.ValidateJWT(token, cfg.JwtSecret)
		if err != nil {
			RespondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
			return
		}

		video, err := cfg.Db.GetVideo(videoID)
		if err != nil {
			RespondWithError(w, http.StatusNotFound, "Couldn't get video", err)
			return
		}
		if video.UserID != userID {
			RespondWithError(w, http.StatusForbidden, "You can't delete this video", err)
			return
		}

		err = cfg.Db.DeleteVideo(videoID)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Couldn't delete video", err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func HandlerVideoGet(cfg *config.ApiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		videoIDString := r.PathValue("videoID")
		videoID, err := uuid.Parse(videoIDString)
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, "Invalid video ID", err)
			return
		}

		video, err := cfg.Db.GetVideo(videoID)
		if err != nil {
			RespondWithError(w, http.StatusNotFound, "Couldn't get video", err)
			return
		}

		RespondWithJSON(w, http.StatusOK, video)
	}
}

func HandlerVideosRetrieve(cfg *config.ApiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := auth.GetBearerToken(r.Header)
		if err != nil {
			RespondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
			return
		}
		userID, err := auth.ValidateJWT(token, cfg.JwtSecret)
		if err != nil {
			RespondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
			return
		}

		videos, err := cfg.Db.GetVideos(userID)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Couldn't retrieve videos", err)
			return
		}

		RespondWithJSON(w, http.StatusOK, videos)
	}
}
