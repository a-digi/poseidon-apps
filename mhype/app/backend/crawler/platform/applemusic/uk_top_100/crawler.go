package uk_top_100

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"mhype-plugin/crawler"
)

const (
	feedURL   = "https://rss.marketingtools.apple.com/api/v2/gb/music/most-played/100/songs.json"
	crawlerID = "applemusic-uk-top-100"
	chartName = "Apple Music UK Top 100"
	country   = "UK"
)

type impl struct{}

func New() crawler.Crawler { return impl{} }

func (impl) ID() string              { return crawlerID }
func (impl) DisplayName() string     { return chartName }
func (impl) Source() string          { return "music.apple.com" }
func (impl) Country() string         { return country }
func (impl) Interval() time.Duration { return 6 * time.Hour }

type appleFeed struct {
	Feed struct {
		Updated string        `json:"updated"`
		Results []appleResult `json:"results"`
	} `json:"feed"`
}

type appleResult struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	ArtistName    string `json:"artistName"`
	ArtworkURL100 string `json:"artworkUrl100"`
	URL           string `json:"url"`
	ReleaseDate   string `json:"releaseDate"`
	Genres        []struct {
		Name string `json:"name"`
	} `json:"genres"`
}

func (impl) Crawl(ctx context.Context, client *http.Client) (crawler.CrawlResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
	if err != nil {
		return crawler.CrawlResult{}, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "mhype-plugin/1.0")

	res, err := client.Do(req)
	if err != nil {
		return crawler.CrawlResult{}, err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return crawler.CrawlResult{}, fmt.Errorf("applemusic: %s", res.Status)
	}

	body, err := io.ReadAll(io.LimitReader(res.Body, 4*1024*1024))
	if err != nil {
		return crawler.CrawlResult{}, err
	}

	var feed appleFeed
	if err := json.Unmarshal(body, &feed); err != nil {
		return crawler.CrawlResult{}, fmt.Errorf("parse feed: %w", err)
	}

	updated := parseUpdated(feed.Feed.Updated)
	periodDate := updated.UTC().Format("2006-01-02")
	bundleKey := strconv.FormatInt(updated.Unix(), 10)
	scrapedAtMs := time.Now().UnixMilli()

	items := make([]crawler.Item, 0, len(feed.Feed.Results))
	for i, r := range feed.Feed.Results {
		position := i + 1
		artists := splitArtists(r.ArtistName)
		if len(artists) == 0 || r.Name == "" {
			continue
		}
		sourceURL := r.URL
		if sourceURL == "" {
			sourceURL = feedURL
		}
		item := crawler.Item{
			ID:         itemID(crawlerID, updated.Unix(), position, r.Name, artists[0]),
			Kind:       crawler.KindChartEntry,
			CrawlerID:  crawlerID,
			SourceURL:  sourceURL,
			Title:      r.Name,
			ScrapedAt:  scrapedAtMs,
			Artists:    artists,
			Genres:     filterGenres(r.Genres),
			ReleasedAt: r.ReleaseDate,
			ArtworkURL: r.ArtworkURL100,
			SourceID:   r.ID,
			Chart: &crawler.ChartContext{
				Name:        chartName,
				Position:    position,
				PeriodStart: periodDate,
				PeriodEnd:   periodDate,
			},
		}
		if err := item.Validate(); err != nil {
			continue
		}
		items = append(items, item)
	}
	return crawler.CrawlResult{Items: items, BundleKey: bundleKey}, nil
}

func parseUpdated(s string) time.Time {
	if s == "" {
		return time.Now().UTC()
	}
	if t, err := time.Parse(time.RFC1123Z, s); err == nil {
		return t.UTC()
	}
	log.Printf("[applemusic] could not parse updated %q, using now", s)
	return time.Now().UTC()
}

func filterGenres(in []struct {
	Name string `json:"name"`
}) []string {
	out := make([]string, 0, len(in))
	for _, g := range in {
		if g.Name == "" || g.Name == "Music" {
			continue
		}
		out = append(out, g.Name)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func splitArtists(s string) []string {
	parts := []string{s}
	parts = splitMany(parts, func(p string) []string {
		return splitOn(p, []string{" feat. ", " feat ", " ft. ", " ft "}, true)
	})
	parts = splitMany(parts, func(p string) []string { return strings.Split(p, " & ") })
	parts = splitMany(parts, func(p string) []string { return splitOn(p, []string{" x "}, true) })
	parts = splitMany(parts, func(p string) []string { return strings.Split(p, ", ") })
	parts = splitMany(parts, func(p string) []string { return strings.Split(p, " / ") })
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func splitMany(in []string, split func(string) []string) []string {
	out := make([]string, 0, len(in))
	for _, p := range in {
		out = append(out, split(p)...)
	}
	return out
}

func splitOn(s string, seps []string, caseInsensitive bool) []string {
	cmp := s
	if caseInsensitive {
		cmp = strings.ToLower(s)
	}
	for _, sep := range seps {
		idx := strings.Index(cmp, strings.ToLower(sep))
		if idx < 0 {
			continue
		}
		return []string{s[:idx], s[idx+len(sep):]}
	}
	return []string{s}
}

func itemID(cID string, updatedUnix int64, position int, title, artist string) string {
	key := strings.Join([]string{
		cID,
		strconv.FormatInt(updatedUnix, 10),
		strconv.Itoa(position),
		title,
		artist,
	}, "|")
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:])
}
