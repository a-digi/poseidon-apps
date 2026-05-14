package storage

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type YouTubeCacheEntry struct {
	VideoID      string `json:"videoId"`
	Title        string `json:"title"`
	ChannelTitle string `json:"channelTitle"`
	ThumbnailURL string `json:"thumbnailUrl"`
	CachedAt     int64  `json:"cachedAt"`
}

type YouTubeCacheStore struct {
	path string
	mem  map[string]YouTubeCacheEntry
}

func NewYouTubeCacheStore(dataDir string) (*YouTubeCacheStore, error) {
	dir := filepath.Join(dataDir, "youtube")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("youtube cache store: %w", err)
	}
	path := filepath.Join(dir, "cache.json")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) || (err == nil && len(data) == 0) {
		return &YouTubeCacheStore{path: path, mem: map[string]YouTubeCacheEntry{}}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("youtube cache store: %w", err)
	}
	var mem map[string]YouTubeCacheEntry
	if err := json.Unmarshal(data, &mem); err != nil {
		return nil, fmt.Errorf("youtube cache store: %w", err)
	}
	return &YouTubeCacheStore{path: path, mem: mem}, nil
}

func (s *YouTubeCacheStore) Get(artist, title string) (YouTubeCacheEntry, bool) {
	e, ok := s.mem[youTubeCacheKey(artist, title)]
	return e, ok
}

func (s *YouTubeCacheStore) Put(artist, title string, entry YouTubeCacheEntry) error {
	s.mem[youTubeCacheKey(artist, title)] = entry
	data, err := json.MarshalIndent(s.mem, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, s.path); err != nil {
		os.Remove(tmp)
		return err
	}
	return nil
}

func youTubeCacheKey(artist, title string) string {
	raw := strings.ToLower(strings.TrimSpace(title)) + "-" + strings.ToLower(strings.TrimSpace(artist))
	sum := md5.Sum([]byte(raw))
	return hex.EncodeToString(sum[:])
}
