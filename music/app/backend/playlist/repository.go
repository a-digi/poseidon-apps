package playlist

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type PlaylistRepository struct {
	DB *sql.DB
}

func NewPlaylistRepository(db *sql.DB) (*PlaylistRepository, error) {
	r := &PlaylistRepository{DB: db}
	if err := r.migrate(); err != nil {
		return nil, fmt.Errorf("playlist migration: %w", err)
	}
	return r, nil
}

func (r *PlaylistRepository) migrate() error {
	_, err := r.DB.Exec(`
		CREATE TABLE IF NOT EXISTS playlists (
			id         TEXT PRIMARY KEY,
			name       TEXT NOT NULL,
			created_at INTEGER NOT NULL DEFAULT (strftime('%s','now'))
		)
	`)
	return err
}

func (r *PlaylistRepository) FindAll() ([]PlaylistIndex, error) {
	rows, err := r.DB.Query(`SELECT id, name FROM playlists ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []PlaylistIndex{}
	for rows.Next() {
		var p PlaylistIndex
		if err := rows.Scan(&p.ID, &p.Name); err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, rows.Err()
}

func (r *PlaylistRepository) Insert(name string) (PlaylistIndex, error) {
	id := uuid.NewString()
	_, err := r.DB.Exec(`INSERT INTO playlists (id, name) VALUES (?, ?)`, id, name)
	if err != nil {
		return PlaylistIndex{}, err
	}
	return PlaylistIndex{ID: id, Name: name}, nil
}

func (r *PlaylistRepository) Update(id, name string) (PlaylistIndex, error) {
	res, err := r.DB.Exec(`UPDATE playlists SET name = ? WHERE id = ?`, name, id)
	if err != nil {
		return PlaylistIndex{}, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return PlaylistIndex{}, errors.New("playlist not found")
	}
	return PlaylistIndex{ID: id, Name: name}, nil
}

func (r *PlaylistRepository) Delete(id string) error {
	res, err := r.DB.Exec(`DELETE FROM playlists WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errors.New("playlist not found")
	}
	return nil
}
