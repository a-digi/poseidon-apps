package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"

	"mhype-plugin/crawler"
	"mhype-plugin/queueclient"
	"mhype-plugin/storage"
)

const (
	heartbeatFileName   = "orchestrator.heartbeat"
	triggersDirName     = "triggers"
	keepPerCrawler      = 100
	perCrawlTimeout     = 30 * time.Second
	workerPoolSize      = 2
	heartbeatInterval   = 30 * time.Second
	triggerPollInterval = 1 * time.Second
	shutdownGracePeriod = 5 * time.Second
	httpClientTimeout   = 25 * time.Second

	youTubeQueueEnabled = false
	youTubeQueueName    = "mhype_chart-youtube-video"
)

type crawlJob struct {
	id      string
	crawler crawler.Crawler
}

func Run(dataDir string, registered []crawler.Crawler) error {
	store, err := storage.NewStore(dataDir)
	if err != nil {
		return fmt.Errorf("init store: %w", err)
	}
	state, err := NewStateStore(dataDir)
	if err != nil {
		return fmt.Errorf("init orchestrator state: %w", err)
	}

	httpClient := &http.Client{Timeout: httpClientTimeout}

	suggestionsStore, err := storage.NewSuggestionsStore(dataDir)
	if err != nil {
		return fmt.Errorf("init suggestions store: %w", err)
	}

	parentCtx, cancel := context.WithCancel(context.Background())

	var queueCli *queueclient.Client
	if qc, qcErr := queueclient.New(httpClient); qcErr != nil {
		log.Printf("[orchestrator] queue client unavailable: %v (publishing disabled)", qcErr)
	} else if youTubeQueueEnabled {
		queueCli = qc
		regCtx, regCancel := context.WithTimeout(parentCtx, 10*time.Second)
		if regErr := queueCli.Register(regCtx,
			youTubeQueueName,
			"Resolve YouTube video suggestions for crawled chart entries",
			queueclient.Consumer{
				Action:        "queue_chart_youtube",
				Workers:       4,
				MaxAttempts:   3,
				BackoffMillis: []int{2_000, 30_000, 300_000},
				BufferSize:    64,
			}); regErr != nil {
			log.Printf("[orchestrator] register queue: %v (publishing disabled until next start)", regErr)
			queueCli = nil
		}
		regCancel()
	} else {
		deregCtx, deregCancel := context.WithTimeout(parentCtx, 10*time.Second)
		if derr := qc.Deregister(deregCtx, youTubeQueueName); derr != nil {
			log.Printf("[orchestrator] deregister queue %q: %v", youTubeQueueName, derr)
		} else {
			log.Printf("[orchestrator] queue %q detached (youTubeQueueEnabled=false)", youTubeQueueName)
		}
		deregCancel()
	}

	defer cancel()

	log.Printf("[orchestrator] starting (pid=%d, dataDir=%s)", os.Getpid(), dataDir)

	jobs := make(chan crawlJob, workerPoolSize)
	var workerWG sync.WaitGroup
	for i := 0; i < workerPoolSize; i++ {
		workerWG.Add(1)
		go runWorker(parentCtx, &workerWG, jobs, store, httpClient, state, suggestionsStore, queueCli)
	}

	var schedWG sync.WaitGroup
	schedWG.Add(1)
	go runHeartbeat(parentCtx, &schedWG, dataDir)

	byID := make(map[string]crawler.Crawler, len(registered))
	for _, c := range registered {
		byID[c.ID()] = c
		schedWG.Add(1)
		go runScheduler(parentCtx, &schedWG, c, jobs, state)
	}

	schedWG.Add(1)
	go runTriggerWatcher(parentCtx, &schedWG, dataDir, byID, jobs)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	defer signal.Stop(sigCh)

	select {
	case <-sigCh:
		log.Printf("[orchestrator] shutting down (signal received), waiting for workers...")
	case <-parentCtx.Done():
	}

	cancel()
	schedWG.Wait()
	close(jobs)

	done := make(chan struct{})
	go func() {
		workerWG.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(shutdownGracePeriod):
		log.Printf("[orchestrator] grace period expired, some crawls may be in flight")
	}

	log.Printf("[orchestrator] exited cleanly")
	return nil
}

func runHeartbeat(ctx context.Context, wg *sync.WaitGroup, dataDir string) {
	defer wg.Done()
	path := filepath.Join(dataDir, heartbeatFileName)
	writeHeartbeat(path)
	t := time.NewTicker(heartbeatInterval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			writeHeartbeat(path)
		}
	}
}

