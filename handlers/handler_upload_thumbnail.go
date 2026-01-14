package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/madza/tubley/config"
	"github.com/madza/tubley/internal/auth"
)

func HandlerUploadThumbnail(cfg *config.ApiConfig) http.HandlerFunc {
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

		fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

		var maxMemory int64 = 10 << 20
		err = r.ParseMultipartForm(maxMemory)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Couldn't parse form", err)
			return
		}
		data, header, err := r.FormFile("thumbnail")
		mediaType := header.Header.Get("Content-Type")
		allowedTypes := []string{"image/jpeg", "image/png"}
		found := slices.Contains(allowedTypes, mediaType)
		if !found {
			RespondWithError(w, http.StatusInternalServerError, "Invalid Content-Type", err)
			return
		}
		extension, ok := strings.CutPrefix(mediaType, "image/")
		if !ok {
			RespondWithError(w, http.StatusInternalServerError, "Unable to get extension", err)
			return
		}

		imgData, err := io.ReadAll(data)
		metadata, err := cfg.Db.GetVideo(videoID)
		if metadata.UserID != userID {
			RespondWithError(w, http.StatusUnauthorized, "User not authorized to access the video", err)
			return
		}
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Unable to get video", err)
			return
		}
		key := make([]byte, 32)
		rand.Read(key)
		encodedID := base64.RawURLEncoding.EncodeToString(key)

		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Unable create video ID", err)
			return
		}
		thnPath := filepath.Join(cfg.AssetsRoot, encodedID) + "." + extension
		thnUrl := fmt.Sprintf("http://localhost:8091/%s/%s.%s", cfg.AssetsRoot, encodedID, extension)

		err = os.WriteFile(thnPath, imgData, 0666)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Unable save file", err)
			return
		}

		metadata.ThumbnailURL = &thnUrl

		metadata.UpdatedAt = time.Now()
		err = cfg.Db.UpdateVideo(metadata)
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, "Unable to update video", err)
			return
		}
		RespondWithJSON(w, http.StatusOK, metadata)
	}
}
