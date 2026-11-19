package main

import (
	"fmt"
	"io"
	"log"
	"net/http"

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
	r.ParseMultipartForm(maxMemory)

	// "thumbnail" should match the HTML form input name
	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}
	defer file.Close()

	mediaType := header.Header.Get("Content-Type")
	image_data, err := io.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}

	metadata, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to retreive the data", err)
		return
	}
	if userID != metadata.UserID {
		respondWithError(w, http.StatusUnauthorized, "Not allowed", nil)
		return
	}

	thumbnail := thumbnail{
		data:      image_data,
		mediaType: mediaType,
	}

	videoThumbnails[videoID] = thumbnail
	thumbURL := fmt.Sprintf("http://localhost:8091/api/thumbnails/%s", videoID)
	metadata.ThumbnailURL = &thumbURL

	err = cfg.db.UpdateVideo(metadata)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to update the video", err)
		return
	}

	respondWithJSON(w, http.StatusOK, metadata)
}
