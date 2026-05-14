package crawler

import (
	"errors"
	"fmt"
)

// ItemKind discriminates between music-domain record types so consumers
// can dispatch on the type without inspecting field presence.
type ItemKind string

const (
	KindTrack      ItemKind = "track"
	KindChartEntry ItemKind = "chart-entry"
	KindRelease    ItemKind = "release"
	KindNews       ItemKind = "news"
)

var validKinds = map[ItemKind]bool{
	KindTrack: true, KindChartEntry: true, KindRelease: true, KindNews: true,
}

// Item is the canonical music-domain record every crawler emits.
// Field-level rules:
//   - Required fields (id, kind, crawlerId, sourceUrl, title, scrapedAt)
//     must be non-zero on every item.
//   - Kind-specific contexts (Chart, News) are non-nil only for the
//     matching Kind. Validate() enforces this.
//   - All other fields are optional; absence means "unknown", not "zero".
//
// Item must stay field-for-field identical to storage.Item — the composer
// uses a manual 1:1 adapter to bridge the two packages (we can't share
// the type without an import cycle).
type Item struct {
	// Obligatory.
	ID        string   `json:"id"`
	Kind      ItemKind `json:"kind"`
	CrawlerID string   `json:"crawlerId"`
	SourceURL string   `json:"sourceUrl"`
	Title     string   `json:"title"`
	ScrapedAt int64    `json:"scrapedAt"`

	// Musical content metadata (optional).
	Artists     []string `json:"artists,omitempty"`
	Album       string   `json:"album,omitempty"`
	Label       string   `json:"label,omitempty"`
	Genres      []string `json:"genres,omitempty"`
	DurationSec int      `json:"durationSec,omitempty"`
	ReleasedAt  string   `json:"releasedAt,omitempty"`

	// Cross-source identifiers (optional).
	SourceID    string            `json:"sourceId,omitempty"`
	ExternalIDs map[string]string `json:"externalIds,omitempty"`

	// Media (optional).
	ArtworkURL string `json:"artworkUrl,omitempty"`
	PreviewURL string `json:"previewUrl,omitempty"`

	// Kind-specific contexts (only the matching kind populates these).
	Chart *ChartContext `json:"chart,omitempty"`
	News  *NewsContext  `json:"news,omitempty"`

	// Last-resort source-specific extras.
	Extra map[string]string `json:"extra,omitempty"`
}

// ChartContext is populated when Kind == KindChartEntry.
// Name and Position are required when Chart is non-nil.
type ChartContext struct {
	Name         string `json:"name"`
	Position     int    `json:"position"`
	PrevPosition int    `json:"prevPosition,omitempty"`
	PeakPosition int    `json:"peakPosition,omitempty"`
	WeeksOnChart int    `json:"weeksOnChart,omitempty"`
	PeriodStart  string `json:"periodStart,omitempty"`
	PeriodEnd    string `json:"periodEnd,omitempty"`
}

// NewsContext is populated when Kind == KindNews. All fields are optional
// at the kind level — the wrapper itself just must be non-nil for KindNews.
type NewsContext struct {
	Summary        string   `json:"summary,omitempty"`
	PublishedAt    string   `json:"publishedAt,omitempty"`
	RelatedArtists []string `json:"relatedArtists,omitempty"`
}

// Validate returns the FIRST validation error it finds, or nil if the
// item is well-formed for its declared Kind. Crawlers and the composer's
// bridge call this before persisting.
func (it Item) Validate() error {
	if it.ID == "" {
		return errors.New("item: id is required")
	}
	if !validKinds[it.Kind] {
		return fmt.Errorf("item: invalid kind %q", it.Kind)
	}
	if it.CrawlerID == "" {
		return errors.New("item: crawlerId is required")
	}
	if it.SourceURL == "" {
		return errors.New("item: sourceUrl is required")
	}
	if it.Title == "" {
		return errors.New("item: title is required")
	}
	if it.ScrapedAt <= 0 {
		return errors.New("item: scrapedAt must be > 0")
	}
	switch it.Kind {
	case KindChartEntry:
		if it.Chart == nil {
			return errors.New("item: chart context is required for chart-entry")
		}
		if it.Chart.Name == "" {
			return errors.New("item: chart.name is required")
		}
		if it.Chart.Position <= 0 {
			return errors.New("item: chart.position must be > 0")
		}
	case KindNews:
		if it.News == nil {
			return errors.New("item: news context is required for news")
		}
	}
	return nil
}

