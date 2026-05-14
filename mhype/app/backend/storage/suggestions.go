package storage

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type SuggestionResult struct {
	VideoID      string `json:"videoId"`
	Title        string `json:"title"`
	ChannelTitle string `json:"channelTitle"`
	ThumbnailURL string `json:"thumbnailUrl"`
}

type Suggestions struct {
	Artist    string             `json:"artist"`
	Title     string             `json:"title"`
	FetchedAt int64              `json:"fetchedAt"`
	Results   []SuggestionResult `json:"results"`
}

type SuggestionsStore struct {
	dir string
}

func NewSuggestionsStore(dataDir string) (*SuggestionsStore, error) {
	dir := filepath.Join(dataDir, "youtube")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("suggestions store: %w", err)
	}
	return &SuggestionsStore{dir: dir}, nil
}

func (s *SuggestionsStore) Has(artist, title string) bool {
	_, err := os.Stat(s.Path(artist, title))
	return err == nil
}

func (s *SuggestionsStore) Read(artist, title string) (Suggestions, bool, error) {
	data, err := os.ReadFile(s.Path(artist, title))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Suggestions{}, false, nil
		}
		return Suggestions{}, false, err
	}
	var sg Suggestions
	if err := json.Unmarshal(data, &sg); err != nil {
		return Suggestions{}, false, err
	}
	return sg, true, nil
}

func (s *SuggestionsStore) Write(sg Suggestions) error {
	final := s.Path(sg.Artist, sg.Title)
	tmp := final + ".tmp"
	data, err := json.MarshalIndent(sg, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, final)
}

func (s *SuggestionsStore) Path(artist, title string) string {
	return filepath.Join(s.dir, suggestionKey(artist, title)+".json")
}

func suggestionKey(artist, title string) string {
	raw := strings.ToLower(strings.TrimSpace(artist)) + "|" + strings.ToLower(strings.TrimSpace(title))
	sum := md5.Sum([]byte(raw))
	return hex.EncodeToString(sum[:])
}
