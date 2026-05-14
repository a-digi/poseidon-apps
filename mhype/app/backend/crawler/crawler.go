package crawler

import (
	"context"
	"net/http"
	"time"
)

// CrawlResult is what a Crawler returns from one Crawl run.
type CrawlResult struct {
	// Items is the canonical records the crawler produced. An empty slice
	// is a valid "no data this run" signal — composer logs and continues.
	Items []Item

	// BundleKey, when non-empty, asks the composer to write all Items as a single named JSON array via Store.WriteBundleIfMissing — the basename is BundleKey verbatim (storage appends ".json"). Empty BundleKey selects the per-item write path (one file per Item via Store.Write).
	BundleKey string

	// Snapshot, when true with a non-empty BundleKey, asks the composer to overwrite <crawlerDir>/<BundleKey>.json on every run.
	Snapshot bool
}

type Crawler interface {
	ID() string
	DisplayName() string
	Source() string
	Country() string
	Interval() time.Duration
	Crawl(ctx context.Context, client *http.Client) (CrawlResult, error)
}
