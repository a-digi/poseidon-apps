package playlist

import (
	"database/sql"
	"fmt"
	"sync"

	"music-plugin/digitalitem"
	"music-plugin/response"
)

type DigitalItem struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	URL        string `json:"url"`
	PlaylistId string `json:"playlistId"`
	Artist     string `json:"artist"`
	Album      string `json:"album"`
	Genre      string `json:"genre"`
	Year       int    `json:"year"`
	Track      int    `json:"track"`
	Length     int    `json:"length"`
	Picture    string `json:"picture"`
	MimeType   string `json:"mimeType"`
}

type Playlist struct {
	ID    string        `json:"id"`
	Items []DigitalItem `json:"items"`
}

type PlaylistIndex struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type PlaylistService struct {
	mu                    sync.Mutex
	repository            *PlaylistRepository
	digitalItemRepository *DigitalItemRepository
}

func NewPlaylistService(db *sql.DB) (*PlaylistService, error) {
	repo, err := NewPlaylistRepository(db)
	if err != nil {
		return nil, fmt.Errorf("NewPlaylistRepository: %w", err)
	}

	itemRepo, err := NewDigitalItemRepository(db)
	if err != nil {
		return nil, fmt.Errorf("NewDigitalItemRepository: %w", err)
	}

	return &PlaylistService{
		repository:            repo,
		digitalItemRepository: itemRepo,
	}, nil
}

func (ps *PlaylistService) Create(name string) string {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	pl, err := ps.repository.Insert(name)
	if err != nil {
		return response.ErrorResponse(fmt.Sprintf("Fehler beim Erstellen der Playlist: %v", err))
	}
	return response.SuccessResponse(pl)
}

func (ps *PlaylistService) Edit(id, name string) string {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	pl, err := ps.repository.Update(id, name)
	if err != nil {
		return response.ErrorResponse(fmt.Sprintf("Fehler beim Aktualisieren der Playlist: %v", err))
	}
	return response.SuccessResponse(pl)
}

func (ps *PlaylistService) Delete(id string) string {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if err := ps.repository.Delete(id); err != nil {
		return response.ErrorResponse(fmt.Sprintf("Fehler beim Löschen der Playlist: %v", err))
	}
	return response.SuccessResponse(PlaylistIndex{ID: id})
}

func (ps *PlaylistService) AddItem(playlistID string, item DigitalItem) string {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	item.PlaylistId = playlistID

	if item.URL != "" {
		if meta, err := digitalitem.ExtractMetadata(item.URL); err == nil {
			if meta.Title != "" {
				item.Title = meta.Title
			}
			item.Artist = meta.Artist
			item.Album = meta.Album
			item.Genre = meta.Genre
			item.Year = meta.Year
			item.Track = meta.Track
			item.Length = meta.Length
			item.Picture = meta.Picture
			item.MimeType = meta.MimeType
		}
	}

	inserted, err := ps.digitalItemRepository.Insert(item)
	if err != nil {
		return response.ErrorResponse(fmt.Sprintf("Fehler beim Hinzufügen des Items: %v", err))
	}
	return response.SuccessResponse(inserted)
}

func (ps *PlaylistService) GetByID(id string) string {
	items, err := ps.digitalItemRepository.FindByPlaylistID(id)
	if err != nil {
		return response.ErrorResponse(fmt.Sprintf("Fehler beim Laden der Playlist-Items: %v", err))
	}
	return response.SuccessResponse(Playlist{ID: id, Items: items})
}

func (ps *PlaylistService) ListPlaylists() string {
	entries, err := ps.repository.FindAll()
	if err != nil {
		return response.SuccessResponse([]PlaylistIndex{})
	}
	return response.SuccessResponse(entries)
}

func (ps *PlaylistService) DeleteItem(playlistID string, itemID string) string {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if err := ps.digitalItemRepository.Delete(itemID); err != nil {
		return response.ErrorResponse(fmt.Sprintf("Fehler beim Löschen des Items: %v", err))
	}
	return response.SuccessResponse(DigitalItem{ID: itemID, PlaylistId: playlistID})
}
