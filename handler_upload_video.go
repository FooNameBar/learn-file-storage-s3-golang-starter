package main

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {
	MAX_BYTES := int64(1 << 30)

	videoIDStr := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't validate JWT", err)
		return
	}

	fmt.Println("uploading video with ID", videoID)

	metadata, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Video not found", err)
	}

	if metadata.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Not the video owner", err)
		return
	}

	file, fileHeader, err := r.FormFile("video")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Missing video file", err)
		return
	}
	defer file.Close()
	limitedFile := http.MaxBytesReader(w, file, MAX_BYTES)
	defer limitedFile.Close()

	mediaTypeStr := fileHeader.Header.Get("Content-type")
	mediaType, _, err := mime.ParseMediaType(mediaTypeStr)
	if err != nil || mediaType != "video/mp4" {
		respondWithError(w, http.StatusBadRequest, "Wrong media type", err)
		return
	}

	filename := "tubely-upload.mp4"
	tmpFile, err := os.CreateTemp("", filename)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Creating temp file error", err)
		return
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	written, err := io.Copy(tmpFile, limitedFile)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Copying file error", err)
		return
	}

	_, err = tmpFile.Seek(0, io.SeekStart)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Resetting file pointer error", err)
		return
	}

	aspect, err := getVideoAspectRatio(tmpFile.Name())
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "File error", err)
		return
	}

	keyVar := make([]byte, 32)
	_, err = rand.Read(keyVar)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Generating random key error", err)
	}
	key := base32.HexEncoding.EncodeToString(keyVar)
	key = path.Join(aspect, key)

	_, err = cfg.s3Client.PutObject(r.Context(), &s3.PutObjectInput{
		Bucket:      &cfg.s3Bucket,
		Key:         &key,
		Body:        tmpFile,
		ContentType: &mediaType,
		ContentLength: &written,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Storing video on server error", err)
		return
	}

	url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", cfg.s3Bucket, cfg.s3Region, key)
	metadata.VideoURL = &url
	err = cfg.db.UpdateVideo(metadata)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Updating video infor error", err)
		return
	}
}
