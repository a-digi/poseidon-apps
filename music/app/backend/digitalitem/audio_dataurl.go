package digitalitem

import (
	"encoding/base64"
	"fmt"
	"os"
)

// GetAudioDataURL liest eine Audiodatei ein und gibt eine base64-Data-URL zurück
func GetAudioDataURL(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	ext := ""
	if len(path) > 0 {
		for i := len(path) - 1; i >= 0; i-- {
			if path[i] == '.' {
				ext = path[i:]
				break
			}
		}
	}
	mime := "audio/mpeg"
	switch ext {
	case ".wav":
		mime = "audio/wav"
	case ".ogg":
		mime = "audio/ogg"
	case ".flac":
		mime = "audio/flac"
	case ".aac":
		mime = "audio/aac"
	case ".m4a":
		mime = "audio/mp4"
	}
	base64Data := base64.StdEncoding.EncodeToString(data)
	return fmt.Sprintf("data:%s;base64,%s", mime, base64Data), nil
}

