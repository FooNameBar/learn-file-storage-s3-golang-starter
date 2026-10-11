package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
)

func generatePresignedURL(s3Client *s3.Client, bucket, key string, expireTime time.Duration) (string, error) {
	presigned := s3.NewPresignClient(s3Client)
	req, err := presigned.PresignGetObject(context.Background(), &s3.GetObjectInput{Bucket: &bucket, Key: &key}, s3.WithPresignExpires(expireTime))
	if err != nil {
		return "", err
	}
	return req.URL, nil
}

func (cfg *apiConfig) dbVideoToSignedVideo(video database.Video) (database.Video, error) {
	if video.VideoURL == nil {
		return video, nil
	}

	urlSlc := strings.Split(*video.VideoURL, ",")
	if len(urlSlc) != 2 {
		return database.Video{}, fmt.Errorf("Invalid video url: %s\n", *video.VideoURL)
	}
	bucket, key := urlSlc[0], urlSlc[1]

	presignedUrl, err := generatePresignedURL(cfg.s3Client, bucket, key, time.Minute * 5)
	if err != nil {
	}

	signedVid := database.Video{
		ID: video.ID,
		CreatedAt: video.CreatedAt,
		UpdatedAt: video.UpdatedAt,
		ThumbnailURL: video.ThumbnailURL,
		VideoURL: &presignedUrl,
		CreateVideoParams: video.CreateVideoParams,
	}
	return signedVid, nil
}
