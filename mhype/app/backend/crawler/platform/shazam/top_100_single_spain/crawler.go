package top_100_single_spain

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"mhype-plugin/crawler"
)

const (
	csvURL    = "https://www.shazam.com/services/charts/csv/top-200/spain/"
	htmlURL   = "https://www.shazam.com/de-de/charts/top-200/spain"
	crawlerID = "shazam-top-100-single-spain"
	chartName = "Shazam Top 100 Spain"
	country   = "ES"
	limit     = 100
	userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Safari/605.1.15"
)

type impl struct{}

func New() crawler.Crawler { return impl{} }

func (impl) ID() string              { return crawlerID }
func (impl) DisplayName() string     { return chartName }
func (impl) Source() string          { return "shazam.com" }
func (impl) Country() string         { return country }
func (impl) Interval() time.Duration { return 12 * time.Hour }

func (impl) Crawl(ctx context.Context, client *http.Client) (crawler.CrawlResult, error) {
	csvBody, err := fetch(ctx, client, csvURL, 256*1024)
	if err != nil {
		return crawler.CrawlResult{}, fmt.Errorf("shazam csv: %w", err)
	}
	htmlBody, _ := fetch(ctx, client, htmlURL, 4*1024*1024)
	return extract(csvBody, htmlBody, time.Now().UnixMilli())
}

func fetch(ctx context.Context, client *http.Client, url string, maxBytes int64) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept-Language", "de-DE,de;q=0.9,en;q=0.8")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,text/csv,*/*;q=0.8")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("status %s", res.Status)
	}
	return io.ReadAll(io.LimitReader(res.Body, maxBytes))
}
