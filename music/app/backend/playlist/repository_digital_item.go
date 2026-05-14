package playlist

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type DigitalItemRepository struct {
	DB *sql.DB
}

func NewDigitalItemRepository(db *sql.DB) (*DigitalItemRepository, error) {
	r := &DigitalItemRepository{DB: db}
	if err := r.migrate(); err != nil {
		return nil, fmt.Errorf("playlist_item migration: %w", err)
	}
	return r, nil
}

func (r *DigitalItemRepository) migrate() error {
	_, err := r.DB.Exec(`
		CREATE TABLE IF NOT EXISTS playlist_items (
			id          TEXT PRIMARY KEY,
			playlist_id TEXT NOT NULL,
			title       TEXT NOT NULL DEFAULT '',
			url         TEXT NOT NULL DEFAULT '',
			artist      TEXT NOT NULL DEFAULT '',
			album       TEXT NOT NULL DEFAULT '',
			genre       TEXT NOT NULL DEFAULT '',
			year        INTEGER NOT NULL DEFAULT 0,
			track       INTEGER NOT NULL DEFAULT 0,
			length      INTEGER NOT NULL DEFAULT 0,
			picture     TEXT NOT NULL DEFAULT '',
			mime_type   TEXT NOT NULL DEFAULT '',
			FOREIGN KEY (playlist_id) REFERENCES playlists(id)
		)
	`)
	return err
}

func (r *DigitalItemRepository) FindByPlaylistID(playlistID string) ([]DigitalItem, error) {
	rows, err := r.DB.Query(`
		SELECT id, playlist_id, title, url, artist, album, genre, year, track, length, picture, mime_type
		FROM playlist_items
		WHERE playlist_id = ?`, playlistID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []DigitalItem{}
	for rows.Next() {
		var item DigitalItem
		if err := rows.Scan(
			&item.ID, &item.PlaylistId, &item.Title, &item.URL,
			&item.Artist, &item.Album, &item.Genre,
			&item.Year, &item.Track, &item.Length,
			&item.Picture, &item.MimeType,
		); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *DigitalItemRepository) Insert(item DigitalItem) (DigitalItem, error) {
	item.ID = uuid.NewString()
	_, err := r.DB.Exec(`
		INSERT INTO playlist_items
			(id, playlist_id, title, url, artist, album, genre, year, track, length, picture, mime_type)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		item.ID, item.PlaylistId, item.Title, item.URL,
		item.Artist, item.Album, item.Genre,
		item.Year, item.Track, item.Length,
		item.Picture, item.MimeType,
	)
	if err != nil {
		return DigitalItem{}, err
	}
	return item, nil
}

func (r *DigitalItemRepository) Delete(id string) error {
	res, err := r.DB.Exec(`DELETE FROM playlist_items WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errors.New("playlist item not found")
	}
	return nil
}
