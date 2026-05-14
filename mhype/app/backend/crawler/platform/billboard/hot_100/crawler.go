package hot_100

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"mhype-plugin/crawler"
)

const (
	chartURL  = "https://www.billboard.com/charts/hot-100/"
	crawlerID = "billboard-hot-100"
	chartName = "Billboard Hot 100"
	country   = "US"
)

type impl struct{}

func New() crawler.Crawler { return impl{} }

func (impl) ID() string              { return crawlerID }
func (impl) DisplayName() string     { return chartName }
func (impl) Source() string          { return "billboard.com" }
func (impl) Country() string         { return country }
func (impl) Interval() time.Duration { return 12 * time.Hour }

func (impl) Crawl(ctx context.Context, client *http.Client) (crawler.CrawlResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, chartURL, nil)
	if err != nil {
		return crawler.CrawlResult{}, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	res, err := client.Do(req)
	if err != nil {
		return crawler.CrawlResult{}, err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return crawler.CrawlResult{}, fmt.Errorf("billboard: %s", res.Status)
	}

	body, err := io.ReadAll(io.LimitReader(res.Body, 4*1024*1024))
	if err != nil {
		return crawler.CrawlResult{}, err
	}

	return extract(body, time.Now().UnixMilli())
}
