package digitalitem

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/dhowden/tag"
)

type Metadata struct {
	Title    string `json:"title,omitempty"`
	Artist   string `json:"artist,omitempty"`
	Album    string `json:"album,omitempty"`
	Genre    string `json:"genre,omitempty"`
	Year     int    `json:"year,omitempty"`
	Track    int    `json:"track,omitempty"`
	Length   int    `json:"length,omitempty"`
	Picture  string `json:"picture,omitempty"`
	MimeType string `json:"mimeType,omitempty"`
}

func ExtractMetadata(filePath string) (*Metadata, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	m, err := tag.ReadFrom(f)
	if err != nil {
		return nil, err
	}

    track, _ := m.Track()

	meta := &Metadata{
		Title:  m.Title(),
		Artist: m.Artist(),
		Album:  m.Album(),
		Genre:  m.Genre(),
		Year:   m.Year(),
		Track:  track,
	}

	if pic := m.Picture(); pic != nil {
		meta.MimeType = pic.MIMEType
		b64 := base64.StdEncoding.EncodeToString(pic.Data)
		meta.Picture = fmt.Sprintf("data:%s;base64,%s", pic.MIMEType, b64)
	}

	return meta, nil
}

