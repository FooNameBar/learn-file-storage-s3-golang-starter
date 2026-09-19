package main

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
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

	const maxMemory = 10 << 20

	err = r.ParseMultipartForm(maxMemory)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ParseMultipartForm", err)
		return
	}

	imgFile, fileHeader, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Missing thumbnail", err)
		return
	}

	mediaTypeStr := fileHeader.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(mediaTypeStr)
	if err != nil || mediaType != "image/jpeg" || mediaType != "image/png" {
		respondWithError(w, http.StatusBadRequest, "Wrong media type", err)
		return
	}

	extType := strings.TrimPrefix(mediaTypeStr, "image/")

	imgData, err := io.ReadAll(imgFile)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "File error", err)
		return
	}

	vMetadata, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error", err)
		return
	}

	if userID != vMetadata.UserID {
		respondWithError(w, http.StatusForbidden, "Wrong user id", err)
		return
	}

	relImgPath := filepath.Join(cfg.assetsRoot, fmt.Sprintf("%s.%s", videoIDString, extType))
	file, err := os.Create(relImgPath)
	if err != nil {
	}

	n, err := io.Copy(file, bytes.NewBuffer(imgData))
	if err != nil || n != int64(len(imgData)) {
	}

	fullImgPath := fmt.Sprintf("http://localhost:%s/%s", cfg.port, relImgPath)
	vMetadata.ThumbnailURL = &fullImgPath
	err = cfg.db.UpdateVideo(vMetadata)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Updating video failed", err)
	}


	respondWithJSON(w, http.StatusOK, vMetadata)
}
