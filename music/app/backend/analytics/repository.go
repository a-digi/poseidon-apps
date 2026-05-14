package analytics

import (
	"database/sql"
	"fmt"
	"strconv"
)

type PlayEvent struct {
	PlayedAt   int64
	ItemID     string
	PlaylistID string
	Title      string
	Artist     string
	Album      string
	Genre      string
}

type Summary struct {
	TotalPlays            int    `json:"totalPlays"`
	UniqueItemsPlayed     int    `json:"uniqueItemsPlayed"`
	UniquePlaylistsPlayed int    `json:"uniquePlaylistsPlayed"`
	FirstPlayAt           *int64 `json:"firstPlayAt,omitempty"`
	LastPlayAt            *int64 `json:"lastPlayAt,omitempty"`
}

type TopItem struct {
	ItemID string `json:"itemId"`
	Title  string `json:"title"`
	Artist string `json:"artist"`
	Album  string `json:"album"`
	Plays  int    `json:"plays"`
}

type TopPlaylist struct {
	PlaylistID string `json:"playlistId"`
	Name       string `json:"name"`
	Plays      int    `json:"plays"`
}

type TopArtist struct {
	Artist string `json:"artist"`
	Plays  int    `json:"plays"`
}

type TopAlbum struct {
	Artist string `json:"artist"`
	Album  string `json:"album"`
	Plays  int    `json:"plays"`
}

type TopGenre struct {
	Genre string `json:"genre"`
	Plays int    `json:"plays"`
}

type Bucket struct {
	Key   string `json:"key"`
	Plays int    `json:"plays"`
}

type Overview struct {
	Summary      Summary       `json:"summary"`
	TopItems     []TopItem     `json:"topItems"`
	TopPlaylists []TopPlaylist `json:"topPlaylists"`
	TopArtists   []TopArtist   `json:"topArtists"`
	TopAlbums    []TopAlbum    `json:"topAlbums"`
	TopGenres    []TopGenre    `json:"topGenres"`
	ByHour       []Bucket      `json:"byHour"`
	ByWeekday    []Bucket      `json:"byWeekday"`
	ByMonth      []Bucket      `json:"byMonth"`
	ByYear       []Bucket      `json:"byYear"`
}

type AnalyticsRepository struct {
	DB *sql.DB
}

func NewAnalyticsRepository(db *sql.DB) (*AnalyticsRepository, error) {
	r := &AnalyticsRepository{DB: db}
	if err := r.migrate(); err != nil {
		return nil, fmt.Errorf("analytics migration: %w", err)
	}
	return r, nil
}

