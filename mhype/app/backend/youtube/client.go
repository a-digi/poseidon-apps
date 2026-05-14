package youtube

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const defaultInvidiousURL = "https://invidious.io"

type Result struct {
	VideoID      string `json:"videoId"`
	Title        string `json:"title"`
	ChannelTitle string `json:"channelTitle"`
	ThumbnailURL string `json:"thumbnailUrl"`
}

var ErrNoAPIKey = errors.New("YOUTUBE_API_KEY is not set")

type Client struct {
	httpClient   *http.Client
	apiKey       string
	invidiousURL string
}

func NewClient(httpClient *http.Client) *Client {
	invidiousURL := os.Getenv("INVIDIOUS_URL")
	if invidiousURL == "" {
		invidiousURL = defaultInvidiousURL
	}
	return &Client{
		httpClient:   httpClient,
		apiKey:       os.Getenv("YOUTUBE_API_KEY"),
		invidiousURL: invidiousURL,
	}
}

func (c *Client) Search(ctx context.Context, artist, title string) ([]Result, error) {
	if c.apiKey != "" {
		return c.searchOfficial(ctx, artist, title)
	}
	if results, err := c.searchInvidious(ctx, artist, title); err == nil && len(results) > 0 {
		return results, nil
	}
	return c.searchScrape(ctx, artist, title)
}

func (c *Client) searchOfficial(ctx context.Context, artist, title string) ([]Result, error) {
	q := url.Values{}
	q.Set("part", "snippet")
	q.Set("q", artist+" "+title)
	q.Set("type", "video")
	q.Set("maxResults", "3")
	q.Set("key", c.apiKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://www.googleapis.com/youtube/v3/search?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		buf := make([]byte, 4096)
		n, _ := res.Body.Read(buf)
		return nil, fmt.Errorf("youtube search %d: %s", res.StatusCode, buf[:n])
	}

	var apiRes struct {
		Items []struct {
			ID struct {
				VideoID string `json:"videoId"`
			} `json:"id"`
			Snippet struct {
				Title        string `json:"title"`
				ChannelTitle string `json:"channelTitle"`
				Thumbnails   struct {
					High struct {
						URL string `json:"url"`
					} `json:"high"`
				} `json:"thumbnails"`
			} `json:"snippet"`
		} `json:"items"`
	}
	if err := json.NewDecoder(res.Body).Decode(&apiRes); err != nil {
		return nil, fmt.Errorf("youtube decode: %w", err)
	}

	out := make([]Result, 0, len(apiRes.Items))
	for _, it := range apiRes.Items {
		if it.ID.VideoID == "" {
			continue
		}
		out = append(out, Result{
			VideoID:      it.ID.VideoID,
			Title:        it.Snippet.Title,
			ChannelTitle: it.Snippet.ChannelTitle,
			ThumbnailURL: it.Snippet.Thumbnails.High.URL,
		})
	}
	return out, nil
}

func (c *Client) searchInvidious(ctx context.Context, artist, title string) ([]Result, error) {
	q := url.Values{}
	q.Set("q", artist+" "+title)
	q.Set("type", "video")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.invidiousURL+"/api/v1/search?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("invidious: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		buf := make([]byte, 4096)
		n, _ := res.Body.Read(buf)
		return nil, fmt.Errorf("invidious search %d: %s", res.StatusCode, buf[:n])
	}

	var items []struct {
		VideoID    string `json:"videoId"`
		Title      string `json:"title"`
		Author     string `json:"author"`
		Thumbnails []struct {
			Quality string `json:"quality"`
			URL     string `json:"url"`
		} `json:"videoThumbnails"`
	}
	if err := json.NewDecoder(res.Body).Decode(&items); err != nil {
		return nil, fmt.Errorf("invidious decode: %w", err)
	}

	out := make([]Result, 0, 3)
	for _, it := range items {
		if it.VideoID == "" || len(out) == 3 {
			break
		}
		out = append(out, Result{
			VideoID:      it.VideoID,
			Title:        it.Title,
			ChannelTitle: it.Author,
			ThumbnailURL: bestThumbnail(it.Thumbnails),
		})
	}
	return out, nil
}

func (c *Client) searchScrape(ctx context.Context, artist, title string) ([]Result, error) {
	q := url.QueryEscape(artist + " " + title)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://www.youtube.com/results?q="+q+"&hl=en", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("youtube scrape: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("youtube scrape read: %w", err)
	}

	html := string(body)
	const marker = "var ytInitialData = "
	idx := strings.Index(html, marker)
	if idx < 0 {
		return nil, errors.New("youtube scrape: ytInitialData not found")
	}
	rest := html[idx+len(marker):]
	end := strings.Index(rest, ";</script>")
	if end < 0 {
		return nil, errors.New("youtube scrape: ytInitialData end marker not found")
	}

	var data any
	if err := json.Unmarshal([]byte(rest[:end]), &data); err != nil {
		return nil, fmt.Errorf("youtube scrape parse: %w", err)
	}

	renderers := collectVideoRenderers(data)
	out := make([]Result, 0, 5)
	for _, vr := range renderers {
		if len(out) >= 5 {
			break
		}
		r := parseVideoRenderer(vr)
		if r == nil {
			continue
		}
		out = append(out, *r)
	}
	return out, nil
}

// collectVideoRenderers walks the ytInitialData tree and returns every videoRenderer object found.
func collectVideoRenderers(node any) []map[string]any {
	var results []map[string]any
	switch v := node.(type) {
	case map[string]any:
		if vr, ok := v["videoRenderer"].(map[string]any); ok {
			results = append(results, vr)
		}
		for _, child := range v {
			results = append(results, collectVideoRenderers(child)...)
		}
	case []any:
		for _, item := range v {
			results = append(results, collectVideoRenderers(item)...)
		}
	}
	return results
}

func parseVideoRenderer(vr map[string]any) *Result {
	videoID, _ := vr["videoId"].(string)
	if videoID == "" {
		return nil
	}

	title := ""
	if t, ok := vr["title"].(map[string]any); ok {
		if runs, ok := t["runs"].([]any); ok && len(runs) > 0 {
			if run, ok := runs[0].(map[string]any); ok {
				title, _ = run["text"].(string)
			}
		}
	}

	channel := ""
	if oc, ok := vr["ownerText"].(map[string]any); ok {
		if runs, ok := oc["runs"].([]any); ok && len(runs) > 0 {
			if run, ok := runs[0].(map[string]any); ok {
				channel, _ = run["text"].(string)
			}
		}
	}

	thumbnailURL := ""
	if thumb, ok := vr["thumbnail"].(map[string]any); ok {
		if thumbs, ok := thumb["thumbnails"].([]any); ok && len(thumbs) > 0 {
			last, _ := thumbs[len(thumbs)-1].(map[string]any)
			thumbnailURL, _ = last["url"].(string)
		}
	}

	return &Result{
		VideoID:      videoID,
		Title:        title,
		ChannelTitle: channel,
		ThumbnailURL: thumbnailURL,
	}
}

func bestThumbnail(thumbs []struct {
	Quality string `json:"quality"`
	URL     string `json:"url"`
}) string {
	priority := []string{"high", "medium", "default"}
	for _, want := range priority {
		for _, t := range thumbs {
			if t.Quality == want && t.URL != "" {
				return t.URL
			}
		}
	}
	if len(thumbs) > 0 {
		return thumbs[0].URL
	}
	return ""
}
