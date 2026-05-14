package analytics

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	"music-plugin/digitalitem"
)

type Enricher struct {
	db *sql.DB
}

func NewEnricher(db *sql.DB) *Enricher {
	return &Enricher{db: db}
}

// EnrichEvent fills missing title/artist/album/genre on the just-inserted
// play_events row. Best-effort: errors are logged and swallowed.
//
// Strategy:
//   1. Copy from playlist_items (matched by item id) where the play_events
//      column is currently empty. Cheap and reliable for items that were
//      added through the playlist flow.
//   2. If any field is still empty AND the playlist_items row has a local
//      file URL, read ID3 tags via digitalitem.ExtractMetadata and fill
//      remaining empties.
func (e *Enricher) EnrichEvent(eventID int64, itemID string) {
	if itemID == "" {
		return
	}

	if _, err := e.db.Exec(`
		UPDATE play_events
		SET title  = CASE WHEN title  = '' THEN COALESCE((SELECT title  FROM playlist_items WHERE id = ?), '') ELSE title  END,
		    artist = CASE WHEN artist = '' THEN COALESCE((SELECT artist FROM playlist_items WHERE id = ?), '') ELSE artist END,
		    album  = CASE WHEN album  = '' THEN COALESCE((SELECT album  FROM playlist_items WHERE id = ?), '') ELSE album  END,
		    genre  = CASE WHEN genre  = '' THEN COALESCE((SELECT genre  FROM playlist_items WHERE id = ?), '') ELSE genre  END
		WHERE id = ?
	`, itemID, itemID, itemID, itemID, eventID); err != nil {
		fmt.Fprintln(os.Stderr, "[analytics] enrich playlist_items copy:", err)
	}

	var title, artist, album, genre string
	if err := e.db.QueryRow(
		`SELECT title, artist, album, genre FROM play_events WHERE id = ?`, eventID,
	).Scan(&title, &artist, &album, &genre); err != nil {
		fmt.Fprintln(os.Stderr, "[analytics] enrich select:", err)
		return
	}
	if title != "" && artist != "" && album != "" && genre != "" {
		return
	}

	var url sql.NullString
	if err := e.db.QueryRow(
		`SELECT url FROM playlist_items WHERE id = ?`, itemID,
	).Scan(&url); err != nil {
		// Direct/in-memory plays have no playlist_items row; that's expected.
		return
	}
	if !url.Valid || !isLocalPath(url.String) {
		return
	}

	meta, err := digitalitem.ExtractMetadata(url.String)
	if err != nil {
		return
	}

	if _, err := e.db.Exec(`
		UPDATE play_events
		SET title  = CASE WHEN title  = '' THEN ? ELSE title  END,
		    artist = CASE WHEN artist = '' THEN ? ELSE artist END,
		    album  = CASE WHEN album  = '' THEN ? ELSE album  END,
		    genre  = CASE WHEN genre  = '' THEN ? ELSE genre  END
		WHERE id = ?
	`, meta.Title, meta.Artist, meta.Album, meta.Genre, eventID); err != nil {
		fmt.Fprintln(os.Stderr, "[analytics] enrich file-tag UPDATE:", err)
	}
}

func isLocalPath(url string) bool {
	if url == "" {
		return false
	}
	if strings.HasPrefix(url, "blob:") ||
		strings.HasPrefix(url, "http://") ||
		strings.HasPrefix(url, "https://") ||
		strings.HasPrefix(url, "data:") {
		return false
	}
	return true
}
