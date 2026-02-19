package storage

import (
	"context"
	"eavesdropper/dtos/resources"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

func LoadManifest(ctx context.Context, sessionId string, userId string) (*resources.RecordingManifest, error) {

	path := fmt.Sprintf("%s/%s/%s/manifest.json", recordingSessionsDir, userId, sessionId)

	reader, err := Storage.BucketHandle.Object(path).NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("open %s in %s: %w", path, Storage.BucketHandle.BucketName(), err)
	}
	defer reader.Close()

	var m resources.RecordingManifest
	if err := json.NewDecoder(reader).Decode(&m); err != nil {
		return nil, fmt.Errorf("decode manifest: %w", err)
	}
	return &m, nil
}

func Upload(ctx context.Context, storagePath, localPath, contentType string) error {
	file, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer file.Close()

	bucketWritter := Storage.BucketHandle.Object(storagePath).NewWriter(ctx)
	bucketWritter.ChunkSize = 4 * 1024 * 1024 // This makes the post to storage be split into chunks of this size and makes it more resilient to network failures
	bucketWritter.ContentType = contentType
	if _, err := io.Copy(bucketWritter, file); err != nil {
		_ = bucketWritter.Close()
		return err
	}
	return bucketWritter.Close()
}

func Download(ctx context.Context, storagePath, localPath string) error {
	reader, err := Storage.BucketHandle.Object(storagePath).NewReader(ctx)
	if err != nil {
		return err
	}
	defer reader.Close()

	f, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, reader)
	return err
}
