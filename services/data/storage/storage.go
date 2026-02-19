package storage

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/storage"
)

var (
	projectID            = os.Getenv("FIREBASE_PROJECT_ID")
	bucketName           = os.Getenv("STORAGE_BUCKET_NAME")
	outputDir            = "outputs" // TODO Chnage this name
	recordingSessionsDir = "recordingSessions"
)

type StorageManager struct {
	Client       *storage.Client
	BucketHandle *storage.BucketHandle
}

var Storage = initStorage()

func initStorage() *StorageManager {
	client, err := storage.NewClient(context.Background())
	if err != nil {
		log.Fatal(errors.New("storage client: " + err.Error()))
		return nil
	}

	return &StorageManager{
		Client:       client,
		BucketHandle: client.Bucket(bucketName),
	}
}

func FinalAudioUploadPath(userId, sessionId string) string {
	return fmt.Sprintf("%s/%s/%s/final.wav", outputDir, userId, sessionId)
}
