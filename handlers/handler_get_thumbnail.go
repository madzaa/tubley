package handlers

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/madza/tubley/config"
)

func HandlerThumbnailGet(cfg *config.ApiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		videoIDString := r.PathValue("videoID")
		videoID, err := uuid.Parse(videoIDString)
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, "Invalid video ID", err)
			return
		}

		tn, ok := config.VideoThumbnails[videoID]
		if !ok {
			RespondWithError(w, http.StatusNotFound, "Thumbnail not found", nil)
			return
		}

		w.Header().Set("Content-Type", tn.MediaType)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(tn.Data)))

		_, err = w.Write(tn.Data)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Error writing response", err)
			return
		}
	}
}
