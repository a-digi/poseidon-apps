package playerstate

import (
	"database/sql"
	"errors"
	"fmt"
)

type PlayerStateRepository struct {
	DB *sql.DB
}

func NewPlayerStateRepository(db *sql.DB) (*PlayerStateRepository, error) {
	r := &PlayerStateRepository{DB: db}
	if err := r.migrate(); err != nil {
		return nil, fmt.Errorf("player_state migration: %w", err)
	}
	return r, nil
}

func (r *PlayerStateRepository) migrate() error {
	_, err := r.DB.Exec(`
		CREATE TABLE IF NOT EXISTS player_state (
			id                   INTEGER PRIMARY KEY CHECK (id = 1),
			selected_playlist_id TEXT NOT NULL DEFAULT '',
			play_mode            TEXT NOT NULL DEFAULT 'playlist',
			current_item_id      TEXT NOT NULL DEFAULT '',
			updated_at           INTEGER NOT NULL DEFAULT (strftime('%s','now'))
		)
	`)
	return err
}

func (r *PlayerStateRepository) Get() (PlayerState, error) {
	var s PlayerState
	err := r.DB.QueryRow(`
		SELECT selected_playlist_id, play_mode, current_item_id
		FROM player_state
		WHERE id = 1
	`).Scan(&s.SelectedPlaylistID, &s.PlayMode, &s.CurrentItemID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return PlayerState{}, nil
		}
		return PlayerState{}, err
	}
	if s.PlayMode != "playlist" && s.PlayMode != "repeat" && s.PlayMode != "shuffle" {
		s.PlayMode = "playlist"
	}
	return s, nil
}

func (r *PlayerStateRepository) Save(s PlayerState) (PlayerState, error) {
	_, err := r.DB.Exec(`
		INSERT OR REPLACE INTO player_state
			(id, selected_playlist_id, play_mode, current_item_id, updated_at)
		VALUES (1, ?, ?, ?, strftime('%s','now'))`,
		s.SelectedPlaylistID, s.PlayMode, s.CurrentItemID,
	)
	if err != nil {
		return PlayerState{}, err
	}
	return s, nil
}
