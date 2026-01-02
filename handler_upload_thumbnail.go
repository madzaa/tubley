package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
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
	allowedTypes := []string{"image/jpeg", "image/png"}
	found := slices.Contains(allowedTypes, mediaType)
	if !found {
		respondWithError(w, http.StatusInternalServerError, "Invalid Content-Type", err)
		return
	}
	extension, ok := strings.CutPrefix(mediaType, "image/")
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Unable to get extension", err)
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

	thnPath := filepath.Join(cfg.assetsRoot, videoIDString) + "." + extension
	thnUrl := fmt.Sprintf("http://localhost:8091/%s/%s.%s", cfg.assetsRoot, videoID.String(), extension)

	err = os.WriteFile(thnPath, imgData, 0666)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable save file", err)
		return
	}

	metadata.ThumbnailURL = &thnUrl

	metadata.UpdatedAt = time.Now()
	err = cfg.db.UpdateVideo(metadata)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to update video", err)
		return
	}
	respondWithJSON(w, http.StatusOK, metadata)
}