func writeHeartbeat(path string) {
	payload := []byte(strconv.FormatInt(time.Now().UnixMilli(), 10))
	if err := os.WriteFile(path, payload, 0o644); err != nil {
		log.Printf("[orchestrator] heartbeat write failed: %v", err)
	}
}

func runScheduler(ctx context.Context, wg *sync.WaitGroup, c crawler.Crawler, jobs chan<- crawlJob, state *StateStore) {
	defer wg.Done()
	const maxInterval = 12 * time.Hour
	interval := c.Interval()
	if interval > maxInterval {
		interval = maxInterval
	}
	rs := state.Get(c.ID())
	var initialDelay time.Duration
	if rs.LastSuccessAt > 0 {
		elapsed := time.Duration(time.Now().UnixMilli()-rs.LastSuccessAt) * time.Millisecond
		if elapsed < interval {
			initialDelay = interval - elapsed
		}
	}
	log.Printf("[orchestrator] crawler %s scheduled (interval=%s, initialDelay=%s)", c.ID(), interval, initialDelay)
	select {
	case <-ctx.Done():
		return
	case <-time.After(initialDelay):
	}
	submitJob(jobs, crawlJob{id: c.ID(), crawler: c})
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(interval):
			submitJob(jobs, crawlJob{id: c.ID(), crawler: c})
		}
	}
}

func runTriggerWatcher(ctx context.Context, wg *sync.WaitGroup, dataDir string, byID map[string]crawler.Crawler, jobs chan<- crawlJob) {
	defer wg.Done()
	dir := filepath.Join(dataDir, triggersDirName)
	t := time.NewTicker(triggerPollInterval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			scanTriggers(ctx, dir, byID, jobs)
		}
	}
}

func scanTriggers(ctx context.Context, dir string, byID map[string]crawler.Crawler, jobs chan<- crawlJob) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.Printf("[orchestrator] read triggers dir: %v", err)
		}
		return
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		c, ok := byID[name]
		if !ok {
			continue
		}
		if err := os.Remove(filepath.Join(dir, name)); err != nil && !errors.Is(err, os.ErrNotExist) {
			log.Printf("[orchestrator] remove trigger %s: %v", name, err)
			continue
		}
		log.Printf("[orchestrator] trigger received for %s", name)
		submitJobBlocking(ctx, jobs, crawlJob{id: c.ID(), crawler: c})
	}
}

func submitJob(jobs chan<- crawlJob, j crawlJob) {
	select {
	case jobs <- j:
	default:
		log.Printf("[orchestrator] worker pool full, dropping run for %s", j.id)
	}
}

func submitJobBlocking(ctx context.Context, jobs chan<- crawlJob, j crawlJob) {
	select {
	case <-ctx.Done():
	case jobs <- j:
	}
}

func runWorker(ctx context.Context, wg *sync.WaitGroup, jobs <-chan crawlJob, store *storage.Store, httpClient *http.Client, state *StateStore, ss *storage.SuggestionsStore, qc *queueclient.Client) {
	defer wg.Done()
	for j := range jobs {
		executeCrawl(ctx, j, store, httpClient, state, ss, qc)
	}
}

