package storage

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type ArtistStat struct {
	Name          string `json:"name"`
	Count         int    `json:"count"`
	LatestArtwork string `json:"latestArtworkUrl,omitempty"`
	LatestPlayedAt int64 `json:"latestPlayedAt"`
}

type SongStat struct {
	Title          string   `json:"title"`
	Artists        []string `json:"artists"`
	ArtworkURL     string   `json:"artworkUrl,omitempty"`
	ChartName      string   `json:"chartName,omitempty"`
	Position       int      `json:"position,omitempty"`
	Country        string   `json:"country,omitempty"`
	Count          int      `json:"count"`
	LatestPlayedAt int64    `json:"latestPlayedAt"`
}

type ListStat struct {
	CrawlerID      string `json:"crawlerId"`
	DisplayName    string `json:"displayName"`
	Country        string `json:"country,omitempty"`
	Count          int    `json:"count"`
	LatestPlayedAt int64  `json:"latestPlayedAt"`
}

type PlayEvent struct {
	Title       string
	Artists     []string
	ArtworkURL  string
	ChartName   string
	CrawlerID   string
	DisplayName string
	Position    int
	Country     string
}

type AnalyticsStore struct {
	path    string
	mu      sync.Mutex
	artists map[string]ArtistStat
	songs   map[string]SongStat
	lists   map[string]ListStat
}

type analyticsPayload struct {
	Artists map[string]ArtistStat `json:"artists"`
	Songs   map[string]SongStat   `json:"songs"`
	Lists   map[string]ListStat   `json:"lists"`
}

func NewAnalyticsStore(dataDir string) (*AnalyticsStore, error) {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, fmt.Errorf("analytics store: %w", err)
	}
	path := filepath.Join(dataDir, "analytics.json")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &AnalyticsStore{
			path:    path,
			artists: map[string]ArtistStat{},
			songs:   map[string]SongStat{},
			lists:   map[string]ListStat{},
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("analytics store: %w", err)
	}
	var p analyticsPayload
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("analytics store: %w", err)
	}
	if p.Artists == nil {
		p.Artists = map[string]ArtistStat{}
	}
	if p.Songs == nil {
		p.Songs = map[string]SongStat{}
	}
	if p.Lists == nil {
		p.Lists = map[string]ListStat{}
	}
	return &AnalyticsStore{path: path, artists: p.Artists, songs: p.Songs, lists: p.Lists}, nil
}

func (s *AnalyticsStore) TrackPlay(ev PlayEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UnixMilli()

	firstArtist := ""
	if len(ev.Artists) > 0 {
		firstArtist = ev.Artists[0]
	}

	if ak := artistKey(firstArtist); ak != "" {
		stat := s.artists[ak]
		stat.Name = firstArtist
		stat.Count++
		stat.LatestPlayedAt = now
		if ev.ArtworkURL != "" {
			stat.LatestArtwork = ev.ArtworkURL
		}
		s.artists[ak] = stat
	}

	if sk := songKey(ev.Title, firstArtist); sk != "" {
		stat := s.songs[sk]
		stat.Title = ev.Title
		stat.Artists = ev.Artists
		stat.Count++
		stat.LatestPlayedAt = now
		if ev.ArtworkURL != "" {
			stat.ArtworkURL = ev.ArtworkURL
		}
		if ev.ChartName != "" {
			stat.ChartName = ev.ChartName
		}
		if ev.Position != 0 {
			stat.Position = ev.Position
		}
		if ev.Country != "" {
			stat.Country = ev.Country
		}
		s.songs[sk] = stat
	}

	if lk := listKey(ev.CrawlerID); lk != "" {
		stat := s.lists[lk]
		stat.CrawlerID = ev.CrawlerID
		stat.Count++
		stat.LatestPlayedAt = now
		if ev.DisplayName != "" {
			stat.DisplayName = ev.DisplayName
		}
		if ev.Country != "" {
			stat.Country = ev.Country
		}
		s.lists[lk] = stat
	}

	return s.persist()
}

func (s *AnalyticsStore) TopArtists(limit int) []ArtistStat {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]ArtistStat, 0, len(s.artists))
	for _, v := range s.artists {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].LatestPlayedAt > out[j].LatestPlayedAt
	})
	if limit < len(out) {
		out = out[:limit]
	}
	return out
}

func (s *AnalyticsStore) TopSongs(limit int) []SongStat {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]SongStat, 0, len(s.songs))
	for _, v := range s.songs {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].LatestPlayedAt > out[j].LatestPlayedAt
	})
	if limit < len(out) {
		out = out[:limit]
	}
	return out
}

func (s *AnalyticsStore) TopLists(limit int) []ListStat {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]ListStat, 0, len(s.lists))
	for _, v := range s.lists {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].LatestPlayedAt > out[j].LatestPlayedAt
	})
	if limit < len(out) {
		out = out[:limit]
	}
	return out
}

func (s *AnalyticsStore) persist() error {
	data, err := json.MarshalIndent(analyticsPayload{
		Artists: s.artists,
		Songs:   s.songs,
		Lists:   s.lists,
	}, "", "  ")
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

func artistKey(name string) string {
	s := strings.TrimSpace(name)
	if s == "" {
		return ""
	}
	sum := md5.Sum([]byte(strings.ToLower(s)))
	return hex.EncodeToString(sum[:])
}

func songKey(title, firstArtist string) string {
	t := strings.TrimSpace(title)
	a := strings.TrimSpace(firstArtist)
	if t == "" && a == "" {
		return ""
	}
	sum := md5.Sum([]byte(strings.ToLower(t) + "|" + strings.ToLower(a)))
	return hex.EncodeToString(sum[:])
}

func listKey(crawlerID string) string {
	return crawlerID
}
