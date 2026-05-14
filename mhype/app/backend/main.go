package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"mhype-plugin/orchestrator"
	"mhype-plugin/registry"
	"mhype-plugin/storage"
	"mhype-plugin/youtube"
)

type envelope struct {
	Result any    `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
}

const (
	triggersDirName  = "triggers"
	defaultListLimit = 50
	maxListLimit     = 100
)

func main() {
	log.SetPrefix("[mhype] ")
	log.SetFlags(log.LstdFlags)
	for _, a := range os.Args[1:] {
		if a == "--orchestrator" {
			runOrchestratorMode()
			return
		}
	}
	runCLIMode()
}

func runOrchestratorMode() {
	dataDir := resolveDataDir()
	log.SetPrefix("")
	if err := orchestrator.Run(dataDir, registry.Registered()); err != nil {
		log.Fatalf("orchestrator: %v", err)
	}
}

func runCLIMode() {
	enc := json.NewEncoder(os.Stdout)
	var params map[string]any
	if err := json.NewDecoder(os.Stdin).Decode(&params); err != nil {
		enc.Encode(envelope{Error: "bad request: " + err.Error()})
		return
	}
	action, _ := params["action"].(string)
	dataDir := resolveDataDir()
	result, err := dispatch(action, params, dataDir)
	if err != nil {
		enc.Encode(envelope{Error: err.Error()})
		return
	}
	enc.Encode(envelope{Result: result})
}

func dispatch(action string, params map[string]any, dataDir string) (any, error) {
	switch action {
	case "list_crawlers":
		return listCrawlers(dataDir)
	case "list_items":
		return listItems(params, dataDir)
	case "trigger_crawl":
		return triggerCrawl(params, dataDir)
	case "queue_chart_youtube":
		return queueChartYouTube(params, dataDir)
	case "get_youtube_suggestions":
		return getYouTubeSuggestions(params, dataDir)
	case "get_settings":
		return getSettings(dataDir)
	case "set_crawler_active":
		return setCrawlerActive(params, dataDir)
	case "open_browser":
		return openBrowser(params)
	case "find_youtube_video":
		return findYouTubeVideo(params, dataDir)
	case "track_play":
		return trackPlay(params, dataDir)
	case "get_highlights":
		return getHighlights(params, dataDir)
	default:
		return nil, errors.New("unknown action: " + action)
	}
}

func resolveDataDir() string {
	if dir := os.Getenv("PLUGIN_DATA_DIR"); dir != "" {
		return dir
	}
	exe, err := os.Executable()
	if err != nil {
		return "./data"
	}
	pluginRoot := filepath.Dir(filepath.Dir(exe))
	return filepath.Join(pluginRoot, "data")
}

func listCrawlers(dataDir string) (any, error) {
	store, err := storage.NewStore(dataDir)
	if err != nil {
		return nil, err
	}
	state, err := orchestrator.NewStateStore(dataDir)
	if err != nil {
		return nil, fmt.Errorf("init orchestrator state: %w", err)
	}
	ss, err := storage.NewSettingsStore(dataDir)
	if err != nil {
		return nil, err
	}
	registered := registry.Registered()
	out := make([]map[string]any, 0, len(registered))
	for _, c := range registered {
		fileCount, _ := store.FileCount(c.ID())
		lastSuccessAt := state.Get(c.ID()).LastSuccessAt
		out = append(out, map[string]any{
			"id":            c.ID(),
			"displayName":   c.DisplayName(),
			"source":        c.Source(),
			"country":       c.Country(),
			"intervalSec":   int(c.Interval() / time.Second),
			"fileCount":     fileCount,
			"lastSuccessAt": lastSuccessAt,
			"active":        ss.IsActive(c.ID()),
		})
	}
	return out, nil
}

func listItems(params map[string]any, dataDir string) (any, error) {
	crawlerID, _ := params["crawlerId"].(string)
	if crawlerID == "" {
		return nil, errors.New("crawlerId required")
	}
	limit := defaultListLimit
	if raw, ok := params["limit"]; ok {
		if n, ok := toPositiveInt(raw); ok {
			limit = n
		}
	}
	if limit > maxListLimit {
		limit = maxListLimit
	}
	store, err := storage.NewStore(dataDir)
	if err != nil {
		return nil, err
	}
	items, err := store.List(crawlerID, limit)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = []storage.Item{}
	}
	return items, nil
}

func toPositiveInt(raw any) (int, bool) {
	switch v := raw.(type) {
	case float64:
		if v > 0 {
			return int(v), true
		}
	case string:
		n, err := strconv.Atoi(v)
		if err == nil && n > 0 {
			return n, true
		}
	}
	return 0, false
}

func triggerCrawl(params map[string]any, dataDir string) (any, error) {
	crawlerID, _ := params["crawlerId"].(string)
	if crawlerID == "" {
		return nil, errors.New("crawlerId required")
	}
	known := false
	for _, c := range registry.Registered() {
		if c.ID() == crawlerID {
			known = true
			break
		}
	}
	if !known {
		return nil, errors.New("unknown crawler")
	}

	triggerDir := filepath.Join(dataDir, triggersDirName)
	if err := os.MkdirAll(triggerDir, 0o755); err != nil {
		return nil, err
	}
	clickTs := time.Now().UnixMilli() - 1
	if err := os.WriteFile(filepath.Join(triggerDir, crawlerID), []byte{}, 0o644); err != nil {
		return nil, err
	}

	const (
		pollEvery = 250 * time.Millisecond
		deadline  = 45 * time.Second
	)
	start := time.Now()
	for {
		time.Sleep(pollEvery)
		states, err := orchestrator.LoadStateFromDisk(dataDir)
		if err != nil {
			return nil, err
		}
		rs := states[crawlerID]
		if rs.LastSuccessAt > clickTs {
			return map[string]any{"ok": true}, nil
		}
		if rs.LastFailureAt > clickTs {
			msg := rs.LastError
			if msg == "" {
				msg = "crawl failed"
			}
			return nil, errors.New(msg)
		}
		if time.Since(start) >= deadline {
			return nil, errors.New("timed out waiting for crawler to complete")
		}
	}
}

func queueChartYouTube(params map[string]any, dataDir string) (any, error) {
	payload, _ := params["payload"].(map[string]any)
	artist, _ := payload["artist"].(string)
	title, _ := payload["title"].(string)
	if artist == "" || title == "" {
		return nil, errors.New("artist and title are required")
	}

	ss, err := storage.NewSuggestionsStore(dataDir)
	if err != nil {
		return nil, err
	}
	if ss.Has(artist, title) {
		return map[string]any{"ok": true, "cached": true}, nil
	}

	httpClient := &http.Client{Timeout: 25 * time.Second}
	yt := youtube.NewClient(httpClient)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	results, err := yt.Search(ctx, artist, title)
	if err != nil {
		return nil, err
	}
	sg := storage.Suggestions{
		Artist:    artist,
		Title:     title,
		FetchedAt: time.Now().UnixMilli(),
		Results:   make([]storage.SuggestionResult, 0, len(results)),
	}
	for _, r := range results {
		sg.Results = append(sg.Results, storage.SuggestionResult{
			VideoID: r.VideoID, Title: r.Title, ChannelTitle: r.ChannelTitle, ThumbnailURL: r.ThumbnailURL,
		})
	}
	if err := ss.Write(sg); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "fetched": len(sg.Results)}, nil
}

func getYouTubeSuggestions(params map[string]any, dataDir string) (any, error) {
	artist, _ := params["artist"].(string)
	title, _ := params["title"].(string)
	if artist == "" || title == "" {
		return nil, errors.New("artist and title are required")
	}
	ss, err := storage.NewSuggestionsStore(dataDir)
	if err != nil {
		return nil, err
	}
	sg, found, err := ss.Read(artist, title)
	if err != nil {
		return nil, err
	}
	if !found {
		return map[string]any{"found": false, "results": []storage.SuggestionResult{}}, nil
	}
	return map[string]any{"found": true, "results": sg.Results}, nil
}

func getSettings(dataDir string) (any, error) {
	ss, err := storage.NewSettingsStore(dataDir)
	if err != nil {
		return nil, err
	}
	return ss.Load()
}

func setCrawlerActive(params map[string]any, dataDir string) (any, error) {
	crawlerID, _ := params["crawlerId"].(string)
	if crawlerID == "" {
		return nil, errors.New("crawlerId required")
	}
	active, ok := params["active"].(bool)
	if !ok {
		return nil, errors.New("active required")
	}
	known := false
	for _, c := range registry.Registered() {
		if c.ID() == crawlerID {
			known = true
			break
		}
	}
	if !known {
		return nil, errors.New("unknown crawler")
	}
	ss, err := storage.NewSettingsStore(dataDir)
	if err != nil {
		return nil, err
	}
	if err := ss.SetActive(crawlerID, active); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true}, nil
}

func findYouTubeVideo(params map[string]any, dataDir string) (any, error) {
	artist, _ := params["artist"].(string)
	title, _ := params["title"].(string)
	if artist == "" && title == "" {
		return nil, errors.New("artist or title required")
	}

	cache, err := storage.NewYouTubeCacheStore(dataDir)
	if err != nil {
		return nil, err
	}
	if e, ok := cache.Get(artist, title); ok {
		return map[string]any{
			"videoId":      e.VideoID,
			"title":        e.Title,
			"channelTitle": e.ChannelTitle,
			"thumbnailUrl": e.ThumbnailURL,
		}, nil
	}

	httpClient := &http.Client{Timeout: 25 * time.Second}
	yt := youtube.NewClient(httpClient)

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	results, err := yt.Search(ctx, artist, title)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, errors.New("no YouTube video found")
	}
	r := results[0]
	_ = cache.Put(artist, title, storage.YouTubeCacheEntry{
		VideoID:      r.VideoID,
		Title:        r.Title,
		ChannelTitle: r.ChannelTitle,
		ThumbnailURL: r.ThumbnailURL,
		CachedAt:     time.Now().UnixMilli(),
	})
	return map[string]any{
		"videoId":      r.VideoID,
		"title":        r.Title,
		"channelTitle": r.ChannelTitle,
		"thumbnailUrl": r.ThumbnailURL,
	}, nil
}

func asStringSlice(raw any) []string {
	arr, _ := raw.([]any)
	out := make([]string, 0, len(arr))
	for _, v := range arr {
		if s, ok := v.(string); ok && s != "" {
			out = append(out, s)
		}
	}
	return out
}

func trackPlay(params map[string]any, dataDir string) (any, error) {
	title, _ := params["title"].(string)
	artists := asStringSlice(params["artists"])
	if title == "" && len(artists) == 0 {
		return nil, errors.New("title or artists required")
	}
	artworkURL, _ := params["artworkUrl"].(string)
	chartName, _ := params["chartName"].(string)
	crawlerID, _ := params["crawlerId"].(string)
	country, _ := params["country"].(string)
	position, _ := toPositiveInt(params["position"])

	displayName := chartName
	for _, c := range registry.Registered() {
		if c.ID() == crawlerID {
			displayName = c.DisplayName()
			if country == "" {
				country = c.Country()
			}
			break
		}
	}

	store, err := storage.NewAnalyticsStore(dataDir)
	if err != nil {
		return nil, err
	}
	if err := store.TrackPlay(storage.PlayEvent{
		Title: title, Artists: artists, ArtworkURL: artworkURL,
		ChartName: chartName, CrawlerID: crawlerID, DisplayName: displayName,
		Position: position, Country: country,
	}); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true}, nil
}

func getHighlights(params map[string]any, dataDir string) (any, error) {
	limit := 10
	if n, ok := toPositiveInt(params["limit"]); ok && n <= 100 {
		limit = n
	}
	store, err := storage.NewAnalyticsStore(dataDir)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"artists":   store.TopArtists(limit),
		"songs":     store.TopSongs(limit),
		"playlists": store.TopLists(limit),
	}, nil
}

func openBrowser(params map[string]any) (any, error) {
	rawURL, _ := params["url"].(string)
	if rawURL == "" {
		return nil, errors.New("url required")
	}
	if !strings.HasPrefix(rawURL, "https://") {
		return nil, errors.New("url must start with https://")
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", rawURL)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", rawURL)
	default:
		cmd = exec.Command("xdg-open", rawURL)
	}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("open browser: %w", err)
	}
	return map[string]any{"ok": true}, nil
}
