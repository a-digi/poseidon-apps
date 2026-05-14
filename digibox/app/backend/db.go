package main

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

func openDB() (*sql.DB, error) {
	dir := os.Getenv("PLUGIN_DATA_DIR")
	if dir == "" {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, "Library", "Application Support", "poseidon", "plugins", "digibox", "data")
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	return sql.Open("sqlite3", filepath.Join(dir, "digibox.db"))
}