func executeCrawl(parent context.Context, j crawlJob, store *storage.Store, httpClient *http.Client, state *StateStore, ss *storage.SuggestionsStore, qc *queueclient.Client) {
	ctx, cancel := context.WithTimeout(parent, perCrawlTimeout)
	defer cancel()
	if err := state.RecordAttempt(j.id, time.Now().UnixMilli()); err != nil {
		log.Printf("[orchestrator] state record attempt %s: %v", j.id, err)
	}
	result, err := j.crawler.Crawl(ctx, httpClient)
	if err == nil {
		if serr := state.RecordSuccess(j.id, time.Now().UnixMilli()); serr != nil {
			log.Printf("[orchestrator] state record success %s: %v", j.id, serr)
		}
	}
	if err != nil {
		log.Printf("[orchestrator] crawler %s failed: %v", j.id, err)
		if rerr := state.RecordFailure(j.id, time.Now().UnixMilli(), err.Error()); rerr != nil {
			log.Printf("[orchestrator] state record failure %s: %v", j.id, rerr)
		}
		return
	}
	if result.Snapshot && result.BundleKey != "" {
		valid := make([]storage.Item, 0, len(result.Items))
		for _, ci := range result.Items {
			if err := ci.Validate(); err != nil {
				log.Printf("[orchestrator] crawler %s: invalid item %q dropped: %v", j.id, ci.ID, err)
				continue
			}
			valid = append(valid, toStorageItem(ci))
		}
		if werr := store.WriteBundle(j.id, result.BundleKey, valid); werr != nil {
			log.Printf("[orchestrator] crawler %s snapshot %s write: %v", j.id, result.BundleKey, werr)
		} else {
			log.Printf("[orchestrator] crawler %s: snapshot %s (%d items)", j.id, result.BundleKey, len(valid))
		}
	} else if result.BundleKey == "" {
		for _, ci := range result.Items {
			if err := ci.Validate(); err != nil {
				log.Printf("[orchestrator] crawler %s: invalid item %q dropped: %v", j.id, ci.ID, err)
				continue
			}
			if err := store.Write(j.id, toStorageItem(ci)); err != nil {
				log.Printf("[orchestrator] crawler %s write item %s: %v", j.id, ci.ID, err)
			}
		}
		log.Printf("[orchestrator] crawler %s: %d items", j.id, len(result.Items))
	} else {
		valid := make([]storage.Item, 0, len(result.Items))
		for _, ci := range result.Items {
			if err := ci.Validate(); err != nil {
				log.Printf("[orchestrator] crawler %s: invalid item %q dropped: %v", j.id, ci.ID, err)
				continue
			}
			valid = append(valid, toStorageItem(ci))
		}
		created, werr := store.WriteBundleIfMissing(j.id, result.BundleKey, valid)
		if werr != nil {
			log.Printf("[orchestrator] crawler %s bundle %s write: %v", j.id, result.BundleKey, werr)
		} else {
			log.Printf("[orchestrator] crawler %s: bundle %s (%d items) created=%v", j.id, result.BundleKey, len(valid), created)
		}
	}
	removed, err := store.EvictOldest(j.id, keepPerCrawler)
	if err != nil {
		log.Printf("[orchestrator] crawler %s evict: %v", j.id, err)
	}
	if removed > 0 {
		log.Printf("[orchestrator] crawler %s: evicted %d old files", j.id, removed)
	}
	publishYouTubeTasks(parent, qc, ss, result.Items)
}

func toStorageItem(c crawler.Item) storage.Item {
	return storage.Item{
		ID:        c.ID,
		Kind:      storage.ItemKind(c.Kind),
		CrawlerID: c.CrawlerID,
		SourceURL: c.SourceURL,
		Title:     c.Title,
		ScrapedAt: c.ScrapedAt,

		Artists:     c.Artists,
		Album:       c.Album,
		Label:       c.Label,
		Genres:      c.Genres,
		DurationSec: c.DurationSec,
		ReleasedAt:  c.ReleasedAt,

		SourceID:    c.SourceID,
		ExternalIDs: c.ExternalIDs,

		ArtworkURL: c.ArtworkURL,
		PreviewURL: c.PreviewURL,

		Chart: toStorageChart(c.Chart),
		News:  toStorageNews(c.News),

		Extra: c.Extra,
	}
}

func toStorageChart(c *crawler.ChartContext) *storage.ChartContext {
	if c == nil {
		return nil
	}
	return &storage.ChartContext{
		Name:         c.Name,
		Position:     c.Position,
		PrevPosition: c.PrevPosition,
		PeakPosition: c.PeakPosition,
		WeeksOnChart: c.WeeksOnChart,
		PeriodStart:  c.PeriodStart,
		PeriodEnd:    c.PeriodEnd,
	}
}

func toStorageNews(n *crawler.NewsContext) *storage.NewsContext {
	if n == nil {
		return nil
	}
	return &storage.NewsContext{
		Summary:        n.Summary,
		PublishedAt:    n.PublishedAt,
		RelatedArtists: n.RelatedArtists,
	}
}

func publishYouTubeTasks(parent context.Context, qc *queueclient.Client, ss *storage.SuggestionsStore, items []crawler.Item) {
	if qc == nil {
		return
	}
	for _, ci := range items {
		if string(ci.Kind) != "chart-entry" {
			continue
		}
		if len(ci.Artists) == 0 || ci.Title == "" {
			continue
		}
		artist := ci.Artists[0]
		if ss.Has(artist, ci.Title) {
			continue
		}
		ctx, cancel := context.WithTimeout(parent, 5*time.Second)
		if _, err := qc.Publish(ctx, "mhype_chart-youtube-video", map[string]string{
			"artist": artist,
			"title":  ci.Title,
		}); err != nil {
			log.Printf("[orchestrator] publish %q/%q: %v", artist, ci.Title, err)
		}
		cancel()
	}
}
