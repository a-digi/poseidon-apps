package analytics

import (
	"database/sql"
	"sync"
	"time"

	"music-plugin/response"
)

type AnalyticsService struct {
	mu       sync.Mutex
	repo     *AnalyticsRepository
	enricher *Enricher
}

func NewAnalyticsService(db *sql.DB) (*AnalyticsService, error) {
	repo, err := NewAnalyticsRepository(db)
	if err != nil {
		return nil, err
	}
	return &AnalyticsService{repo: repo, enricher: NewEnricher(db)}, nil
}

func (s *AnalyticsService) Record(itemID, playlistID, title, artist, album string) string {
	if itemID == "" {
		return response.ErrorResponse("itemId required")
	}

	s.mu.Lock()
	ev := PlayEvent{
		PlayedAt:   time.Now().Unix(),
		ItemID:     itemID,
		PlaylistID: playlistID,
		Title:      title,
		Artist:     artist,
		Album:      album,
	}
	eventID, err := s.repo.Insert(ev)
	s.mu.Unlock()
	if err != nil {
		return response.ErrorResponse(err.Error())
	}

	// Plugin runs as a one-shot CLI binary; main exits when this function
	// returns. Wait so the enrichment commits before the process dies.
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.enricher.EnrichEvent(eventID, itemID)
	}()
	wg.Wait()

	return response.SuccessResponse(map[string]bool{"ok": true})
}

func (s *AnalyticsService) Overview() string {
	ov, err := s.repo.Overview()
	if err != nil {
		return response.ErrorResponse(err.Error())
	}
	return response.SuccessResponse(ov)
}