func (r *AnalyticsRepository) migrate() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS play_events (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			played_at   INTEGER NOT NULL,
			item_id     TEXT    NOT NULL,
			playlist_id TEXT,
			title       TEXT    NOT NULL DEFAULT '',
			artist      TEXT    NOT NULL DEFAULT '',
			album       TEXT    NOT NULL DEFAULT '',
			genre       TEXT    NOT NULL DEFAULT ''
		)`,
		`CREATE INDEX IF NOT EXISTS idx_play_events_item     ON play_events(item_id)`,
		`CREATE INDEX IF NOT EXISTS idx_play_events_playlist ON play_events(playlist_id)`,
		`CREATE INDEX IF NOT EXISTS idx_play_events_at       ON play_events(played_at)`,
	}
	for _, stmt := range stmts {
		if _, err := r.DB.Exec(stmt); err != nil {
			return err
		}
	}
	var hasGenre int
	if err := r.DB.QueryRow(
		`SELECT COUNT(*) FROM pragma_table_info('play_events') WHERE name = 'genre'`,
	).Scan(&hasGenre); err != nil {
		return err
	}
	if hasGenre == 0 {
		if _, err := r.DB.Exec(`ALTER TABLE play_events ADD COLUMN genre TEXT NOT NULL DEFAULT ''`); err != nil {
			return err
		}
	}
	return nil
}

func (r *AnalyticsRepository) Insert(ev PlayEvent) (int64, error) {
	var playlistID sql.NullString
	if ev.PlaylistID == "" {
		playlistID = sql.NullString{Valid: false}
	} else {
		playlistID = sql.NullString{String: ev.PlaylistID, Valid: true}
	}
	res, err := r.DB.Exec(
		`INSERT INTO play_events (played_at, item_id, playlist_id, title, artist, album, genre) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		ev.PlayedAt, ev.ItemID, playlistID, ev.Title, ev.Artist, ev.Album, ev.Genre,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *AnalyticsRepository) Overview() (Overview, error) {
	var ov Overview
	ov.TopItems = []TopItem{}
	ov.TopPlaylists = []TopPlaylist{}
	ov.TopArtists = []TopArtist{}
	ov.TopAlbums = []TopAlbum{}
	ov.TopGenres = []TopGenre{}
	ov.ByHour = []Bucket{}
	ov.ByWeekday = []Bucket{}
	ov.ByMonth = []Bucket{}
	ov.ByYear = []Bucket{}

	if err := r.DB.QueryRow(`SELECT COUNT(*) FROM play_events`).Scan(&ov.Summary.TotalPlays); err != nil {
		return Overview{}, err
	}
	if err := r.DB.QueryRow(`SELECT COUNT(DISTINCT item_id) FROM play_events`).Scan(&ov.Summary.UniqueItemsPlayed); err != nil {
		return Overview{}, err
	}
	if err := r.DB.QueryRow(`SELECT COUNT(DISTINCT playlist_id) FROM play_events WHERE playlist_id IS NOT NULL`).Scan(&ov.Summary.UniquePlaylistsPlayed); err != nil {
		return Overview{}, err
	}

	var minAt, maxAt sql.NullInt64
	if err := r.DB.QueryRow(`SELECT MIN(played_at), MAX(played_at) FROM play_events`).Scan(&minAt, &maxAt); err != nil {
		return Overview{}, err
	}
	if minAt.Valid {
		ov.Summary.FirstPlayAt = &minAt.Int64
	}
	if maxAt.Valid {
		ov.Summary.LastPlayAt = &maxAt.Int64
	}

	itemRows, err := r.DB.Query(`
		SELECT item_id, MAX(title), MAX(artist), MAX(album), COUNT(*) AS plays
		FROM play_events
		GROUP BY item_id
		ORDER BY plays DESC, MAX(played_at) DESC
		LIMIT 10
	`)
	if err != nil {
		return Overview{}, err
	}
	defer itemRows.Close()
	for itemRows.Next() {
		var t TopItem
		if err := itemRows.Scan(&t.ItemID, &t.Title, &t.Artist, &t.Album, &t.Plays); err != nil {
			return Overview{}, err
		}
		ov.TopItems = append(ov.TopItems, t)
	}
	if err := itemRows.Err(); err != nil {
		return Overview{}, err
	}

	plRows, err := r.DB.Query(`
		SELECT pe.playlist_id, COALESCE(p.name, '(deleted)') AS name, COUNT(*) AS plays
		FROM play_events pe
		LEFT JOIN playlists p ON p.id = pe.playlist_id
		WHERE pe.playlist_id IS NOT NULL
		GROUP BY pe.playlist_id
		ORDER BY plays DESC
		LIMIT 10
	`)
	if err != nil {
		return Overview{}, err
	}
	defer plRows.Close()
	for plRows.Next() {
		var tp TopPlaylist
		if err := plRows.Scan(&tp.PlaylistID, &tp.Name, &tp.Plays); err != nil {
			return Overview{}, err
		}
		ov.TopPlaylists = append(ov.TopPlaylists, tp)
	}
	if err := plRows.Err(); err != nil {
		return Overview{}, err
	}

	artistRows, err := r.DB.Query(`
		SELECT artist, COUNT(*) AS plays
		FROM play_events
		WHERE artist != ''
		GROUP BY artist
		ORDER BY plays DESC
		LIMIT 10
	`)
	if err != nil {
		return Overview{}, err
	}
	defer artistRows.Close()
	for artistRows.Next() {
		var t TopArtist
		if err := artistRows.Scan(&t.Artist, &t.Plays); err != nil {
			return Overview{}, err
		}
		ov.TopArtists = append(ov.TopArtists, t)
	}
	if err := artistRows.Err(); err != nil {
		return Overview{}, err
	}

	albumRows, err := r.DB.Query(`
		SELECT artist, album, COUNT(*) AS plays
		FROM play_events
		WHERE album != ''
		GROUP BY artist, album
		ORDER BY plays DESC
		LIMIT 10
	`)
	if err != nil {
		return Overview{}, err
	}
	defer albumRows.Close()
	for albumRows.Next() {
		var t TopAlbum
		if err := albumRows.Scan(&t.Artist, &t.Album, &t.Plays); err != nil {
			return Overview{}, err
		}
		ov.TopAlbums = append(ov.TopAlbums, t)
	}
	if err := albumRows.Err(); err != nil {
		return Overview{}, err
	}

	genreRows, err := r.DB.Query(`
		SELECT genre, COUNT(*) AS plays
		FROM play_events
		WHERE genre != ''
		GROUP BY genre
		ORDER BY plays DESC
		LIMIT 10
	`)
	if err != nil {
		return Overview{}, err
	}
	defer genreRows.Close()
	for genreRows.Next() {
		var t TopGenre
		if err := genreRows.Scan(&t.Genre, &t.Plays); err != nil {
			return Overview{}, err
		}
		ov.TopGenres = append(ov.TopGenres, t)
	}
	if err := genreRows.Err(); err != nil {
		return Overview{}, err
	}

	hourRows, err := r.DB.Query(`
		SELECT CAST(strftime('%H', played_at, 'unixepoch', 'localtime') AS INTEGER) AS hour, COUNT(*) AS plays
		FROM play_events
		GROUP BY hour
		ORDER BY hour
	`)
	if err != nil {
		return Overview{}, err
	}
	defer hourRows.Close()
	hourMap := make(map[int]int)
	for hourRows.Next() {
		var h, plays int
		if err := hourRows.Scan(&h, &plays); err != nil {
			return Overview{}, err
		}
		hourMap[h] = plays
	}
	if err := hourRows.Err(); err != nil {
		return Overview{}, err
	}
	for i := 0; i < 24; i++ {
		ov.ByHour = append(ov.ByHour, Bucket{Key: strconv.Itoa(i), Plays: hourMap[i]})
	}

	wdRows, err := r.DB.Query(`
		SELECT CAST(strftime('%w', played_at, 'unixepoch', 'localtime') AS INTEGER) AS weekday, COUNT(*) AS plays
		FROM play_events
		GROUP BY weekday
		ORDER BY weekday
	`)
	if err != nil {
		return Overview{}, err
	}
	defer wdRows.Close()
	wdMap := make(map[int]int)
	for wdRows.Next() {
		var wd, plays int
		if err := wdRows.Scan(&wd, &plays); err != nil {
			return Overview{}, err
		}
		wdMap[wd] = plays
	}
	if err := wdRows.Err(); err != nil {
		return Overview{}, err
	}
	for i := 0; i < 7; i++ {
		ov.ByWeekday = append(ov.ByWeekday, Bucket{Key: strconv.Itoa(i), Plays: wdMap[i]})
	}

	monthRows, err := r.DB.Query(`
		SELECT strftime('%Y-%m', played_at, 'unixepoch', 'localtime') AS month, COUNT(*) AS plays
		FROM play_events
		GROUP BY month
		ORDER BY month
	`)
	if err != nil {
		return Overview{}, err
	}
	defer monthRows.Close()
	for monthRows.Next() {
		var b Bucket
		if err := monthRows.Scan(&b.Key, &b.Plays); err != nil {
			return Overview{}, err
		}
		ov.ByMonth = append(ov.ByMonth, b)
	}
	if err := monthRows.Err(); err != nil {
		return Overview{}, err
	}

	yearRows, err := r.DB.Query(`
		SELECT strftime('%Y', played_at, 'unixepoch', 'localtime') AS year, COUNT(*) AS plays
		FROM play_events
		GROUP BY year
		ORDER BY year
	`)
	if err != nil {
		return Overview{}, err
	}
	defer yearRows.Close()
	for yearRows.Next() {
		var b Bucket
		if err := yearRows.Scan(&b.Key, &b.Plays); err != nil {
			return Overview{}, err
		}
		ov.ByYear = append(ov.ByYear, b)
	}
	if err := yearRows.Err(); err != nil {
		return Overview{}, err
	}

	return ov, nil
}
