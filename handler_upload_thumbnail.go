package main

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/madza/tubley/internal/auth"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	// TODO: implement the upload here
	var maxMemory int64 = 10 << 20
	err = r.ParseMultipartForm(maxMemory)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't parse form", err)
		return
	}
	data, header, err := r.FormFile("thumbnail")
	mediaType := header.Header.Get("Content-Type")
	if mediaType == "" {
		respondWithError(w, http.StatusInternalServerError, "Unable to  get content-type", err)
		return
	}
	imgData, err := io.ReadAll(data)
	metadata, err := cfg.db.GetVideo(videoID)
	if metadata.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "User not authorized to access the video", err)
		return
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to get video", err)
		return
	}
	thn := thumbnail{
		data:      imgData,
		mediaType: mediaType,
	}
	videoThumbnails[videoID] = thn

	thnUrl := "http://localhost:8091/api/thumbnails/" + videoID.String()
	metadata.ThumbnailURL = &thnUrl
	metadata.UpdatedAt = time.Now()
	err = cfg.db.UpdateVideo(metadata)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to update video", err)
		return
	}
	respondWithJSON(w, http.StatusOK, metadata)
}
