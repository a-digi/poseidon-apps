package storage

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Item mirrors crawler.Item field-for-field. The composer adapts between
// the two via a manual 1:1 copy (we can't share the type without an
// import cycle). Updating this struct without updating crawler.Item, or
// vice versa, is a bug.
type Item struct {
	ID        string   `json:"id"`
	Kind      ItemKind `json:"kind"`
	CrawlerID string   `json:"crawlerId"`
	SourceURL string   `json:"sourceUrl"`
	Title     string   `json:"title"`
	ScrapedAt int64    `json:"scrapedAt"`

	Artists     []string `json:"artists,omitempty"`
	Album       string   `json:"album,omitempty"`
	Label       string   `json:"label,omitempty"`
	Genres      []string `json:"genres,omitempty"`
	DurationSec int      `json:"durationSec,omitempty"`
	ReleasedAt  string   `json:"releasedAt,omitempty"`

	SourceID    string            `json:"sourceId,omitempty"`
	ExternalIDs map[string]string `json:"externalIds,omitempty"`

	ArtworkURL string `json:"artworkUrl,omitempty"`
	PreviewURL string `json:"previewUrl,omitempty"`

	Chart *ChartContext `json:"chart,omitempty"`
	News  *NewsContext  `json:"news,omitempty"`

	Extra map[string]string `json:"extra,omitempty"`
}

type ItemKind string

const (
	KindTrack      ItemKind = "track"
	KindChartEntry ItemKind = "chart-entry"
	KindRelease    ItemKind = "release"
	KindNews       ItemKind = "news"
)

type ChartContext struct {
	Name         string `json:"name"`
	Position     int    `json:"position"`
	PrevPosition int    `json:"prevPosition,omitempty"`
	PeakPosition int    `json:"peakPosition,omitempty"`
	WeeksOnChart int    `json:"weeksOnChart,omitempty"`
	PeriodStart  string `json:"periodStart,omitempty"`
	PeriodEnd    string `json:"periodEnd,omitempty"`
}

type NewsContext struct {
	Summary        string   `json:"summary,omitempty"`
	PublishedAt    string   `json:"publishedAt,omitempty"`
	RelatedArtists []string `json:"relatedArtists,omitempty"`
}

type Store struct {
	dataDir string
}

func NewStore(dataDir string) (*Store, error) {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, err
	}
	return &Store{dataDir: dataDir}, nil
}

func (s *Store) Write(crawlerID string, item Item) error {
	item.CrawlerID = crawlerID
	if item.ScrapedAt == 0 {
		item.ScrapedAt = time.Now().UnixMilli()
	}
	dir := s.crawlerDir(crawlerID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	name := fmt.Sprintf("%013d-%s.json", item.ScrapedAt, sha8(item.ID))
	final := filepath.Join(dir, name)
	tmp := final + ".tmp"
	data, err := json.MarshalIndent(item, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, final)
}

// WriteBundleIfMissing writes items as a single named JSON array file.
// Returns (true, nil) on a successful write, (false, nil) if <basename>.json
// already exists in this crawler's directory — the dedup contract.
// Returns (false, err) on any other failure.
//
// Unlike Write, this does NOT re-stamp CrawlerID or ScrapedAt on items —
// the caller (composer) is expected to have populated them. Bundle writes
// are an explicit "store this batch as-is" operation.
func (s *Store) WriteBundleIfMissing(crawlerID string, basename string, items []Item) (bool, error) {
	if basename == "" {
		return false, errors.New("storage: bundle basename is empty")
	}
	dir := filepath.Join(s.dataDir, "crawlers", crawlerID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return false, err
	}
	final := filepath.Join(dir, basename+".json")
	if _, err := os.Stat(final); err == nil {
		return false, nil
	} else if !errors.Is(err, fs.ErrNotExist) {
		return false, err
	}
	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return false, err
	}
	tmp := final + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return false, err
	}
	if err := os.Rename(tmp, final); err != nil {
		os.Remove(tmp)
		return false, err
	}
	return true, nil
}

func (s *Store) WriteBundle(crawlerID, basename string, items []Item) error {
	if basename == "" {
		return errors.New("storage: bundle basename is empty")
	}
	dir := filepath.Join(s.dataDir, "crawlers", crawlerID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	final := filepath.Join(dir, basename+".json")
	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}
	tmp := final + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, final); err != nil {
		os.Remove(tmp)
		return err
	}
	return nil
}

func (s *Store) List(crawlerID string, limit int) ([]Item, error) {
	dir := s.crawlerDir(crawlerID)
	names, err := sortedJSONNamesDesc(dir)
	if err != nil {
		return nil, err
	}
	out := make([]Item, 0, len(names))
	for _, name := range names {
		path := filepath.Join(dir, name)
		content, err := os.ReadFile(path)
		if err != nil {
			log.Printf("[storage] read %s: %v", path, err)
			continue
		}
		trimmed := bytes.TrimLeft(content, " \t\r\n")
		if len(trimmed) == 0 {
			continue
		}
		switch trimmed[0] {
		case '[':
			var bundle []Item
			if err := json.Unmarshal(content, &bundle); err != nil {
				log.Printf("[storage] decode bundle %s: %v", path, err)
				continue
			}
			out = append(out, bundle...)
		case '{':
			var single Item
			if err := json.Unmarshal(content, &single); err != nil {
				log.Printf("[storage] decode item %s: %v", path, err)
				continue
			}
			out = append(out, single)
		default:
			continue
		}
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (s *Store) FileCount(crawlerID string) (int, error) {
	dir := s.crawlerDir(crawlerID)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return 0, nil
		}
		return 0, err
	}
	count := 0
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
			count++
		}
	}
	return count, nil
}

func (s *Store) EvictOldest(crawlerID string, keep int) (int, error) {
	if keep <= 0 {
		return 0, nil
	}
	dir := s.crawlerDir(crawlerID)
	names, err := sortedJSONNamesDesc(dir)
	if err != nil {
		return 0, err
	}
	if len(names) <= keep {
		return 0, nil
	}
	removed := 0
	for _, name := range names[keep:] {
		if err := os.Remove(filepath.Join(dir, name)); err != nil {
			return removed, err
		}
		removed++
	}
	return removed, nil
}

func (s *Store) crawlerDir(crawlerID string) string {
	return filepath.Join(s.dataDir, "crawlers", crawlerID)
}

func sortedJSONNamesDesc(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []string{}, nil
		}
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		names = append(names, name)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(names)))
	return names, nil
}

func sha8(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])[:8]
}
