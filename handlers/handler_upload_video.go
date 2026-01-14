package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/madza/tubley/config"
	"github.com/madza/tubley/internal/auth"
	"github.com/madza/tubley/internal/database"
)

func HandlerUploadVideo(cfg *config.ApiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		r.Body = http.MaxBytesReader(w, r.Body, 1<<30)
		videoPath := r.PathValue("videoID")
		videoId, err := uuid.Parse(videoPath)
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

		metadata, err := cfg.Db.GetVideo(videoId)
		if err != nil || metadata.UserID != userID {
			if err != nil {
				RespondWithError(w, http.StatusUnauthorized, "Couldn't find video", err)
				return
			}
		}
		file, header, err := r.FormFile("video")
		if err != nil {
			RespondWithError(w, http.StatusUnauthorized, "Couldn't parse video file", err)
			return
		}
		defer file.Close()
		mediaType, _, err := mime.ParseMediaType(header.Header.Get("Content-Type"))
		if err != nil && mediaType != "video/mp4" {
			RespondWithError(w, http.StatusUnauthorized, "Unable to  parse media type", err)
			return
		}
		temp, err := os.CreateTemp("", "tubley-temp.mp4")
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Unable to create temp file for upload", err)
			return
		}
		defer os.Remove(temp.Name())
		defer temp.Close()
		_, err = io.Copy(temp, file)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Unable to copy contents to file for upload", err)
			return
		}
		key := make([]byte, 32)
		rand.Read(key)
		encodedID := base64.RawURLEncoding.EncodeToString(key) + ".mp4"

		_, err = temp.Seek(0, io.SeekStart)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Unable to seek temp file", err)
			return
		}
		s3Input := s3.PutObjectInput{
			Bucket:      &cfg.S3Bucket,
			Key:         &encodedID,
			Body:        temp,
			ContentType: &mediaType,
		}
		_, err = cfg.S3Client.PutObject(r.Context(), &s3Input)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Unable to upload", err)
			return
		}
		url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", cfg.S3Bucket, cfg.S3Region, encodedID)
		err = cfg.Db.UpdateVideo(
			database.Video{
				ID:                metadata.ID,
				CreatedAt:         metadata.CreatedAt,
				UpdatedAt:         time.Now(),
				ThumbnailURL:      metadata.ThumbnailURL,
				VideoURL:          &url,
				CreateVideoParams: metadata.CreateVideoParams,
			})
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, "Unable to update video", err)
			return
		}
	}
}
