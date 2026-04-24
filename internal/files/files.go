package files

import (
	"GIN/internal/env"
	"fmt"
	"mime/multipart"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

func GenerateFilePath(file *multipart.FileHeader) string {
	defer func() {
		e := recover()
		if e != nil {
			return
		}
	}()
	dir := ""
	extension := filepath.Ext(file.Filename)
	switch strings.ToLower(extension) {
	case ".mp3", ".aac", ".wav", ".flac":
		dir = "/audios/"
	case ".jpeg", ".jpg", ".png", ".gif", ".tiff", ".raw", ".svg":
		dir = "/photos/"
	}

	uniqID := uuid.New().String()
	dst := "." + dir + uniqID + "-" + file.Filename
	return dst
}

func AvailablePhotoTypes() []string {
	return strings.Split(os.Getenv("AVAILABLE_PHOTO_TYPES"), ",")
}

func AvailableSongTypes() []string {
	return strings.Split(os.Getenv("AVAILABLE_SONG_TYPES"), ",")
}

func AvailableSizeCheck(file *multipart.FileHeader) bool {
	maxFileSIze := env.GetMaxFileSize()
	return file.Size <= int64(maxFileSIze)
}

func GetAudioDuration(filePath string) (float64, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		filePath,
	)

	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to run ffprobe: %w", err)
	}

	durationStr := strings.TrimSpace(string(output))
	duration, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration: %w", err)
	}

	return duration, nil
}
