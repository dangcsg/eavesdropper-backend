package audio

import (
	"context"
	"eavesdropper/errs"
	cloudStorage "eavesdropper/services/data/storage"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// TODO Double check if it's not possible to pass storage audio file path to gemini instead of uploading the local to gemini
func ProcessAudioChunks(
	ctx context.Context,
	sessionID string,
	userID string,
	directory string,
) (
	finalAudioLocalPath string,
	finalAudioStoragePath string,
	audioDurationSeconds int,
	err error,
) {

	// Gets the recording session manifest json
	manifest, err := cloudStorage.LoadManifest(ctx, sessionID, userID)
	if err != nil {
		return "", "", 0, err
	}
	if manifest.Count == 0 || len(manifest.Chunks) == 0 {
		return "", "", 0, errs.NoAudioChunksInManifest
	}

	// Downloads the audio chunks locally in sorted order
	chunkNames := make([]string, len(manifest.Chunks))
	copy(chunkNames, manifest.Chunks)
	sort.Slice(chunkNames, func(i, j int) bool {
		return chunkIndex(chunkNames[i]) < chunkIndex(chunkNames[j])
	})

	localFiles := []string{}
	for _, storagePath := range chunkNames {
		local := filepath.Join(directory, filepath.Base(storagePath))
		err = cloudStorage.Download(ctx, storagePath, local)
		if err != nil {
			return "", "", 0, errors.New("download " + storagePath + ": " + err.Error())
		}
		localFiles = append(localFiles, local)
	}
	fmt.Printf("\ndownloaded %d chunks", len(localFiles))

	// Byte-append all chunks into a single webm file
	joined := filepath.Join(directory, "joined.webm")
	out, err := os.Create(joined)
	if err != nil {
		return "", "", 0, errors.New("create joined: " + err.Error())
	}
	for _, local := range localFiles {
		f, _err := os.Open(local)
		if _err != nil {
			out.Close()
			return "", "", 0, errors.New("open chunk: " + _err.Error())
		}
		_, err = io.Copy(out, f)
		if err != nil {
			f.Close()
			out.Close()
			return "", "", 0, errors.New("append chunk: " + err.Error())
		}
		f.Close()
	}
	_err := out.Close()
	if _err != nil {
		return "", "", 0, errors.New("close joined: " + err.Error())
	}

	// Transcode once: joined.webm -> final.wav (mono, 16kHz, 16-bit PCM)
	finalAudioLocalPath = filepath.Join(directory, "final.wav")
	cmd := exec.Command("ffmpeg", "-y", "-hide_banner", "-nostdin", // TODO change to exec.CommandContext(ctx, ...)
		"-fflags", "+genpts",
		"-i", joined,
		"-ac", "1", "-ar", "16000", "-c:a", "pcm_s16le",
		finalAudioLocalPath,
	)
	outb, err := cmd.CombinedOutput()
	if err != nil {
		return "", "", 0, errors.New("ffmpeg transcode: " + string(outb) + " err: " + err.Error())
	}

	// Measure audio duration using ffprobe
	audioDurationSeconds, err = getAudioDuration(finalAudioLocalPath)
	if err != nil {
		return "", "", 0, errors.New("get audio duration: " + err.Error())
	}

	// Uploads the complete audio in wav to storage
	finalAudioStoragePath = cloudStorage.FinalAudioUploadPath(manifest.UID, manifest.SessionID)
	err = cloudStorage.Upload(ctx, finalAudioStoragePath, finalAudioLocalPath, "audio/wav")
	if err != nil {
		return "", "", 0, errors.New("upload final: " + err.Error())
	}

	return finalAudioLocalPath, finalAudioStoragePath, audioDurationSeconds, nil
}

func getAudioDuration(audioPath string) (int, error) {
	cmd := exec.Command("ffprobe", "-v", "quiet", "-show_entries", "format=duration", "-of", "csv=p=0", audioPath)
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	durationStr := strings.TrimSpace(string(output))
	duration, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, err
	}

	return int(math.Ceil(duration)), nil
}

func chunkIndex(path string) int {
	base := filepath.Base(path)
	// chunk-000123.webm
	s := strings.TrimPrefix(base, "chunk-")
	s = strings.TrimSuffix(s, ".webm")
	var n int
	fmt.Sscanf(s, "%d", &n)
	return n
}

// // TODO MOVE TO STORAGE SERVICE
// func upload(ctx context.Context, bkt *storage.BucketHandle, gspath, local, contentType string) error {
// 	f, err := os.Open(local)
// 	if err != nil {
// 		return err
// 	}
// 	defer f.Close()

// 	w := bkt.Object(gspath).NewWriter(ctx)
// 	w.ContentType = contentType
// 	if _, err := io.Copy(w, f); err != nil {
// 		_ = w.Close()
// 		return err
// 	}
// 	return w.Close()
// }
// func download(ctx context.Context, bkt *storage.BucketHandle, gspath, local string) error {
// 	reader, err := bkt.Object(gspath).NewReader(ctx)
// 	if err != nil {
// 		return err
// 	}
// 	defer reader.Close()

// 	f, err := os.Create(local)
// 	if err != nil {
// 		return err
// 	}
// 	defer f.Close()

// 	_, err = io.Copy(f, reader)
// 	return err
// }
